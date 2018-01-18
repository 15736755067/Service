package ffrule

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"Service/common"
	"Service/config"
	"Service/db"
	"Service/log"
	"Service/request"
	"Service/units/blacklist"
	"Service/util/ipcmp"
)

const (
	ConditionHeaderPV        = "PV"
	ConditionHeaderUserAgent = "USERAGENT"
	ConditionHeaderClick     = "CLICKS"
)
const (
	ReqTypeImpression = 0x1
	ReqTypeVisit      = 0x2
	ReqTypeClick      = 0x4
)
const (
	KeyHeaderPV        = "PV"
	KeyHeaderUserAgent = "UA"
	KeyHeaderClick     = "CK"
)

type Condition struct {
	T     int    // 用来快速判断是否需要应用于某种case;是ReqType的按位或
	Key   string // all upper
	Op    string // all upper;只有可能是>符号
	Value int    // all upper
}

func ParseConditions(sCondition string) (conditions []Condition) {
	v := int64(0)
	var err error
	sCondition = strings.ToUpper(sCondition)
	for _, c := range strings.Split(sCondition, ",") {
		op := ""
		switch {
		case strings.Contains(c, ">"):
			op = ">"
		case strings.Contains(c, "<"): // 不可能
			op = "<"
		case strings.Contains(c, "="): // 不可能
			op = "="
		default:
			log.Errorf("ParseConditions:Invalid operation(%s)\n", sCondition)
			return nil
		}
		if op != "" {
			splitC := strings.Split(c, op)
			if len(splitC) < 2 {
				log.Errorf("ParseConditions:Invalid conditions(%s) elements not enough\n", sCondition)
				return nil
			}
			v, err = strconv.ParseInt(splitC[1], 10, 64)
			if err != nil {
				log.Errorf("ParseConditions:Invalid conditions(%s) element[1] not valid integer\n", sCondition)
				return nil
			}
			t := 0
			switch splitC[0] {
			case ConditionHeaderPV:
				t = ReqTypeImpression | ReqTypeVisit
			case ConditionHeaderUserAgent:
				t = ReqTypeImpression | ReqTypeVisit
			case ConditionHeaderClick:
				t = ReqTypeClick
			default:
				log.Errorf("ParseConditions:Invalid conditions(%s) element[0] not valid header\n", sCondition)
				return nil
			}
			conditions = append(conditions, Condition{t, splitC[0], op, int(v)})
		}
	}

	return
}

func (c Condition) ReferKey(reqType int, req request.Request) (head, tail string, value int64) {
	if c.T&reqType == 0 {
		// 该Condition不需要处理这种请求类型
		return
	}
	sipInt := ""
	ip, err := ipcmp.IPToInt64(req.RemoteIp())
	if err == nil {
		sipInt = fmt.Sprintf("%d", ip)
	}
	switch c.Key {
	case ConditionHeaderPV:
		if reqType == ReqTypeImpression {
			head = KeyHeaderPV
			tail = sipInt
			value = req.ImpTimeStamp() / 1000
		}
		if reqType == ReqTypeVisit {
			head = KeyHeaderPV
			tail = sipInt
			value = req.VisitTimeStamp() / 1000
		}
	case ConditionHeaderUserAgent:
		if reqType == ReqTypeImpression {
			head = KeyHeaderUserAgent
			tail = sipInt + "_" + base64.URLEncoding.EncodeToString([]byte(req.UserAgent()))
			value = req.ImpTimeStamp() / 1000
		}
		if reqType == ReqTypeVisit {
			head = KeyHeaderUserAgent
			tail = sipInt + "_" + base64.URLEncoding.EncodeToString([]byte(req.UserAgent()))
			value = req.VisitTimeStamp() / 1000
		}
	case ConditionHeaderClick:
		if reqType == ReqTypeClick {
			head = KeyHeaderClick
			tail = sipInt
			value = req.ClickTimeStamp() / 1000
		}
	}
	return
}

