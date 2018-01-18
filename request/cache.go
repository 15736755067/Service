package request

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"Service/common"
	//"Service/config"
	"Service/db"
	"Service/log"
	"time"
)

const LocalCacheSvrTitle = "LOCALREQCACHE"
const remoteCacheSvrTitle = "REMOTEREQCACHE"
const defaultAsyncBuffer = 100000

var remoteCacheAsync bool // 是否异步保存
var remoteSetStmt *sql.Stmt

var mu sync.RWMutex // protects the following
var reqCache = make(map[string]string)

func InitRemoteCacheStmt(async bool, buffer int) {
	svr := db.GetDB(remoteCacheSvrTitle)
	if svr == nil {
		panic(fmt.Sprintf("[InitRemoteCache]%s remote cache DB does not exist", remoteCacheSvrTitle))
	}
	var err error
	remoteSetStmt, err = svr.Prepare("INSERT INTO Cache.ReqCache(clickId,value) VALUES(?,?) ON DUPLICATE KEY UPDATE value=?")
	if err != nil {
		panic(err.Error())
	}
	remoteCacheAsync = async
	if buffer <= 0 {
		buffer = defaultAsyncBuffer
	}
	toSave = make(chan *cacheValue, buffer)
}

func CloseRemoteCacheStmt() {
	if remoteSetStmt != nil {
		remoteSetStmt.Close()
	}
}

func get(token string) (caStr string) {
	if token == "" {
		return ""
	}
	mu.RLock()
	defer mu.RUnlock()
	caStr, _ = reqCache[token]
	return
}
func del(token string) {
	if token == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	delete(reqCache, token)
}
func set(token, caStr string) error {
	if token == "" {
		return errors.New("token is empty")
	}
	mu.Lock()
	defer mu.Unlock()
	reqCache[token] = caStr
	return nil
}

type cacheValue struct {
	key   string
	value string
}

var toSave chan *cacheValue

func AsyncingRemoteCache(stop chan struct{}) {
	for {
		select {
		case m := <-toSave:
			if m == nil {
				log.Error("saveRemoteCache failed:m is nil")
				continue
			}
			err := saveRemoteCache(false, m.key, m.value)
			if err != nil {
				log.Errorf("saveRemoteCache failed:%v", err)
			}
		case <-stop:
			// 收所有的数据，防止的未写入数据库的
			for {
				select {
				case m := <-toSave:
					if m == nil {
						log.Error("saveRemoteCache failed:m is nil")
						continue
					}
					err := saveRemoteCache(false, m.key, m.value)
					if err != nil {
						log.Errorf("saveRemoteCache failed:%v", err)
					}
				default:
					goto allreceived
				}
			}
		allreceived:
			return
		}
	}
}

func saveRemoteCache(async bool, reqId, value string) (err error) {
	if !async { // 同步保存RemoteCache
		//TODO 远程的cache暂时用MySQL来做，没有用上expire;上线一个月之后必须修改
		_, err = remoteSetStmt.Exec(reqId, value, value)
	} else {
		// save remote async
		select {
		case toSave <- &cacheValue{
			key:   reqId,
			value: value,
		}:
		default:
			log.Errorf("[saveRemoteCache]toSave is full, cacheValue lost,%s:%s\n", reqId, value)
		}
	}
	return
}

func setReqCache(req *reqbase, localExpire, remoteExpire time.Duration) (err error) {
	if req == nil {
		return errors.New("req is nil for setReqCache")
	}
	if req.Id() == "" {
		return errors.New("req.Id() is empty for setReqCache")
	}

	strategy := 1
	switch strategy {
	case 1: //方案1：使用中心Cache服务器。
		v := Req2cacheStr(req)
		log.Infof("[request][setReqCache]key:%s value:%s\n", req.Id(), v)
		{ // local cache
			if localExpire >= 0 {
				svr := db.GetRedisClient(LocalCacheSvrTitle)
				if svr == nil {
					return fmt.Errorf("[setReqCache]%s local cache DB does not exist", LocalCacheSvrTitle)
				}
				err = svr.Set(req.Id(), v, localExpire).Err()
				log.Infof("[request][setReqCache]local key:%s err:%v\n", req.Id(), err)
			}
		}
		{ // remote cache
			if remoteExpire >= 0 {
				err = saveRemoteCache(remoteCacheAsync, req.Id(), v)
				log.Infof("[request][setReqCache]remote key:%s err:%v\n", req.Id(), err)
			}
		}
	case 2: //方案2：使用程序内部的Cache。外部通过userIdText做分流。
		{
			err = set(req.Id(), Req2cacheStr(req))
		}
	}
	return
}

