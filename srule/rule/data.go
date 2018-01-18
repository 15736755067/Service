package rule

import (
	"Service/config"
	"Service/log"
	"Service/util/timezone"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mysqlDb *gorm.DB // mysql
var mcsDb *gorm.DB   //mariadb columnstore

type campaign struct {
	id   int64
	name string
}

const defaultMaxOpenConns = 100
const defaultMaxIdleConns = 100

var campaignCacheMux sync.RWMutex
var campaignCache map[int64]*campaign = make(map[int64]*campaign)

func InitDB() error {
	if getMysqlDB() == nil || getColumnStoreDB() == nil {
		return errors.New("init db fails")
	}
	return nil
}

func getMysqlDB() *gorm.DB {
	if mysqlDb != nil && mysqlDb.DB().Ping() == nil {
		return mysqlDb
	}
	var err error
	host := config.String("DB", "host")
	user := config.String("DB", "user")
	pass := config.String("DB", "pass")
	port := config.Int("DB", "port")
	dbname := config.String("DB", "dbname")
	jdbcConnUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, port, dbname)

	db, err := gorm.Open("mysql", jdbcConnUrl)
	if err != nil {
		log.Errorf("mysql db error: %v", err)
		return nil
	}
	if config.Bool("DB", "logmode") {
		db.LogMode(true)
	}

	maxopen := config.Int("DB", "max_open_conns")
	maxidle := config.Int("DB", "max_idle_conns")

	if maxopen <= 0 {
		maxopen = defaultMaxOpenConns
	}
	db.DB().SetMaxOpenConns(maxopen)
	if maxidle <= 0 {
		maxidle = defaultMaxIdleConns
	}
	db.DB().SetMaxIdleConns(maxidle)

	if err := db.DB().Ping(); err == nil {
		mysqlDb = db
		return db
	}
	return nil
}

func getColumnStoreDB() *gorm.DB {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	if mcsDb != nil && mcsDb.DB().Ping() == nil {
		return mcsDb
	}
	host := config.String("ColumnStore", "host")
	user := config.String("ColumnStore", "user")
	pass := config.String("ColumnStore", "pass")
	port := config.Int("ColumnStore", "port")
	dbname := config.String("ColumnStore", "dbname")
	jdbcConnUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, pass, host, port, dbname)

	db, err := gorm.Open("mysql", jdbcConnUrl)
	if err != nil {
		log.Errorf("ColumnStore db error: %v", err)
		return nil
	}

	maxopen := config.Int("ColumnStore", "max_open_conns")
	maxidle := config.Int("ColumnStore", "max_idle_conns")

	if maxopen <= 0 {
		maxopen = defaultMaxOpenConns
	}
	db.DB().SetMaxOpenConns(maxopen)
	if maxidle <= 0 {
		maxidle = defaultMaxIdleConns
	}
	db.DB().SetMaxIdleConns(maxidle)

	if config.Bool("ColumnStore", "logmode") {
		db.LogMode(true)
	}
	if err := db.DB().Ping(); err == nil {
		mcsDb = db
		return db
	}
	return nil
}

func GetOneTimeList() []string {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	oneTimeList := make([]string, 0)

	db := getMysqlDB()
	whereStmt := "Status =1 and deleted = 0"
	rows, err := db.Table("SuddenChangeRule").Select("`oneTime`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return oneTimeList
	}
	defer rows.Close()

	for rows.Next() {
		var oneTimeStr string
		rows.Scan(&oneTimeStr)

		oneTime := GetTime(oneTimeStr)
		if oneTime.After(time.Now().UTC()) {
			oneTimeList = append(oneTimeList, oneTimeStr)
		}

	}

	return oneTimeList
}

func GetScheduleList() []string {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	scheduleList := make([]string, 0)

	db := getMysqlDB()
	whereStmt := "Status =1 and deleted = 0 "
	rows, err := db.Table("SuddenChangeRule").Select("`schedule`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return scheduleList
	}
	defer rows.Close()

	for rows.Next() {
		var schedule string
		rows.Scan(&schedule)
		scheduleList = append(scheduleList, schedule)
	}

	return scheduleList
}

