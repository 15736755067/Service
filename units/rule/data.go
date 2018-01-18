package rule

import (
	"Service/common"
	"Service/db"
	"Service/log"
	"Service/units/path"
	"database/sql"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

// 都是获取deleted=0的记录

// no cache
func dbGetAvailableRules() []RuleConfig {
	d := dbgetter()
	sql := "SELECT id, userId, json, status, `type` FROM Rule WHERE deleted=0"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[rule][dbGetAvailableRules]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var c RuleConfig
	var arr []RuleConfig
	for rows.Next() {
		if err := rows.Scan(&c.Id, &c.UserId, &c.Json, &c.Status, &c.Type); err != nil {
			log.Errorf("[rule][dbGetAvailableRules] scan failed:%v", err)
			return nil
		}
		arr = append(arr, c)
	}
	return arr
}

//func DBGetUserRules(userId int64) []RuleConfig {
//	d := dbgetter()
//	sql := "SELECT id, userId, json, status, `type` FROM Rule WHERE userId=? AND deleted=0"
//	rows, err := d.Query(sql, userId)
//	if err != nil {
//		log.Errorf("[rule][DBGetUserRules]Query: %s failed:%v", sql, err)
//		return nil
//	}
//	defer rows.Close()

//	var c RuleConfig
//	var arr []RuleConfig
//	for rows.Next() {
//		if err := rows.Scan(&c.Id, &c.UserId, &c.Json, &c.Status, &c.Type); err != nil {
//			log.Errorf("[rule][DBGetUserRules] scan failed:%v", err)
//			return nil
//		}
//		arr = append(arr, c)
//	}
//	return arr
//}

// with cache
func DBGetRule(ruleId int64) (c RuleConfig) {
	if ruleDBCache.Enabled() {
		v := ruleDBCache.Get(ruleId)
		if v == nil {
			log.Errorf("[DBGetRule]Rule(%d) does not exist in ruleDBCache", ruleId)
			return
		}
		return v.(RuleConfig)
	}

	d := dbgetter()
	sql := "SELECT id, userId, json, status, `type` FROM Rule WHERE id=?"
	row := d.QueryRow(sql, ruleId)

	if err := row.Scan(&c.Id, &c.UserId, &c.Json, &c.Status, &c.Type); err != nil {
		log.Errorf("[rule][DBGetRule] ruleId:%v scan failed:%v", ruleId, err)
		return
	}
	return
}

// with cache
func DBGetRulePaths(ruleId int64) (paths []RulePath) {
	if rulePathDBCache.Enabled() {
		v := rulePathDBCache.Get(ruleId)
		if v == nil {
			log.Errorf("rulePaths data not exist in rulePathDBCache for id(%d)", ruleId)
			return
		}
		paths = make([]RulePath, len(v))
		for i, _ := range v {
			paths[i] = v[i].(RulePath)
		}
		return
	}
	d := dbgetter()
	sql := "SELECT pathId, weight, status FROM Path2Rule WHERE ruleId = ? AND deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql, ruleId)
	if err != nil {
		log.Errorf("[rule][DBGetRulePaths]Query sql:%v failed:%v", sql, err)
		return
	}
	defer rows.Close()

	var rp RulePath
	for rows.Next() {
		err = rows.Scan(&rp.PathId, &rp.Weight, &rp.Status)
		if err != nil {
			log.Errorf("[rule][DBGetRulePaths]Scan RulePath failed:%v", err)
			return nil
		}
		paths = append(paths, rp)
	}
	return paths
}

//no cache
func dbGetAvailableRulePaths() (rulePaths map[int64][]RulePath) {
	d := dbgetter()
	sql := "SELECT ruleId, pathId, weight, status FROM Path2Rule WHERE deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[rule][DBGetRulePaths]Query sql:%v failed:%v", sql, err)
		return
	}
	defer rows.Close()

	rulePaths = make(map[int64][]RulePath)
	var rp RulePath
	var ruleId int64
	for rows.Next() {
		err = rows.Scan(&ruleId, &rp.PathId, &rp.Weight, &rp.Status)
		if err != nil {
			log.Errorf("[rule][DBGetRulePaths]Scan RulePath failed:%v", err)
			return nil
		}
		rulePaths[ruleId] = append(rulePaths[ruleId], rp)
	}
	return rulePaths
}

var ruleDBCache common.DBCache

func EnableRuleDBCache() {
	log.Info("EnableRuleDBCache Begin")
	path.EnablePathDBCache()
	enableRulePathDBCache()

	origin := dbGetAvailableRules()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	ruleDBCache.Init(data)
	log.Info("EnableRuleDBCache End")
}
func DisableRuleDBCache() {
	ruleDBCache.Clear()
	disableRulePathDBCache()
	path.DisablePathDBCache()
	log.Info("DisableRuleDBCache")
}

var rulePathDBCache common.DBSliceCache

func enableRulePathDBCache() {
	origin := dbGetAvailableRulePaths()
	data := make(map[int64][]common.HasID, len(origin))
	for ruleId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[ruleId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[ruleId][i] = s[i]
		}
	}
	rulePathDBCache.Init(data)
}
func disableRulePathDBCache() {
	rulePathDBCache.Clear()
}