func (c Condition) Hit(header string, timeSpan int64, ts []int64) (hit bool) {
	switch header {
	case KeyHeaderPV:
		if c.Key != ConditionHeaderPV {
			return false
		}
	case KeyHeaderClick:
		if c.Key != ConditionHeaderClick {
			return false
		}
	case KeyHeaderUserAgent:
		if c.Key != ConditionHeaderUserAgent {
			return false
		}
	default:
		return false
	}

	for i := 0; i < len(ts); i++ {
		if i+c.Value >= len(ts) {
			return false
		}
		if ts[i+c.Value]-ts[i] < timeSpan {
			return true
		}
	}

	return false
}

type kv struct {
	k string // (PV|UA|CK)_RuleId_CampaignId_IPINT[_UABase64Encoded]
	v int64  // timeStamp to seconds
}

type RuleConfig struct {
	Id          int64
	Name        string
	Active      bool
	UserId      int64
	Dimension   string
	TimeSpan    int64
	Conditions  string // 'PV>500,UserAgent>100,Clicks>100'
	CampaignIds []int64
}

func (rc RuleConfig) GetCopy() (cc RuleConfig) {
	cc = rc
	cc.CampaignIds = make([]int64, len(rc.CampaignIds))
	for i := range rc.CampaignIds {
		cc.CampaignIds[i] = rc.CampaignIds[i]
	}
	return
}

func (rc RuleConfig) ID() int64 {
	return rc.Id
}

type Rule struct {
	Id          int64
	Name        string
	Active      bool
	UserId      int64
	Dimension   string // "IP"
	TimeSpan    int64  // in seconds
	Conditions  []Condition
	CampaignIds []int64

	t        int // 用来快速判断是否需要应用于某种case;是ReqType的按位或
	syncChan chan kv
}

func (r *Rule) onRequest(reqType int, req request.Request) {
	if r == nil {
		return
	}

	log.Debugf("rule:%+v\n", *r)
	if reqType&r.t == 0 {
		// 该rule不需要处理这种case
		return
	}

	head, tail := "", ""
	body := fmt.Sprintf("_%d_%d_", r.Id, req.CampaignId())
	value := int64(0)
	referKeys := make([]string, 0)
	for _, c := range r.Conditions {
		head, tail, value = c.ReferKey(reqType, req)
		if head != "" {
			referKeys = append(referKeys, head+body+tail)
		}
	}
	for _, key := range referKeys {
		select {
		case r.syncChan <- kv{key, value}: // do not block here
		default:
			log.Error("ffrule.OnImpression syncChan failed for", key, value)
		}
	}
}

func (r *Rule) OnImpression(req request.Request) {
	r.onRequest(ReqTypeImpression, req)
}

func (r *Rule) OnVisits(req request.Request) {
	r.onRequest(ReqTypeVisit, req)
}

func (r *Rule) OnClick(req request.Request) {
	r.onRequest(ReqTypeClick, req)
}

// header:PV/UA/CK;ts:timestamps(in second)
// hit:hit or not;popN:0~N,items should be poped out,always [0,len(ts)]
func (r Rule) Hit(header string, ts []string) (hit bool, popN int64) {
	if !r.Active {
		// 如果Rule已经变成非Active的，则将已有的记录都pop掉，不需要了
		return false, int64(len(ts))
	}

	t := time.Now().Unix() - r.TimeSpan
	ti := make([]int64, len(ts))
	k := int64(0)
	var err error
	for i, s := range ts {
		k, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(err.Error())
		}
		ti[i] = k
		if k > t {
			continue
		}
		popN++
	}

	for _, c := range r.Conditions {
		if !c.Hit(header, r.TimeSpan, ti) {
			return false, popN
		}
	}
	// 如果已经命中，则把全部的数据都pop出来
	return true, int64(len(ts))
}

var cmu sync.RWMutex // protects the following
var rules = make(map[int64]*Rule)