func GetRulesBySchedule(schedule string) []*Rule {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()
	rules := make([]*Rule, 0)

	whereStmt := "Status =1 and deleted = 0 "
	whereStmt = whereStmt + " and " + fmt.Sprintf("Schedule = '%s'", schedule)

	rows, err := db.Table("SuddenChangeRule").Select("`id`,`userId`,`name`,`dimension`,`timeSpan`,`condition`,`emails`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return rules
	}
	defer rows.Close()

	for rows.Next() {
		var rule *Rule = &Rule{}
		var emails sql.NullString
		err := rows.Scan(&rule.ID, &rule.UserId, &rule.Name, &rule.Dimension, &rule.TimeSpan, &rule.Condition, &emails)
		if err != nil {
			log.Infof("scan fail:%v", err)
		}
		if emails.Valid {
			rule.Email = emails.String
		}
		if strings.HasSuffix(rule.Condition, ",") {
			rule.Condition = rule.Condition[:len(rule.Condition)-1]
		}
		rules = append(rules, rule)
	}

	//
	for _, rule := range rules {
		rule.Location = getTimeLocation(rule.UserId)
		rule.CampaignID = getCampaignIds(rule.ID)
		updateCampaigns(strings.Split(rule.CampaignID, ","))
	}

	return rules
}

func GetRulesByOneTime(oneTime string) []*Rule {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()
	rules := make([]*Rule, 0)

	whereStmt := "Status =1 and deleted = 0"
	whereStmt = whereStmt + " and " + fmt.Sprintf("oneTime = '%s'", oneTime)

	rows, err := db.Table("SuddenChangeRule").Select("`id`,`userId`,`name`,`dimension`,`timeSpan`,`condition`,`emails`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return rules
	}
	defer rows.Close()

	for rows.Next() {
		var rule *Rule = &Rule{}
		var emails sql.NullString
		err := rows.Scan(&rule.ID, &rule.UserId, &rule.Name, &rule.Dimension, &rule.TimeSpan, &rule.Condition, &emails)
		if err != nil {
			log.Infof("scan fail:%v", err)
		}
		if emails.Valid {
			rule.Email = emails.String
		}

		if strings.HasSuffix(rule.Condition, ",") {
			rule.Condition = rule.Condition[:len(rule.Condition)-1]
		}

		//fmt.Printf("rule: %v\n", rule)
		rules = append(rules, rule)
	}

	//
	for _, rule := range rules {
		rule.Location = getTimeLocation(rule.UserId)
		rule.CampaignID = getCampaignIds(rule.ID)
		updateCampaigns(strings.Split(rule.CampaignID, ","))
	}

	return rules
}

func getCampaignIds(ruleId int64) string {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()
	whereStmt := fmt.Sprintf("ruleId = %d", ruleId)
	rows, err := db.Table("SCRule2Campaign").Select("`campaignId`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return ""
	}

	defer rows.Close()
	campaignIds := ""

	for rows.Next() {
		var campaignId int64
		rows.Scan(&campaignId)
		if campaignIds == "" {
			campaignIds = strconv.Itoa(int(campaignId))
		} else {
			campaignIds = campaignIds + "," + strconv.Itoa(int(campaignId))
		}
	}

	return campaignIds
}

func updateCampaigns(ids []string) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	db := getMysqlDB()
	idList := strings.Join(ids, ",")
	whereStmt := fmt.Sprintf("id in (%s)", idList)
	rows, err := db.Table("TrackingCampaign").Select("`id`,`name`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return
	}

	defer rows.Close()

	campaignCacheMux.Lock()
	defer campaignCacheMux.Unlock()
	for rows.Next() {
		var campaignId int64
		var campaignName string
		rows.Scan(&campaignId, &campaignName)

		campaignCache[campaignId] = &campaign{id: campaignId, name: campaignName}
	}
}

func getCampaign(campaignId int64) *campaign {
	campaignCacheMux.RLock()
	defer campaignCacheMux.RUnlock()
	if c, ok := campaignCache[campaignId]; ok {
		return c
	}
	return nil
}

func getTimeLocation(userId string) *time.Location {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
		}
	}()
	if userId == "" {
		return time.UTC
	}
	db := getMysqlDB()
	whereStmt := fmt.Sprintf("id = %s", userId)
	rows, err := db.Table("User").Select("`timezone`,`timezoneId`").Where(whereStmt).Rows()
	if err != nil {
		log.Infof("err: %v", err)
		return time.UTC
	}
	defer rows.Close()

	for rows.Next() {
		var timezoneOffset string
		var timezoneId int
		rows.Scan(&timezoneOffset, &timezoneId)
		location := timezone.TimeLocation(strconv.Itoa(timezoneId), timezoneOffset)
		return location
	}
	return time.UTC
}

func GetTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse("Mon Jan 2 15:04 2006", value)
	if err != nil {
		t, err = time.Parse("Mon Jan 2 15:04:05 2006", value)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", value)
		}

		if err != nil {
			t, err = time.Parse("2006-01-02T15:04", value)
		}

		if err != nil {
			t, err = time.Parse("2006-01-02T15", value)
		}

		if err != nil {
			t, err = time.Parse("2006-01-02 15", value)
		}

		if err != nil {
			return time.Time{}
		}
	}

	return t
}
