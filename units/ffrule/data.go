package ffrule

import (
	"database/sql"
	"fmt"

	"Service/common"
	"Service/db"
	"Service/log"
	"encoding/json"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

//no cache
func dbGetAvailableRules() (rules []RuleConfig) {
	d := dbgetter()
	sql := `SELECT id,name,status,userId,dimension,timeSpan,FraudFilterRule.condition FROM FraudFilterRule WHERE deleted=0`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[ffrule][dbGetAvailableRules]Query: %s failed:%v", sql, err)
		return
	}

	var status int
	for rows.Next() {
		var rule RuleConfig
		if err := rows.Scan(&rule.Id, &rule.Name, &status, &rule.UserId, &rule.Dimension, &rule.TimeSpan, &rule.Conditions); err != nil {
			log.Errorf("[ffrule][dbGetAvailableRules]Query: %s failed:%v", sql, err)
			rows.Close()
			return
		}
		rule.Active = (status == 1)
		rules = append(rules, rule)
	}
	rows.Close()

	for i, _ := range rules {
		rules[i].CampaignIds = DBGetRuleCampaigns(rules[i].Id)
	}

	return
}

//no cache
func dbAvailableRuleCampaigns() (ruleCampaigns map[int64][]int64) {
	d := dbgetter()
	sql := `SELECT ruleId, campaignId FROM FFRule2Campaign`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[ffrule][dbAvailableRuleCampaigns]Query: %s failed:%v", sql, err)
		return
	}
	defer rows.Close()

	ruleCampaigns = make(map[int64][]int64)
	var ruleId, campaignId int64
	for rows.Next() {
		if err := rows.Scan(&ruleId, &campaignId); err != nil {
			log.Errorf("[ffrule][dbAvailableRuleCampaigns] scan failed:%v", err)
			return
		}
		if ruleId == 0 || campaignId == 0 {
			continue
		}
		ruleCampaigns[ruleId] = append(ruleCampaigns[ruleId], campaignId)
	}
	return
}

//with cache
func DBGetRule(ruleId int64) (rule RuleConfig) {
	if ruleId <= 0 {
		return
	}

	if ruleDBCache.Enabled() {
		v := ruleDBCache.Get(ruleId)
		if v == nil {
			log.Errorf("rule data not exist in ruleDBCache for id(%d)", ruleId)
			return
		}
		rule = v.(RuleConfig).GetCopy()
		return
	}

	d := dbgetter()
	var status int
	sql := `SELECT id,name,status,userId,dimension,timeSpan,FraudFilterRule.condition FROM FraudFilterRule WHERE id=? AND deleted=0`
	if err := d.QueryRow(sql, ruleId).
		Scan(&rule.Id, &rule.Name, &status, &rule.UserId, &rule.Dimension, &rule.TimeSpan, &rule.Conditions); err != nil {
		log.Errorf("[ffrule][DBGetRule]QueryRow: %s with id:%v failed:%v", sql, ruleId, err)
	}
	rule.Active = (status == 1)
	rule.CampaignIds = DBGetRuleCampaigns(rule.Id)
	return
}

//no cache
func dbGetRuleCampaigns() (ruleCampaignIds map[int64][]int64) {
	d := dbgetter()
	sql := `SELECT ruleId,campaignId FROM FFRule2Campaign`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[ffrule][DBGetRuleCampaigns]Query: %s with failed:%v", sql, err)
		return
	}
	defer rows.Close()

	ruleCampaignIds = make(map[int64][]int64)
	var ruleId, campaignId int64
	for rows.Next() {
		if err := rows.Scan(&ruleId, &campaignId); err != nil {
			log.Errorf("[ffrule][DBGetRuleCampaigns] scan failed:%v", err)
			return
		}
		ruleCampaignIds[ruleId] = append(ruleCampaignIds[ruleId], campaignId)
	}
	return
}

//with cache
func DBGetRuleCampaigns(ruleId int64) (campaignIds []int64) {
	if ruleId <= 0 {
		return
	}

	if ruleCampaignsDBCache.Enabled() {
		v := ruleCampaignsDBCache.Get(ruleId)
		if v == nil {
			log.Errorf("campaignIds data not exist in ruleCampaignsDBCache for id(%d)", ruleId)
			return
		}
		campaignIds = make([]int64, len(v))
		for i, _ := range v {
			campaignIds[i] = int64(v[i].(hasid))
		}
		return
	}

	d := dbgetter()
	sql := `SELECT campaignId FROM FFRule2Campaign WHERE ruleId=?`
	rows, err := d.Query(sql, ruleId)
	if err != nil {
		log.Errorf("[ffrule][DBGetRuleCampaigns]Query: %s with id:%v failed:%v", sql, ruleId, err)
		return
	}
	defer rows.Close()

	var id int64
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Errorf("[ffrule][DBGetRuleCampaigns] scan failed:%v", err)
			return
		}
		campaignIds = append(campaignIds, id)
	}
	return
}

//no cache
func dbGetAllCampaignAvailableFFRuleIds() (campaignRuleIds map[int64][]int64) {
	d := dbgetter()
	sql := `SELECT campaignId, ruleId FROM FFRule2Campaign, FraudFilterRule 
            WHERE FraudFilterRule.id=FFRule2Campaign.ruleId 
            AND FraudFilterRule.deleted=0 AND FraudFilterRule.status=1`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[ffrule][DBGetCampaignFFRuleIds]Query: %s with failed:%v", sql, err)
		return
	}
	defer rows.Close()

	campaignRuleIds = make(map[int64][]int64)
	var campaignId, ruleId int64
	for rows.Next() {
		if err := rows.Scan(&campaignId, &ruleId); err != nil {
			log.Errorf("[ffrule][DBGetCampaignFFRuleIds]scan failed:%v", err)
			return
		}
		campaignRuleIds[campaignId] = append(campaignRuleIds[campaignId], ruleId)
	}
	return
}