func getClickValue(dataSourceName, reqId string) (string, error) {
	var value string
	var err error

	svr := db.GetDB(dataSourceName)
	if svr == nil {
		err = fmt.Errorf("[getClickValue]%s remote cache DB does not exist", dataSourceName)
		return "", err
	}
	err = svr.QueryRow("SELECT value FROM Cache.ReqCache WHERE clickId=?", reqId).Scan(&value)
	if err == nil {
		return value, nil
	} else {
		err = fmt.Errorf("[getClickValue]remote %s get %v failed:%v", dataSourceName, reqId, err)
		return "", err
	}
	return "", nil
}

func getReqCache(reqId string, onlyLocal bool) (req *reqbase, err error) {
	strategy := 1
	switch strategy {
	case 1: //方案1：使用中心Cache服务器。
		svr := db.GetRedisClient(LocalCacheSvrTitle)
		if svr == nil {
			return nil, fmt.Errorf("[getReqCache]%s local cache DB does not exist", LocalCacheSvrTitle)
		}
		cmd := svr.Get(reqId)

		if cmd.Err() != nil { // 在Local没有找到时
			err = fmt.Errorf("[getReqCache]local %s get %v failed:%v", LocalCacheSvrTitle, reqId, cmd.Err())
			log.Error(err.Error())
			// 不能直接return，还有可能要尝试Remote部分
		} else if cmd.Val() != "" {
			req = CacheStr2Req(cmd.Val())
		}

		if req == nil && !onlyLocal { // 当local cache没有找到时，尝试从remote cache查找
			//TODO 在线上所有的clickId，都转化为aes clickId之前，先屏蔽这个检查 2017/3/21
			/*if !common.IndateClickId(reqId, config.ClickCacheTime) {
				// 检查时间，如果不在一个月内，则省去查询这一步
				return nil, fmt.Errorf("[getReqCache]%s exceeds one month, omit searching from %s", reqId, remoteCacheSvrTitle)
			}*/

			value, err := getClickValue(remoteCacheSvrTitle, reqId)
			if err == nil && value != "" {
				req = CacheStr2Req(value)
				return req, nil
			}

			return nil, err
		}
	case 2: //方案2：使用程序内部的Cache。外部通过userIdText做分流。
		{
			req = CacheStr2Req(get(reqId))
		}
	}
	return
}

func delReqCache(token string, local bool) {
	strategy := 1
	switch strategy {
	case 1: //方案1：使用中心Cache服务器。
		if local {
			svr := db.GetRedisClient(LocalCacheSvrTitle)
			if svr == nil {
				log.Errorf("[delReqCache]%s local cache DB does not exist\n", LocalCacheSvrTitle)
			}
			if err := svr.Del(token).Err(); err != nil {
				log.Errorf("[delReqCache]local delReqCache token(%s) with err(%s)\n", token, err.Error())
			}
		} else {
			svr := db.GetDB(remoteCacheSvrTitle)
			if svr == nil {
				log.Errorf("[delReqCache]%s remote cache DB does not exist\n", remoteCacheSvrTitle)
			}
			_, err := svr.Exec("DELETE FROM Cache.ReqCache WHERE clickId=?", token)
			if err != nil {
				log.Errorf("[delReqCache]remote delReqCache token(%s) with err(%s)\n", token, err.Error())
			}
		}
	case 2: //方案2：使用程序内部的Cache。外部通过userIdText做分流。
		{
			del(token)
		}
	case 3: //方案3：有效内容包装在token中。内容太多token太长。开发简便。
		{
			// do nothing
		}
	}
	return
}

