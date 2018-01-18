package flow

import (
	"Service/common"
	"Service/db"
	"Service/log"
	"Service/units/rule"
	"database/sql"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

//func DBGetAllFlows() []FlowConfig {
//	return nil
//}

// no cache
func dbGetAvailableFlows() []FlowConfig {
	d := dbgetter()
	sql := "SELECT id, userId, name, redirectMode FROM Flow WHERE deleted=0"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[flow][dbGetAvailableFlows]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var arr []FlowConfig
	for rows.Next() {
		var c FlowConfig
		if err := rows.Scan(&c.Id, &c.UserId, &c.Name, &c.RedirectMode); err != nil {
			log.Errorf("[lander][dbGetAvailableFlows] scan failed:%v", err)
			return nil
		}
		arr = append(arr, c)
	}
	return arr
}

//func DBGetUserFlows(userId int64) []FlowConfig {
//	d := dbgetter()
//	sql := "SELECT id, userId, name, redirectMode FROM Flow WHERE userId=? AND deleted=0"
//	rows, err := d.Query(sql, userId)
//	if err != nil {
//		log.Errorf("[flow][DBGetUserFlows]Query: %s with userId:%v failed:%v", sql, userId, err)
//		return nil
//	}
//	defer rows.Close()

//	var c FlowConfig
//	var arr []FlowConfig
//	for rows.Next() {
//		if err := rows.Scan(&c.Id, &c.UserId, &c.Name, &c.RedirectMode); err != nil {
//			log.Errorf("[flow][DBGetUserFlows] scan failed:%v", err)
//			return nil
//		}
//		arr = append(arr, c)
//	}
//	return arr
//}

// with cache
func DBGetFlow(flowId int64) (c FlowConfig) {
	if flowDBCache.Enabled() {
		v := flowDBCache.Get(flowId)
		if v == nil {
			log.Errorf("[DBGetFlow]Flow(%d) does not exist in flowDBCache", flowId)
			return
		}
		return v.(FlowConfig)
	}

	d := dbgetter()
	sql := "SELECT id, userId, name, redirectMode FROM Flow WHERE id=?"
	row := d.QueryRow(sql, flowId)

	if err := row.Scan(&c.Id, &c.UserId, &c.Name, &c.RedirectMode); err != nil {
		log.Errorf("[lander][DBGetFlow] scan failed:%v", err)
		return c
	}
	return c
}

// no cache
func dbGetAllFlowRuleIds() (ruleIds map[int64][]FlowRule) {
	d := dbgetter()
	sql := "SELECT flowId, ruleId, status FROM Rule2Flow WHERE deleted=0 ORDER BY `order` DESC"

	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[flow][dbGetAllFlowRuleIds]Query: %s failed:%v", sql, err)
		return
	}
	defer rows.Close()

	ruleIds = make(map[int64][]FlowRule)
	var c FlowRule
	var flowId int64
	for rows.Next() {
		if err := rows.Scan(&flowId, &c.RuleId, &c.Status); err != nil {
			log.Errorf("[flow][dbGetAllFlowRuleIds] scan failed:%v", err)
			return
		}
		ruleIds[flowId] = append(ruleIds[flowId], c)
	}
	return
}

// with cache
func DBGetFlowRuleIds(flowId int64) (defaultRuleId FlowRule, ruleIds []FlowRule) {
	if flowRulesDBCache.Enabled() {
		v := flowRulesDBCache.Get(flowId)
		if v == nil {
			log.Errorf("flowRules data not exist in flowRulesDBCache for id(%d)", flowId)
			return
		}
		ruleIds = make([]FlowRule, len(v))
		for i, _ := range v {
			ruleIds[i] = v[i].(FlowRule)
		}
	} else {
		d := dbgetter()
		sql := "SELECT ruleId, status FROM Rule2Flow WHERE flowId=? AND deleted=0 ORDER BY `order` DESC"

		rows, err := d.Query(sql, flowId)
		if err != nil {
			log.Errorf("[flow][DBGetFlowRuleIds]Query: %s with id:%v failed:%v", sql, flowId, err)
			return
		}
		defer rows.Close()

		var c FlowRule
		for rows.Next() {
			if err := rows.Scan(&c.RuleId, &c.Status); err != nil {
				log.Errorf("[flow][DBGetFlowRuleIds] scan failed:%v", err)
				return
			}
			ruleIds = append(ruleIds, c)
		}
	}

	// 从中找出哪个是default的
	for idx, fr := range ruleIds {
		r := rule.DBGetRule(fr.RuleId)
		if r.Type == 0 {
			defaultRuleId = fr
			ruleIds = append(ruleIds[0:idx], ruleIds[idx+1:]...)
			break
		}
	}
	return
}

var flowRulesDBCache common.DBSliceCache

func enableFlowRulesDBCache() {
	origin := dbGetAllFlowRuleIds()
	data := make(map[int64][]common.HasID, len(origin))
	for flowId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[flowId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[flowId][i] = s[i]
		}
	}
	flowRulesDBCache.Init(data)
}
func disableFlowRulesDBCache() {
	flowRulesDBCache.Clear()
}

var flowDBCache common.DBCache

func EnableFlowDBCache() {
	log.Info("EnableFlowDBCache Begin")
	rule.EnableRuleDBCache()
	enableFlowRulesDBCache()

	origin := dbGetAvailableFlows()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	flowDBCache.Init(data)
	log.Info("EnableFlowDBCache End")
}
func DisableFlowDBCache() {
	flowDBCache.Clear()
	disableFlowRulesDBCache()
	rule.DisableRuleDBCache()
	log.Info("DisableFlowDBCache")
}