func setRule(r *Rule) error {
	if r == nil {
		return errors.New("setRule error:r is nil")
	}
	if r.Id <= 0 {
		return fmt.Errorf("setRule error:r.Id(%d) is not positive", r.Id)
	}
	cmu.Lock()
	defer cmu.Unlock()
	rules[r.Id] = r
	return nil
}
func getRule(rId int64) *Rule {
	cmu.RLock()
	defer cmu.RUnlock()
	return rules[rId]
}
func delRule(rId int64) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(rules, rId)
}

func InitAllRules() error {
	rules := dbGetAvailableRules()
	for _, c := range rules {
		r := newRule(c)
		if r == nil {
			return fmt.Errorf("newRule failed with rule(%d)", c.Id)
		}
		setRule(r)
	}
	return nil
}
func InitRule(ruleId int64) error {
	r := newRule(DBGetRule(ruleId))
	if r == nil {
		return fmt.Errorf("newRule failed with rule(%d)", ruleId)
	}
	return setRule(r)
}
func DeleteRule(ruleId int64) {
	delRule(ruleId)
}
func GetRule(ruleId int64) (r *Rule) {
	if ruleId == 0 {
		return
	}
	r = getRule(ruleId)
	if r == nil {
		r = newRule(DBGetRule(ruleId))
		if r == nil {
			return
		}
		if err := setRule(r); err != nil {
			return nil
		}
	}
	return
}

func newRule(c RuleConfig) (r *Rule) {
	log.Debugf("[newRule]%+v\n", c)
	r = &Rule{
		Id:          c.Id,
		Name:        c.Name,
		Active:      c.Active,
		UserId:      c.UserId,
		Dimension:   c.Dimension,
		TimeSpan:    c.TimeSpan,
		Conditions:  ParseConditions(c.Conditions),
		CampaignIds: c.CampaignIds,
		syncChan:    kvSyncChan,
	}
	for _, c := range r.Conditions {
		r.t |= c.T
	}
	return
}

func requestsProducing(syncChan <-chan kv) {
	defer func() {
		if x := recover(); x != nil {
			log.Error("ffrule.requestsProducing", x, string(debug.Stack()))
		}
	}()

	redisCli := db.GetRedisClient("LOCALFFCACHE")
	if redisCli == nil {
		panic("redisCli is nil")
	}
	var err error
	for x := range syncChan {
		if err = redisCli.RPush(x.k, x.v).Err(); err != nil {
			log.Error("ffrule.requestsProducing", x, err.Error())
		}
	}
}

var ffInterval = 10