func Req2cacheStr(req *reqbase) (caStr string) {
	if req == nil {
		return ""
	}
	ku, _ := url.ParseQuery("")
	ku.Add("id", req.id)
	ku.Add("t", req.t)
	ku.Add("ip", req.ip)
	ku.Add("ua", req.ua)

	ku.Add("externalId", req.externalId)
	ku.Add("cost", fmt.Sprintf("%f", req.cost))
	ku.Add("tsCId", req.tsCId)
	ku.Add("websiteId", req.websiteId)
	ku.Add("vars", strings.Join(req.vars, ";"))
	ku.Add("payout", fmt.Sprintf("%f", req.payout))
	ku.Add("txId", req.txid)

	ku.Add("tsId", fmt.Sprintf("%d", req.trafficSourceId))
	ku.Add("tsName", req.trafficSourceName)
	ku.Add("uId", fmt.Sprintf("%d", req.userId))
	ku.Add("uIdText", req.userIdText)
	ku.Add("cHash", req.campaignHash)
	ku.Add("cId", fmt.Sprintf("%d", req.campaignId))
	ku.Add("cName", req.campaignName)
	ku.Add("cCountry", req.campaignCountry)
	ku.Add("fId", fmt.Sprintf("%d", req.flowId))
	ku.Add("flowName", req.flowName)
	ku.Add("rId", fmt.Sprintf("%d", req.ruleId))
	ku.Add("pId", fmt.Sprintf("%d", req.pathId))
	ku.Add("lId", fmt.Sprintf("%d", req.landerId))
	ku.Add("lName", req.landerName)
	ku.Add("oId", fmt.Sprintf("%d", req.offerId))
	ku.Add("oOId", fmt.Sprintf("%d", req.optOfferId))
	ku.Add("oName", req.offerName)
	ku.Add("affId", fmt.Sprintf("%d", req.affiliateId))
	ku.Add("oAffId", fmt.Sprintf("%d", req.optAffiliateId))
	ku.Add("affName", req.affiliateName)

	ku.Add("impTs", fmt.Sprintf("%d", req.impTimeStamp))
	ku.Add("visitTs", fmt.Sprintf("%d", req.visitTimeStamp))
	ku.Add("clickTs", fmt.Sprintf("%d", req.clickTimeStamp))
	ku.Add("pbTs", fmt.Sprintf("%d", req.postbackTimeStamp))

	ku.Add("dType", req.deviceType)
	ku.Add("trkDomain", req.trackingDomain)
	ku.Add("trkPath", req.trackingPath)
	ku.Add("ref", req.referrer)
	ku.Add("refDomain", req.referrerdomain)
	ku.Add("language", req.language)
	ku.Add("model", req.model)
	ku.Add("brand", req.brand)
	ku.Add("countryCode", req.countryCode)
	ku.Add("countryName", req.countryName)
	ku.Add("region", req.region)
	ku.Add("city", req.city)
	ku.Add("carrier", req.carrier)
	ku.Add("isp", req.isp)
	ku.Add("os", req.os)
	ku.Add("osv", req.osVersion)
	ku.Add("browser", req.browser)
	ku.Add("browserv", req.browserVersion)
	ku.Add("connType", req.connectionType)
	ku.Add("bot", fmt.Sprintf("%v", req.bot))
	ku.Add("cpaValue", fmt.Sprintf("%f", req.cpaValue))
	//	req.tsExternalId
	//	if req.tsExternalId != nil {
	//		ku.Add("tsEId", req.tsExternalId.Encode())
	//	}
	//	req.tsCost
	//	if req.tsCost != nil {
	//		ku.Add("tsCost", req.tsCost.Encode())
	//	}
	//	req.tsVars
	if len(req.tsVars) > 0 {
		ku.Add("tsVars", common.EncodeParams(req.tsVars))
	}

	return base64.URLEncoding.EncodeToString([]byte(ku.Encode()))
}

