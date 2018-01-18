package units

import (
	"Service/config"
	"Service/db"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"Service/common"
	"Service/log"
	"Service/request"
	"Service/tracking"
	"Service/units/affiliate"
	"Service/units/blacklist"
	"Service/units/campaign"
	"Service/units/ffrule"
	"Service/units/offer"
	"Service/units/user"
	"Service/util/ip"
	"strings"
)

func Init() (err error) {
	err = user.InitAllUsers()
	if err != nil {
		return
	}
	err = ffrule.Start(ffrule.ModeProducer)
	if err != nil {
		return
	}
	started = true
	return nil
}

/**
 * Request处理
**/
var page404 = `<html><head><title>Error: Page not found. If you want to change the content of this page, go to your account Settings / Root domain.</title></head><body><h3>Error 404</h3><p>Page not found. If you want to change the content of this page, go to your account Settings / Root domain.</p></body></html>`
var domainNotAssociated = `<html><head><title>Error: Domain %s is not associated with your account. If you want to handle this traffic, you need to configure the domain in the Settings tab.</title></head><body><h3>Error 404</h3><p>Domain %s is not associated with your account. If you want to handle this traffic, you need to configure the domain in the Settings tab.</p></body></html>`

func isCampaignHashValid(h string) bool {
	for _, c := range h {
		// 4801780f-ae8b-463d-be39-e1f51d805913
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' || c == '-') {
			return false
		}
	}
	return true
}