// keep reading kv requests from redis,
// and checking conditions,
// and updating ffrule logs & black lists
// and notifying blacklist updates if there is any
func requestHandling() {
	defer func() {
		if x := recover(); x != nil {
			log.Error("ffrule.requestsProducing", x, string(debug.Stack()))
		}
	}()

	ticker := time.NewTicker(time.Second * time.Duration(ffInterval))
	redisCli := db.GetRedisClient("LOCALFFCACHE")
	if redisCli == nil {
		panic("redisCli is nil")
	}
	var err error
	var pvKeys, ckKeys, uaKeys []string
	var pvTS, ckTS, uaTS []string
	var ruleId, campaignId, ip int64
	var ips string
	var uab []byte
	var hit bool
	var popN int64
	var timeStamp int64
	for range ticker.C {
		userList := []userBotBlacklist{}
		timeStamp = time.Now().Unix()
		pvKeys, err = redisCli.Keys(KeyHeaderPV + "_*").Result()
		if err != nil {
			panic(err.Error())
		}
		for _, key := range pvKeys {
			pvTS, err = redisCli.LRange(key, 0, -1).Result()
			if err != nil {
				panic(err.Error())
			}

			ks := strings.Split(key, "_")
			ruleId, _ = strconv.ParseInt(ks[1], 10, 64)
			rule := GetRule(ruleId)
			if rule == nil { // 说明Rule已经被删除掉，直接pop掉所有相关的记录即可
				hit, popN = false, int64(len(pvTS))
			} else {
				hit, popN = rule.Hit(KeyHeaderPV, pvTS)
			}
			log.Debug("requestHandling.PV:", key, "hit:", hit, "popN:", popN)
			if hit {
				//TODO hit的时候应该del key
				campaignId, _ = strconv.ParseInt(ks[2], 10, 64)
				ip, _ = strconv.ParseInt(ks[3], 10, 64)
				ips, _ = ipcmp.Int64ToIP(ip)
				recordFFHitRecords(ruleId, campaignId, timeStamp, ips, nil)
				userList = append(userList, userBotBlacklist{
					name:    fmt.Sprintf("FFRule(%s) Generated", rule.Name),
					userId:  GetRule(ruleId).UserId,
					ipRange: []string{ips},
				})
			}
			if popN > 0 {
				// remove unnessory values from redis
				err = redisCli.LTrim(key, popN, -1).Err()
				if err != nil {
					panic(err.Error())
				}
			}
		}

		ckKeys, err = redisCli.Keys(KeyHeaderClick + "_*").Result()
		if err != nil {
			panic(err.Error())
		}
		for _, key := range ckKeys {
			ckTS, err = redisCli.LRange(key, 0, -1).Result()
			if err != nil {
				panic(err.Error())
			}

			ks := strings.Split(key, "_")
			ruleId, _ = strconv.ParseInt(ks[1], 10, 64)
			rule := GetRule(ruleId)
			if rule == nil { // 说明Rule已经被删除掉，直接pop掉所有相关的记录即可
				hit, popN = false, int64(len(ckTS))
			} else {
				hit, popN = rule.Hit(KeyHeaderClick, ckTS)
			}
			log.Debug("requestHandling.CK:", key, "hit:", hit, "popN:", popN)
			if hit {
				//TODO hit的时候应该del key
				campaignId, _ = strconv.ParseInt(ks[2], 10, 64)
				ip, _ = strconv.ParseInt(ks[3], 10, 64)
				ips, _ = ipcmp.Int64ToIP(ip)
				uab, _ = base64.URLEncoding.DecodeString(ks[4])
				recordFFHitRecords(ruleId, campaignId, timeStamp, ips, uab)
				userList = append(userList, userBotBlacklist{
					name:    fmt.Sprintf("FFRule(%s) Generated", rule.Name),
					userId:  GetRule(ruleId).UserId,
					ipRange: []string{ips},
				})
			}
			if popN > 0 {
				// remove unnessory values from redis
				err = redisCli.LTrim(key, popN, -1).Err()
				if err != nil {
					panic(err.Error())
				}
			}
		}

		uaKeys, err = redisCli.Keys(KeyHeaderUserAgent + "_*").Result()
		if err != nil {
			panic(err.Error())
		}

		for _, key := range uaKeys {
			uaTS, err = redisCli.LRange(key, 0, -1).Result()
			if err != nil {
				panic(err.Error())
			}

			ks := strings.Split(key, "_")
			ruleId, _ = strconv.ParseInt(ks[1], 10, 64)
			rule := GetRule(ruleId)
			if rule == nil { // 说明Rule已经被删除掉，直接pop掉所有相关的记录即可
				hit, popN = false, int64(len(uaTS))
			} else {
				hit, popN = rule.Hit(KeyHeaderUserAgent, uaTS)
			}
			log.Debug("requestHandling.UA:", key, "hit:", hit, "popN:", popN)
			if hit {
				//TODO hit的时候应该del key
				campaignId, _ = strconv.ParseInt(ks[2], 10, 64)
				ip, _ = strconv.ParseInt(ks[3], 10, 64)
				ips, _ = ipcmp.Int64ToIP(ip)
				uab, _ = base64.URLEncoding.DecodeString(ks[4])
				recordFFHitRecords(ruleId, campaignId, timeStamp, ips, uab)
				userList = append(userList, userBotBlacklist{
					name:      fmt.Sprintf("FFRule(%s) Generated", rule.Name),
					userId:    GetRule(ruleId).UserId,
					ipRange:   []string{ips},
					userAgent: []string{string(uab)},
				})
			}
			if popN > 0 {
				// remove unnessory values from redis
				err = redisCli.LTrim(key, popN, -1).Err()
				if err != nil {
					panic(err.Error())
				}
			}
		}

		if err = updateUserBlacklist(userList); err != nil {
			panic(err.Error())
		}

		if err = flushFFHitRecords(); err != nil {
			panic(err.Error())
		}
	}
}