//with cache
func DBGetCampaignAvailableFFRuleIds(campaignId int64) (rules []int64) {
	if campaignId <= 0 {
		return
	}

	if campaignRuleIdsDBCache.Enabled() {
		v := campaignRuleIdsDBCache.Get(campaignId)
		if v == nil {
			log.Errorf("ruleIds data not exist in campaignRuleIdsDBCache for id(%d)", campaignId)
			return
		}
		rules = make([]int64, len(v))
		for i, _ := range v {
			rules[i] = int64(v[i].(hasid))
		}
		return
	}

	d := dbgetter()
	sql := `SELECT ruleId FROM FFRule2Campaign, FraudFilterRule 
            WHERE FFRule2Campaign.campaignId=? AND FraudFilterRule.id=FFRule2Campaign.ruleId 
            AND FraudFilterRule.deleted=0 AND FraudFilterRule.status=1`
	rows, err := d.Query(sql, campaignId)
	if err != nil {
		log.Errorf("[ffrule][DBGetCampaignFFRuleIds]Query: %s with id:%v failed:%v", sql, campaignId, err)
		return
	}
	defer rows.Close()

	var ruleId int64
	for rows.Next() {
		if err := rows.Scan(&ruleId); err != nil {
			log.Errorf("[ffrule][DBGetCampaignFFRuleIds]scan failed:%v", err)
			return
		}
		rules = append(rules, ruleId)
	}
	return
}

var ruleDBCache common.DBCache

func EnableRuleDBCache() {
	log.Info("EnableFFRuleDBCache Begin")
	enableCampaignRuleIdsDBCache()
	enableRuleCampaignsDBCache()

	origin := dbGetAvailableRules()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	ruleDBCache.Init(data)
	log.Info("EnableFFRuleDBCache End")
}

func DisableRuleDBCache() {
	ruleDBCache.Clear()
	disableRuleCampaignsDBCache()
	disableCampaignRuleIdsDBCache()
	log.Info("DisableFFRuleDBCache")
}

var ruleCampaignsDBCache common.DBSliceCache

type hasid int64

func (h hasid) ID() int64 {
	return int64(h)
}

func enableRuleCampaignsDBCache() {
	origin := dbGetRuleCampaigns()
	data := make(map[int64][]common.HasID, len(origin))
	for ruleId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[ruleId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[ruleId][i] = hasid(s[i])
		}
	}
	ruleCampaignsDBCache.Init(data)
}
func disableRuleCampaignsDBCache() {
	ruleCampaignsDBCache.Clear()
}

var campaignRuleIdsDBCache common.DBSliceCache

func enableCampaignRuleIdsDBCache() {
	origin := dbGetAllCampaignAvailableFFRuleIds()
	data := make(map[int64][]common.HasID, len(origin))
	for campaignId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[campaignId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[campaignId][i] = hasid(s[i])
		}
	}
	campaignRuleIdsDBCache.Init(data)
}
func disableCampaignRuleIdsDBCache() {
	campaignRuleIdsDBCache.Clear()
}

func DBSaveFFHitRecords(l hitLog) (int64, error) {
	d := dbgetter()
	sqlStr := "INSERT INTO FraudFilterLog(`ruleId`,`dimension`,`hit`,`condition`,`timeStamp`) VALUES(?,?,?,?,?)"
	r := GetRule(l.RuleId)

	condition := ""
	for i, c := range r.Conditions {
		if i == 0 {
			condition += fmt.Sprintf("%s%s%d", c.Key, c.Op, c.Value)
		} else {
			condition += fmt.Sprintf(",%s%s%d", c.Key, c.Op, c.Value)
		}
	}
	res, err := d.Exec(sqlStr, l.RuleId, r.Dimension, 1, condition, l.TimeStamp)
	if err != nil {
		log.Errorf("[ffrule][DBSaveFFHitRecords]Exec: %s with id:%v failed:%v", sqlStr, l.CampaignId, err)
		return 0, err
	}
	logId, _ := res.LastInsertId()
	return logId, nil
}

func DBSaveFraudFilterLogDetail(logId int64, l hitLog) error {
	d := dbgetter()
	sqlStr := "INSERT INTO FraudFilterLogDetail(`logId`,`campaignID`,`data`) VALUES(?,?,?)"
	bs, err := json.Marshal(l.LogData)
	if err != nil {
		log.Errorf("[ffrule][DBSaveFraudFilterLogDetail]Marshal: %+v with id:%v logId:%v failed:%v", l.LogData, l.CampaignId, logId, err)
		return err
	}

	_, err = d.Exec(sqlStr, logId, l.CampaignId, string(bs))
	if err != nil {
		log.Errorf("[ffrule][DBSaveFraudFilterLogDetail]Exec: %s with id:%v logId:%v failed:%v", sqlStr, l.CampaignId, logId, err)
		return err
	}
	return nil
}

func PublishMsg(msg string) error {
	var msgChannel = "channel_blacklist_changed_users"
	redis := db.GetRedisClient("MSGQUEUE")
	pubSub := redis.Publish(msgChannel, msg)
	if err := pubSub.Err(); err != nil {
		log.Errorf("[ffrule][PublishMsg]Publish: %s failed:%v", msg, err)
		return err
	}
	return nil
}