func OnLPOfferRequest(w http.ResponseWriter, r *http.Request) {
	log.Infof("[Units][OnLPOfferRequest] campaign url: %s\n", r.URL.String())
	if !started {
		log.Errorf("[Units][OnLPOfferRequest]Not started for :%s\n", common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr := ip.GetIP(r)
	desc, in := blacklist.G.AddrIn(remoteAddr)
	if in {
		log.Warnf("[units][OnLPOfferRequest] RemoteAddr:%v is blocked by:%v", remoteAddr, desc)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestId, _ := common.GenClickId()
	log.Infof("[Units][OnLPOfferRequest]Received request %s:%s\n", requestId, common.SchemeHostURI(r))
	if !started {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	domain := common.HostWithoutPort(r)
	u := user.GetUserByDomain(domain)
	if u == nil {
		log.Errorf("[Units][OnLPOfferRequest]Invalid userdomain:%s for %s:%s\n", domain, requestId, common.SchemeHostURI(r))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, fmt.Sprintf(domainNotAssociated, domain, domain))
		return
	}

	if !u.Active() {
		log.Errorf("[Units][OnLPOfferRequest]User not active for %s:%s\n", requestId, common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	allowed := blacklist.UserReqAllowed(u.Id, remoteAddr, r.UserAgent())
	if !allowed {
		log.Errorf("[Units][OnLPOfferRequest]User(%d) does not accept %s with ua(%s) for %s:%s\n", u.Id, remoteAddr, r.UserAgent(), requestId, common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaignHash := common.GetCampaignHash(r)
	if campaignHash == "" || !isCampaignHashValid(campaignHash) {
		log.Errorf("[Units][OnLPOfferRequest]Invalid campaignHash for %s:%s hash:%s\n", requestId, common.SchemeHostURI(r), campaignHash)
		if u.RootDomainRedirect == "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, page404)
		} else {
			http.Redirect(w, r, u.RootDomainRedirect, http.StatusMovedPermanently)
		}
		return
	}

	var req request.Request
	var err error
	if r.URL.Query().Encode() == "" { // 没有任何tracking token传递过来，就只能尝试从cookie中获取
		req, err = ParseCookie(request.ReqLPOffer, r)
	}

	if req == nil || err != nil { // 从cookie解析并load request失败的话，再从url的方式获取
		req, err = request.CreateRequest(requestId, true, request.ReqLPOffer, r)
		if req == nil || err != nil {
			log.Errorf("[Units][OnLPOfferRequest]CreateRequest failed for %s;%s;%v\n", requestId, common.SchemeHostURI(r), err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	req.SetUserId(u.Id)
	req.SetVisitTimeStamp(time.Now().UnixNano() / int64(time.Millisecond))

	req.SetCampaignHash(campaignHash)
	var ca *campaign.Campaign
	if req.CampaignId() > 0 { // 如果是从cache中获取campaignId的话，需要对比下campaignHash和campaignId是否匹配
		ca = campaign.GetCampaignByHash(campaignHash)
		if ca == nil {
			log.Errorf("[Units][OnImpression]Invalid campaignHash for %s:%s:%s\n", requestId, common.SchemeHostURI(r), campaignHash)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ca.Id != req.CampaignId() {
			log.Errorf("[Units][OnImpression]CampaignHash(%s) does not match existing campaignId(%d) for %s:%s:%s\n",
				campaignHash, req.CampaignId(), requestId, common.SchemeHostURI(r), campaignHash)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if r.URL.Query().Encode() != "" { // 如果UrlToken非空的话，需要通过campaign拿到traffic source，然后拿到其参数配置格式
		if ca == nil {
			ca = campaign.GetCampaignByHash(campaignHash)
			if ca == nil {
				log.Errorf("[Units][OnImpression]Invalid campaignHash for %s:%s:%s\n", requestId, common.SchemeHostURI(r), campaignHash)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		// 解析参数，传递到request中
		req.ParseTSParams(ca.TrafficSource.ExternalId,
			ca.TrafficSource.Cost,
			ca.TrafficSource.Vars,
			ca.TrafficSource.TSCampaignId,
			ca.TrafficSource.TSWebsitId,
			r.URL.Query())
	}

	theirCamp := r.FormValue("campaignid")
	if len(theirCamp) != 0 {
		tracking.InsertCampMap(ca.Id, theirCamp)
	}

	SetCookie(w, request.ReqLPOffer, req)

	if err := u.OnLPOfferRequest(w, req); err != nil {
		log.Errorf("[Units][OnLPOfferRequest]user.OnLPOfferRequest failed for %s;%s\n", req.String(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if req.OfferId() > 0 {
	// 	// Clicks增加
	// 	timestamp := tracking.Timestamp()
	// 	tracking.AddClick(req.AdStatisKey(timestamp), 1)
	// 	tracking.IP.AddClick(req.IPKey(timestamp), 1)
	// 	tracking.Domain.AddClick(req.DomainKey(timestamp), 1)
	// 	tracking.Ref.AddClick(req.ReferrerKey(timestamp), 1)

	// } else {
	// Visits增加
	timestamp := tracking.Timestamp()
	tracking.AddVisit(req.AdStatisKey(timestamp), 1)
	tracking.IP.AddVisit(req.IPKey(timestamp), 1)
	tracking.Domain.AddVisit(req.DomainKey(timestamp), 1)
	tracking.Ref.AddVisit(req.ReferrerKey(timestamp), 1)
	// }

	remoteCacheTime := time.Duration(-1)
	if req.OfferId() > 0 || campaign.GetCampaign(req.CampaignId()).TargetType == campaign.TargetTypeUrl {
		// 如果已经涉及到Offer，或者Campaign是直接打到某个Url，则保存时间变长，用于后面的Postback动作
		remoteCacheTime = config.ClickCacheTime
	}
	if !req.CacheSave(config.ReqCacheTime, remoteCacheTime) {
		log.Errorf("[Units][OnLPOfferRequest]req.CacheSave() failed for %s:%s\n", req.String(), common.SchemeHostURI(r))
	}
}

func OnLandingPageClick(w http.ResponseWriter, r *http.Request) {
	if !started {
		log.Errorf("[Units][OnLandingPageClick]Not started for :%s\n", common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	remoteAddr := ip.GetIP(r)
	desc, in := blacklist.G.AddrIn(remoteAddr)
	if in {
		log.Warnf("[units][OnLandingPageClick] RemoteAddr:%v is blocked by:%v", remoteAddr, desc)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req, err := resolveRequest(request.ReqLPClick, r)
	if err != nil || req == nil {
		//TODO add error log
		log.Errorf("resolveRequest failed Cookies:%+v err:%v for :%s\n", r.Cookies(), err, common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.CampaignId() <= 0 ||
		req.FlowId() <= 0 ||
		req.RuleId() <= 0 ||
		req.PathId() <= 0 ||
		req.LanderId() <= 0 {
		log.Errorf("[Units][OnLandingPageClick]CampaignId|FlowId|RuleId|PathId|LanderId is 0 for %s:%s\n",
			req.Id(), common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof("[Units][OnLandingPageClick]Received request for %s:%s:%s\n", req.Id(), req.CampaignHash(), common.SchemeHostURI(r))

	domain := common.HostWithoutPort(r)
	u := user.GetUserByDomain(domain)
	if u == nil {
		log.Errorf("[Units][OnLandingPageClick]Invalid userdomain:%s for %s:%s\n", domain, req.Id(), common.SchemeHostURI(r))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, fmt.Sprintf(domainNotAssociated, domain, domain))
		return
	}
	if !u.Active() {
		log.Errorf("[Units][OnLandingPageClick]User not active for %s:%s\n", req.Id(), common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.SetUserId(u.Id)
	req.SetUserIdText(u.IdText)
	req.SetClickTimeStamp(time.Now().UnixNano() / int64(time.Millisecond))

	SetCookie(w, request.ReqLPClick, req)

	if err := u.OnLandingPageClick(w, req); err != nil {
		log.Errorf("[Units][OnLandingPageClick]user.OnLandingPageClick failed for %s;%s\n", req.String(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 统计信息的添加
	timestamp := tracking.Timestamp()
	tracking.AddClick(req.AdStatisKey(timestamp), 1)
	tracking.IP.AddClick(req.IPKey(timestamp), 1)
	tracking.Domain.AddClick(req.DomainKey(timestamp), 1)
	tracking.Ref.AddClick(req.ReferrerKey(timestamp), 1)

	remoteCacheTime := time.Duration(-1)
	if req.OfferId() > 0 {
		// 如果已经涉及到Offer，则保存时间变长
		remoteCacheTime = config.ClickCacheTime
	}
	if !req.CacheSave(config.ReqCacheTime, remoteCacheTime) {
		log.Errorf("[Units][OnLandingPageClick]req.CacheSave() failed for %s:%s\n", req.String(), common.SchemeHostURI(r))
	}
}

func OnImpression(w http.ResponseWriter, r *http.Request) {
	// URL格式：http://zx1jg.voluumtrk.com/impression/be8da5d9-7955-4400-95e3-05c9231a6e92?keyword={keyword}&keyword_id={keyword_id}&creative_id={creative_id}&campaign_id={campaign_id}&country={country}&bid={bid}&click_id={click_id}
	// 1. 通过链接拿到user和campaign，以及trafficsource
	// 2. 通过IP拿到其它信息，如Language, Model, Country, City, ....
	// 3. 根据参数解析v1-v10
	// 4. 增加统计信息，结束

	requestId, _ := common.GenClickId()
	log.Infof("[Units][OnImpression]Received request %s:%s\n", requestId, common.SchemeHostURI(r))

	// w.Header().
	// req.AddCookie("reqid", requestId)

	domain := common.HostWithoutPort(r)
	u := user.GetUserByDomain(domain)
	if u == nil {
		log.Errorf("[Units][OnLandingPageClick]Invalid userdomain:%s for %s:%s\n", domain, requestId, common.SchemeHostURI(r))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, fmt.Sprintf(domainNotAssociated, domain, domain))
		return
	}
	if !u.Active() {
		log.Errorf("[Units][OnImpression]User not active for %s:%s\n", requestId, common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	campaignHash := common.GetCampaignHash(r)
	if campaignHash == "" {
		log.Errorf("[Units][OnImpression]Invalid campaignHash for %s:%s\n", requestId, common.SchemeHostURI(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req, err := request.CreateRequest(requestId, true, request.ReqImpression, r)
	if req == nil || err != nil {
		log.Errorf("[Units][OnImpression]CreateRequest failed for %s;%s with err(%v)\n", requestId, common.SchemeHostURI(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.SetUserId(u.Id)
	req.SetUserIdText(u.IdText)
	req.SetCampaignHash(campaignHash)
	req.SetImpTimeStamp(time.Now().UnixNano() / int64(time.Millisecond))
	// if err := u.OnImpression(w, req); err != nil {
	// 	log.Errorf("[Units][OnImpression]user.OnImpression failed for %s;%s\n", req.String(), err.Error())
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// 通过campaign拿到traffic source，然后拿到其参数配置格式
	ca := campaign.GetCampaignByHash(campaignHash)
	if ca == nil {
		log.Errorf("[Units][OnImpression]Invalid campaignHash for %s:%s:%s\n", requestId, common.SchemeHostURI(r), campaignHash)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := ca.OnImpression(w, req); err != nil {
		log.Errorf("[Units][OnImpression]campaign.OnImpression failed for %s;%s\n", req.String(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.SetCampaignId(ca.Id)
	req.SetCampaignName(ca.Name)
	// 解析参数，传递到request中
	req.ParseTSParams(ca.TrafficSource.ExternalId,
		ca.TrafficSource.Cost,
		ca.TrafficSource.Vars,
		ca.TrafficSource.TSCampaignId,
		ca.TrafficSource.TSWebsitId,
		r.URL.Query())

	// 统计信息的添加
	timestamp := tracking.Timestamp()
	tracking.AddImpression(req.AdStatisKey(timestamp), 1)
	tracking.IP.AddImpression(req.IPKey(timestamp), 1)
	tracking.Domain.AddImpression(req.DomainKey(timestamp), 1)
	tracking.Ref.AddImpression(req.ReferrerKey(timestamp), 1)

	if ca.CostModel == 3 { // CPM
		user.TrackingCost(req, ca.CPMValue/1000.0)
	} else if ca.CostModel == 4 { // Auto
		user.TrackingCost(req, req.Cost())
	}

	SetCookie(w, request.ReqImpression, req)

	// OfferId不可能>0，所以只用保存一小段时间
	if !req.CacheSave(config.ReqCacheTime, -1) {
		log.Errorf("[Units][OnImpression]req.CacheSave() failed for %s:%s\n", req.String(), common.SchemeHostURI(r))
	}
}

// parseIP 从192.168.0.155:61233解析出192.168.0.155
func parseIP(remoteAddr string) string {
	pos := strings.Index(remoteAddr, ":")
	if pos == -1 {
		return remoteAddr
	}
	return remoteAddr[:pos]
}

func OnS2SPostback(w http.ResponseWriter, r *http.Request) {
	if !started {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originClickId := r.URL.Query().Get(common.UrlTokenClickId)
	clickId := originClickId
	underlinePos := strings.Index(clickId, "_")
	if underlinePos != -1 {
		clickId = clickId[:underlinePos]
	}

	payoutStr := r.URL.Query().Get(common.UrlTokenPayout)
	txId := r.URL.Query().Get(common.UrlTokenTransactionId)
	log.Infof("[Units][OnS2SPostback]Received postback with %s(%s;%s;%s)\n", common.SchemeHostURI(r), clickId, payoutStr, txId)

	req, err := request.CreateRequest(clickId, false, request.ReqS2SPostback, r)
	if req == nil || err != nil {
		log.Errorf("[Units][OnS2SPostback]CreateRequest failed for %s;%v\n", common.SchemeHostURI(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 后面统计信息要使用
	req.SetTransactionId(txId)
	log.Infof("[Units][OnS2SPostback]Received valid postback with %s(%s;%s;%s;%s)\n", common.SchemeHostURI(r), clickId, req.CampaignHash(), payoutStr, txId)

	if underlinePos != -1 {
		oid, err := strconv.ParseInt(originClickId[underlinePos+1:], 10, 64)
		if err != nil {
			log.Errorf("parse offer id from:%s failed:%v", originClickId, err)
		} else {
			originOfferId := req.OfferId()
			req.SetOfferId(oid)
			log.Infof("originClickId:%s originOfferId:%d newOfferId:%d", originClickId, originOfferId, oid)
		}
	}

	isFirstCallback, finalPayout, err := checkPostback(req, clickId, txId, payoutStr, r)
	if !isFirstCallback {
		log.Warnf("[Units][OnS2SPostback]clickId:%v txId:%v payoutStr:%v postbackurl:%v discarded", clickId, txId, payoutStr, r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 后面会用到payout，所以这里要提前设置好
	req.SetPayout(finalPayout)
	req.SetPostbackTimeStamp(time.Now().UnixNano() / int64(time.Millisecond))

	domain := common.HostWithoutPort(r)
	u := user.GetUserByDomain(domain)

	if u == nil {
		log.Errorf("[Units][OnS2SPostback]Invalid userdomain:%s for %s:%s\n", domain, clickId, common.SchemeHostURI(r))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, fmt.Sprintf(domainNotAssociated, domain, domain))
		return
	}

	if err := u.OnS2SPostback(w, req); err != nil {
		log.Errorf("[Units][OnS2SPostback]user.OnLPOfferRequest failed for %s;%s\n", req.String(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.TrackingRevenue(req, finalPayout)
	user.TrackingConversion(req, 1)

	// 统计conversion
	conv := req.ConversionKey()
	tracking.SaveConversion(&conv)

	remoteCacheTime := time.Duration(-1)
	if req.OfferId() > 0 || campaign.GetCampaign(req.CampaignId()).TargetType == campaign.TargetTypeUrl {
		// 如果已经涉及到Offer，或者Campaign是直接打到某个Url，则保存时间变长
		remoteCacheTime = config.ClickCacheTime
	}

	if !req.CacheSave(config.ReqCacheTime, remoteCacheTime) {
		log.Errorf("[Units][OnS2SPostback]req.CacheSave() failed for %s:%s\n", req.String(), common.SchemeHostURI(r))
	}
}

var uploadConvsFormat = regexp.MustCompile(`^[0-9a-zA-Z]+(\_[0-9]+)*(,\s*(\s*|[0-9]+\.?[0-9]*)\s*(,\s*[0-9a-zA-Z]*\s*)*)*$`)

func OnUploadConversions(w http.ResponseWriter, r *http.Request) {
	if !started {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	domain := common.HostWithoutPort(r)
	u := user.GetUserByDomain(domain)
	if u == nil {
		log.Errorf("[Units][OnS2SPostback]Invalid userdomain:%s for %s:%s\n", domain, common.SchemeHostURI(r))
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, fmt.Sprintf(domainNotAssociated, domain, domain))
		return
	}

	var err error
	if err = r.ParseForm(); err != nil {
		log.Errorf("[Units][OnUploadConversions]ParseForm failed for %s;%v\n", common.SchemeHostURI(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("[Units][OnUploadConversions]ReadAll body failed for %s;%v\n", common.SchemeHostURI(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vs := make([]string, 0)
	if err = json.Unmarshal(body, &vs); err != nil {
		log.Errorf("[Units][OnUploadConversions]json.Unmarshal body failed for %s;%v\n", common.SchemeHostURI(r), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	type Record struct {
		I int    // line number, 0~
		V string // line content
		E string // error message
	}
	errRecord := make([]Record, 0)
	for i, v := range vs {
		// v:cccccccccccccclickId[_oId][,[cost][,[txId]]]
		/* 合法的Case
		* 71442644828b17b36cc87266dad9bae6
		* 71442644828b17b36cc87266dad9bae6,
		* 71442644828b17b36cc87266dad9bae6,,
		* 71442644828b17b36cc87266dad9bae6,10.0
		* 71442644828b17b36cc87266dad9bae6,10.0,
		* 71442644828b17b36cc87266dad9bae6,10.0,1234abc
		* 71442644828b17b36cc87266dad9bae6,,1234abc
		* 71442644828b17b36cc87266dad9bae6_74
		* 71442644828b17b36cc87266dad9bae6_74,
		* 71442644828b17b36cc87266dad9bae6_74,,
		* 71442644828b17b36cc87266dad9bae6_74,10.0
		* 71442644828b17b36cc87266dad9bae6_74,10.0,
		* 71442644828b17b36cc87266dad9bae6_74,10.0,1234abc
		* 71442644828b17b36cc87266dad9bae6_74,,1234abc
		 */
		if !uploadConvsFormat.Match([]byte(v)) { // format check
			errRecord = append(errRecord, Record{i, vs[i], "invalid format"})
			continue
		}
		vi := strings.Split(strings.Replace(v, " ", "", -1), ",")
		var co, clickId, cost, txId string
		var offerId int64
		switch len(vi) {
		case 0:
			errRecord = append(errRecord, Record{i, vs[i], "invalid format"})
			continue
		case 1:
			co = vi[0]
		case 2:
			co = vi[0]
			cost = vi[1]
		default: // >=3
			co = vi[0]
			cost = vi[1]
			txId = vi[2]
		}
		log.Infof("[Units][OnUploadConversions]Received conversions with %s(%s;%s;%s)\n", common.SchemeHostURI(r), co, cost, txId)
		coi := strings.Split(co, "_")
		clickId = coi[0]
		//TODO 在线上所有的clickId，都转化为aes clickId之前，先屏蔽这个检查 2017/3/21
		/*if !common.ValidClickId(clickId) {
			log.Errorf("[Units][OnUploadConversions]Invalid clickId for %s:%s;%v\n", common.SchemeHostURI(r), clickId, err)
			errRecord = append(errRecord, Record{i, vs[i], "invalid data"})
			continue
		}*/
		if len(coi) > 1 {
			offerId, _ = strconv.ParseInt(coi[1], 10, 64)
		}

		req, err := request.CreateRequest(clickId, false, request.ReqUploadConversions, r)
		if req == nil || err != nil {
			log.Errorf("[Units][OnUploadConversions]CreateRequest failed for %s:%s;%v\n", common.SchemeHostURI(r), clickId, err)
			errRecord = append(errRecord, Record{i, vs[i], "invalid data"})
			continue
		}
		if req.OfferId() <= 0 { // 如果之前的click，还没有达到offer的步骤，则忽略掉conversion上传
			log.Errorf("[Units][OnUploadConversions]request.OfferId(%d) is invalid for conversion upload for %s;%s\n", req.OfferId(), common.SchemeHostURI(r), vi[0])
			errRecord = append(errRecord, Record{i, vs[i], "invalid data"})
			continue
		}

		// 后面统计信息要使用
		req.SetTransactionId(txId)

		accept, finalPayout, err := checkUploadConversions(req, offerId, clickId, txId, cost, r)
		if !accept {
			log.Warnf("[Units][OnUploadConversions]clickId:%v txId:%v payoutStr:%v postbackurl:%v discarded", clickId, txId, cost, r.RequestURI)
			errRecord = append(errRecord, Record{i, vs[i], "duplicated conversion"})
			continue
		}

		// 后面会用到payout，所以这里要提前设置好
		if offerId > 0 {
			req.SetOfferId(offerId)
		}
		req.SetPayout(finalPayout)
		req.SetPostbackTimeStamp(time.Now().UnixNano() / int64(time.Millisecond))

		if err := u.OnS2SPostback(w, req); err != nil {
			log.Errorf("[Units][OnUploadConversions]user.OnS2SPostback failed for %s;%s\n", req.String(), err.Error())
			errRecord = append(errRecord, Record{i, vs[i], "invalid data"})
			continue
		}

		user.TrackingRevenue(req, finalPayout)
		user.TrackingConversion(req, 1)

		// 统计conversion
		conv := req.ConversionKey()
		tracking.SaveConversion(&conv)

		// OfferId()肯定>0，所以直接保存长时间的即可
		if !req.CacheSave(config.ReqCacheTime, config.ClickCacheTime) {
			log.Errorf("[Units][OnUploadConversions]req.CacheSave() failed for %s:%s\n", req.String(), common.SchemeHostURI(r))
		}
	}

	if len(errRecord) > 0 {
		// error lines exist
		resp, _ := json.Marshal(errRecord)
		w.Header().Set(common.KHttpContentType, common.KHttpContentTypeJson)
		w.Header().Set(common.KHttpContentLength, strconv.Itoa(len(resp)))
		w.Write(resp)
	}
}

func checkPostback(req request.Request, clickId, txId, payoutStr string, r *http.Request) (firstCallback bool, finalPayout float64, err error) {
	payout, err := strconv.ParseFloat(payoutStr, 64)
	if err != nil {
		log.Errorf("[Units][checkPostback]ParseFloat with payoutStr:%v failed for %s;%v\n", payoutStr, common.SchemeHostURI(r), err)
	}

	if req.OfferId() == 0 {
		// 说明是直接跳转URL
		isFirstCallback := func() bool {
			//TODO 重构成此处不需要直接访问redis
			svr := db.GetRedisClient(request.LocalCacheSvrTitle)
			k := fmt.Sprintf("postback:%s:tx:%s:off:%d", clickId, txId, req.OfferId())
			v := time.Now().Unix()
			cmd := svr.SetNX(k, v, config.ReqCacheTime)
			if cmd.Err() != nil {
				log.Errorf("[units][checkPostback] SetNX k:%v v:%v failed:%v", k, v, cmd.Err())
				return true
			}

			if cmd.Val() {
				// 首次postback
				log.Warnf("firsttime postback:k:%v v:%v", k, v)
				return true
			}

			log.Warnf("Duplicate postback denied: clickId:%v txId:%v", clickId, txId)
			return false
		}()

		return isFirstCallback, payout, nil
	}

	isFirstCallback := func() bool {
		o := offer.GetOffer(req.OfferId())
		if o == nil {
			log.Errorf("GetOffer:%v failed clickId:%v", req.OfferId(), clickId)
			return true
		}

		aff := affiliate.GetAffiliateNetwork(o.AffiliateNetworkId)
		if aff == nil {
			log.Errorf("GetAffiliateNetwork:%v failed clickId:%v", o.AffiliateNetworkId, clickId)
			return true
		}

		remoteAddr := ip.GetIP(r)
		ip := parseIP(remoteAddr)
		if !aff.Allow(ip) {
			log.Warnf("AffiliateNetworkId:%v blocked postback from:%v by it's white-listed IPs clickId:%v",
				o.AffiliateNetworkId, remoteAddr, clickId)
			return false
		}

		if aff.DuplicatePostback != 0 {
			// 允许多次callback
			return true
		}

		// 一天时间
		//TODO 重构成此处不需要直接访问redis
		svr := db.GetRedisClient(request.LocalCacheSvrTitle)
		k := fmt.Sprintf("postback:%s:tx:%s:off:%d", clickId, txId, req.OfferId())
		v := time.Now().Unix()
		cmd := svr.SetNX(k, v, config.ReqCacheTime)
		if cmd.Err() != nil {
			log.Errorf("[units][checkPostback] SetNX k:%v v:%v failed:%v", k, v, cmd.Err())
			return true
		}

		if cmd.Val() {
			// 首次postback
			log.Warnf("firsttime postback:k:%v v:%v", k, v)
			return true
		}

		log.Warnf("Duplicate postback denied: clickId:%v txId:%v", clickId, txId)
		return false
	}()

	// 统计payout
	finalPayout = func() float64 {
		o := offer.GetOffer(req.OfferId())
		if o == nil {
			// 完全有可能是Campaign中自定义Url发生的Postback
			log.Errorf("[Units][checkPostback] offer.GetOffer(%v) failed: no offer found", req.OfferId())
			return payout
		}
		// 并不优先使用postback回传的payout，严格按照用户的设定来做
		switch o.PayoutMode {
		case 0:
			return payout
		case 1:
			return o.PayoutValue
		}
		return 0.0
	}()

	return isFirstCallback, finalPayout, nil
}

func checkUploadConversions(req request.Request, offerId int64, clickId, txId, payoutStr string, r *http.Request) (accept bool, finalPayout float64, err error) {
	payout, err := strconv.ParseFloat(payoutStr, 64)
	if err != nil {
		log.Errorf("[Units][checkPostback]ParseFloat with payoutStr:%v failed for %s;%v\n", payoutStr, common.SchemeHostURI(r), err)
	}

	if req.OfferId() <= 0 {
		// offerId 不能为0
		return false, payout, errors.New("invalid request offer id")
	}
	o := offer.GetOffer(req.OfferId())
	if o == nil {
		return false, payout, fmt.Errorf("GetOffer:%v failed clickId:%v", req.OfferId(), clickId)
	}

	accept = func() bool {
		aff := affiliate.GetAffiliateNetwork(o.AffiliateNetworkId)
		if aff == nil {
			log.Errorf("GetAffiliateNetwork:%v failed clickId:%v", o.AffiliateNetworkId, clickId)
			return true
		}

		remoteAddr := ip.GetIP(r)
		ip := parseIP(remoteAddr)
		if !aff.Allow(ip) {
			log.Warnf("AffiliateNetworkId:%v blocked postback from:%v by it's white-listed IPs clickId:%v",
				o.AffiliateNetworkId, remoteAddr, clickId)
			return false
		}

		if req.PostBackTimeStamp() == 0 {
			// 没有发生过postback，则直接通过
			// 如果不做这个检查，后续的txId和OfferId相等的判断会存在漏洞
			return true
		}

		if aff.DuplicatePostback != 0 {
			// 允许多次callback
			return true
		}

		if req.TransactionId() == txId && (req.OfferId() == offerId || offerId == 0) {
			// 当txId和offerId都和之前的相等(或者新的offerId为0)的时候，表明是重复的转化
			//TODO 当出现clickIdAAA,txIdBBB,OfferId1 -> clickIdAAA,txIdBBB,OfferId2 -> clickIdAAA,txIdBBB,OfferId1
			// 这种case时，会记录两次clickIdAAA,txIdBBB,OfferId1的转化
			log.Warnf("Duplicate conversion denied: clickId:%v txId:%v offerId:%d\n", clickId, txId, offerId)
			return false
		}

		return true
	}()

	// 统计payout
	finalPayout = func() float64 {
		// 并不优先使用postback回传的payout，严格按照用户的设定来做
		switch o.PayoutMode {
		case 0:
			return payout
		case 1:
			return o.PayoutValue
		}
		return 0.0
	}()

	return accept, finalPayout, nil
}

func OnConversionPixel(w http.ResponseWriter, r *http.Request) {
	if !started {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func OnConversionScript(w http.ResponseWriter, r *http.Request) {
	if !started {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

// OnDoubleMetaRefresh 处理double meta refresh 请求
func OnDoubleMetaRefresh(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	dest := r.FormValue("dest") // 最终的跳转地址
	log.Debugf("OnDoubleMetaRefresh dest:%s", dest)

	w.Header().Set("Content-Type", "text/html")
	meta := `<meta http-equiv="refresh" content="0;url=` + html.EscapeString(dest) + `">`
	fmt.Fprintln(w, meta)
}

// 会优先尝试从url query中获取click id，失败的话再从cookie中获取
func resolveRequest(step string, r *http.Request) (req request.Request, err error) {
	switch step {
	case request.ReqLPClick:
	//OK
	default:
		return nil, fmt.Errorf("unsupported step(%s)", step)
	}

	//TODO 优先使用url params里面带过来的requestId
	reqId := r.URL.Query().Get("cid")
	if reqId == "" {
		return ParseCookie(step, r)
	}

	req, err = request.CreateRequest(reqId, false, step, r)
	if req == nil || err != nil {
		return nil, fmt.Errorf("createRequest error(%v) in step(%s)", err, step)
	}

	log.Infof("[resolveRequest]Request(%s) in step(%s) with url(%s)\n",
		req.Id(), step, common.SchemeHostURI(r))

	return
}