type hitLog struct {
	RuleId     int64
	CampaignId int64
	TimeStamp  int64
	LogData    logData
}

type logData struct {
	IP []string `json:"ip"`
	UA []string `json:"ua"`
}

var hitRecords map[int64]hitLog

// 记录一下某条rule，对于某个campaign，在timeStamp的时间点，对ip和ua命中了
func recordFFHitRecords(ruleId, campaignId, timeStamp int64, ip string, ua []byte) {
	//TODO 使用缓存提高性能
	hlog := hitLog{
		RuleId:     ruleId,
		CampaignId: campaignId,
		TimeStamp:  timeStamp,
	}
	hlog.LogData = logData{
		IP: []string{ip},
		UA: []string{string(ua)},
	}

	logId, err := DBSaveFFHitRecords(hlog)
	if err != nil {
		log.Errorf("[recordFFHitRecords]save log fail ruleId(%d)-campaignId(%d)-timeStamp(%d)-ip(%s)-ua(%s)\n", ruleId, campaignId, timeStamp, ip, string(ua))
		return
	}

	err = DBSaveFraudFilterLogDetail(logId, hlog)
	if err != nil {
		log.Errorf("[recordFFHitRecords] save log detail fail ruleId(%d)-campaignId(%d)-timeStamp(%d)-ip(%s)-ua(%s)\n", ruleId, campaignId, timeStamp, ip, string(ua))
	}
	log.Debugf("[recordFFHitRecords]ruleId(%d)-campaignId(%d)-timeStamp(%d)-ip(%s)-ua(%s)\n",
		ruleId, campaignId, timeStamp, ip, string(ua))
}

// 将所有缓存的命中记录，都同步到数据库中
func flushFFHitRecords() error {
	//TODO 使用缓存提高性能
	log.Debugf("[flushFFHitRecords]")
	return nil
}

type userBotBlacklist struct {
	name      string
	userId    int64
	ipRange   []string
	userAgent []string
}

// 根据缓存的命中记录，更新用户的BotBlacklist
func updateUserBlacklist(userList []userBotBlacklist) (err error) {
	//TODO 使用缓存提高性能
	if len(userList) == 0 {
		return nil
	}
	us := []string{}
	for _, u := range userList {
		us = append(us, fmt.Sprintf("%d", u.userId))
		err = blacklist.DBInsertUserBlacklist(u.name, u.userId, u.ipRange, u.userAgent)
		if err != nil {
			log.Errorf("[flushFFHitRecords] fail: %v", err)
		}
	}
	err = PublishMsg(strings.Join(common.RemoveDuplicates(us), ","))
	if err != nil {
		log.Errorf("[flushFFHitRecords] fail: %v", err)
	}
	log.Debugf("[updateUserBlacklist]")
	return nil
}

const (
	ModeProducer = "producer"
	ModeConsumer = "consumer"
)

var kvSyncChan = make(chan kv, 1000000)
var started = false

func Start(mode string) (err error) {
	if started {
		return errors.New("already started")
	}
	switch mode {
	case ModeProducer:
		go requestsProducing(kvSyncChan)
		started = true
	case ModeConsumer:
		ffInterval = config.Int("FFRule", "interval")
		go requestHandling()
		started = true
	default:
		return fmt.Errorf("unsupported mode:%s", mode)
	}

	return nil
}