func CacheStr2Req(caStr string) (req *reqbase) {
	if caStr == "" {
		return nil
	}
	bt, err := base64.URLEncoding.DecodeString(caStr)
	if err != nil {
		log.Errorf("DecodeString:%s failed:%v", caStr, err)
		return
	}
	//bc := xxtea.XxteaDecrypt(bt)
	bd, err := url.ParseQuery(string(bt))
	if err != nil {
		log.Errorf("ParseQuery:%s failed:%v", caStr, err)
		return
	}

	req = &reqbase{
		id: bd.Get("id"),
		t:  bd.Get("t"),
		ip: bd.Get("ip"),
		ua: bd.Get("ua"),

		externalId: bd.Get("externalId"),
		vars:       strings.Split(bd.Get("vars"), ";"),
		txid:       bd.Get("txId"),
		tsCId:      bd.Get("tsCId"),
		websiteId:  bd.Get("websiteId"),

		trafficSourceName: bd.Get("tsName"),
		campaignHash:      bd.Get("cHash"),
		campaignName:      bd.Get("cName"),
		campaignCountry:   bd.Get("cCountry"),
		landerName:        bd.Get("lName"),
		offerName:         bd.Get("oName"),
		affiliateName:     bd.Get("affName"),
		flowName:          bd.Get("flowName"),
		userIdText:        bd.Get("uIdText"),

		deviceType:     bd.Get("dType"),
		trackingDomain: bd.Get("trkDomain"),
		trackingPath:   bd.Get("trkPath"),
		referrer:       bd.Get("ref"),
		referrerdomain: bd.Get("refDomain"),
		language:       bd.Get("language"),
		model:          bd.Get("model"),
		brand:          bd.Get("brand"),
		countryCode:    bd.Get("countryCode"),
		countryName:    bd.Get("countryName"),
		region:         bd.Get("region"),
		city:           bd.Get("city"),
		carrier:        bd.Get("carrier"),
		isp:            bd.Get("isp"),
		os:             bd.Get("os"),
		osVersion:      bd.Get("osv"),
		browser:        bd.Get("browser"),
		browserVersion: bd.Get("browserv"),
		connectionType: bd.Get("connType"),
		cookie:         make(map[string]string),
		urlParam:       make(map[string]string),
	}

	req.cost, _ = strconv.ParseFloat(bd.Get("cost"), 64)
	req.payout, _ = strconv.ParseFloat(bd.Get("payout"), 64)
	req.impTimeStamp, _ = strconv.ParseInt(bd.Get("impTs"), 10, 64)
	req.visitTimeStamp, _ = strconv.ParseInt(bd.Get("visitTs"), 10, 64)
	req.clickTimeStamp, _ = strconv.ParseInt(bd.Get("clickTs"), 10, 64)
	req.postbackTimeStamp, _ = strconv.ParseInt(bd.Get("pbTs"), 10, 64)
	req.trafficSourceId, _ = strconv.ParseInt(bd.Get("tsId"), 10, 64)
	req.userId, _ = strconv.ParseInt(bd.Get("uId"), 10, 64)
	req.campaignId, _ = strconv.ParseInt(bd.Get("cId"), 10, 64)
	req.flowId, _ = strconv.ParseInt(bd.Get("fId"), 10, 64)
	req.ruleId, _ = strconv.ParseInt(bd.Get("rId"), 10, 64)
	req.pathId, _ = strconv.ParseInt(bd.Get("pId"), 10, 64)
	req.landerId, _ = strconv.ParseInt(bd.Get("lId"), 10, 64)
	req.offerId, _ = strconv.ParseInt(bd.Get("oId"), 10, 64)
	req.optOfferId, _ = strconv.ParseInt(bd.Get("oOId"), 10, 64)
	req.affiliateId, _ = strconv.ParseInt(bd.Get("affId"), 10, 64)
	req.optAffiliateId, _ = strconv.ParseInt(bd.Get("oAffId"), 10, 64)
	req.bot, _ = strconv.ParseBool(bd.Get("bot"))
	req.cpaValue, _ = strconv.ParseFloat(bd.Get("cpaValue"), 64)
	//	req.tsExternalId = &common.TrafficSourceParams{}
	//	req.tsExternalId.Decode(bd.Get("tsEId"))
	//	req.tsCost = &common.TrafficSourceParams{}
	//	req.tsCost.Decode(bd.Get("tsCost"))
	req.tsVars = common.DecodeParams(bd.Get("tsVars"))
	return
}
