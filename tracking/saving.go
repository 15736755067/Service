package tracking

import (
	"Service/log"
	"database/sql"
	"sync"
	"time"
)

// Saving 一直执行保存操作
// db参数暂时传过来
func Saving(db *sql.DB, stop chan struct{}) {
	for {
		select {
		case m := <-toSave:
			err := doSave(db, m)
			if err != nil {
				log.Errorf("doSave failed:%v", err)
			}
		case <-stop:
			// 收所有的数据，防止的未写入数据库的
			for {
				select {
				case m := <-toSave:
					err := doSave(db, m)
					if err != nil {
						log.Errorf("doSave failed:%v", err)
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

func doSave(db *sql.DB, m map[string]*adStaticTableFields) error {
	// 这样可以Prepare一下，存储更加快
	start := time.Now()
	defer func() {
		log.Infof("[tracking][doSave] save %d records take: %v", len(m), time.Now().Sub(start))
	}()

	if len(m) == 0 {
		// 如果没有数据需要存储，直接退出
		return nil
	}

	// 提交Prepare可以避免重复解析SQL语句
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Errorf("[tracking][doSave] Prepare[%s] failed:%v", insertSQL, err)
		return err
	}
	defer stmt.Close()

	userEventsCount := make(map[int64]int64)

	w := sync.WaitGroup{}
	for keyMD5, fields := range m {
		w.Add(1)
		go func(keyMD5 string, fields *adStaticTableFields) {
			defer func() {
				w.Done()
				if x := recover(); x != nil {
					log.Error("[tracking][doSave] insert panic:", x)
				}
			}()
			_, err := stmt.Exec(
				fields.UserID,
				fields.CampaignID,
				fields.CampaignName,
				fields.FlowID,
				fields.FlowName,
				fields.LanderID,
				fields.LanderName,
				fields.OfferID,
				fields.OfferName,
				fields.AffiliateNetworkID,
				fields.AffiliateNetworkName,
				fields.TrafficSourceID,
				fields.TrafficSourceName,
				fields.Language,
				fields.Model,
				fields.Country,
				fields.City,
				fields.Region,
				fields.ISP,
				fields.MobileCarrier,
				fields.Domain,
				fields.DeviceType,
				fields.Brand,
				fields.OS,
				fields.OSVersion,
				fields.Browser,
				fields.BrowserVersion,
				fields.ConnectionType,
				fields.Timestamp,
				fields.Visits,
				fields.Clicks,
				fields.Conversions,
				fields.Cost,
				fields.Revenue,
				fields.Impressions,
				keyMD5,
				fields.V1,
				fields.V2,
				fields.V3,
				fields.V4,
				fields.V5,
				fields.V6,
				fields.V7,
				fields.V8,
				fields.V9,
				fields.V10,
				fields.TSCampaignId,
				fields.TSWebsiteId,

				fields.Visits,
				fields.Clicks,
				fields.Conversions,
				fields.Cost,
				fields.Revenue,
				fields.Impressions,
			)

			if err != nil {
				log.Error("[tracking][doSave] insert err:", err.Error(), "with field:", *fields)
			}
		}(keyMD5, fields)
		userEventsCount[fields.UserID] = userEventsCount[fields.UserID] + int64(fields.Impressions+fields.Visits+fields.Clicks+fields.Conversions)
	}
	w.Wait()

	selectMaxId := `SELECT max(id) FROM UserBilling WHERE userId=?`
	maxIdStmt, err := db.Prepare(selectMaxId)
	if err != nil {
		log.Errorf("[tracking][doSave] Prepare[%s] failed:%v", selectMaxId, err)
		return err
	}
	defer maxIdStmt.Close()

	updateSql := `UPDATE UserBilling SET totalEvents = totalEvents + ? where id = ?`
	updateStmt, err := db.Prepare(updateSql)
	if err != nil {
		log.Errorf("[tracking][doSave] Prepare[%s] failed:%v", updateSql, err)
		return err
	}
	defer updateStmt.Close()

	for user, count := range userEventsCount {
		w.Add(1)
		go func(db *sql.DB, user, count int64, maxIdStmt, updateStmt *sql.Stmt) {
			defer func() {
				w.Done()
				if x := recover(); x != nil {
					log.Error("[tracking][doSave] updateUserEvents panic:", x)
				}
			}()
			updateUserEvents(db, user, count, maxIdStmt, updateStmt)
		}(db, user, count, maxIdStmt, updateStmt)
	}
	w.Wait()

	return nil
}

func updateUserEvents(db *sql.DB, user, count int64, maxId *sql.Stmt, update *sql.Stmt) {
	// db.QueryRow(`SELECT max(id) FROM UserBilling WHERE userId=?`)
	row := maxId.QueryRow(user)
	var m sql.NullInt64
	err := row.Scan(&m)
	if err != nil {
		log.Errorf("QueryRow FROM maxId Stmt failed: user:%v err:%v", user, err)
		return
	}

	if !m.Valid {
		log.Errorf("user:%v have no row in table UserBilling count:%v", user, count)
		return
	}

	_, err = update.Exec(count, m)
	if err != nil {
		log.Errorf("Update UserBilling failed user:%v m:%v err:%v", user, m, err)
		return
	}
}

var insertSQL = `INSERT INTO AdStatis
(
UserID,
CampaignID,
CampaignName,
FlowID,
FlowName,
LanderID,
LanderName,
OfferID,
OfferName,
AffiliateNetworkID,
AffilliateNetworkName,
TrafficSourceID,
TrafficSourceName,
Language,
Model,
Country,
City,
Region,
ISP,
MobileCarrier,
Domain,
DeviceType,
Brand,
OS,
OSVersion,
Browser,
BrowserVersion,
ConnectionType,
Timestamp,
Visits,
Clicks,
Conversions,
Cost,
Revenue,
Impressions,
KeysMD5,
V1,
V2,
V3,
V4,
V5,
V6,
V7,
V8,
V9,
V10,
tsCampaignId,
tsWebsiteId)

VALUES (
    ?,?,?,?,?,?,?,?,?,?,
    ?,?,?,?,?,?,?,?,?,?,
    ?,?,?,?,?,?,?,?,?,?,
	?,?,?,?,?,?,?,?,?,?,
	?,?,?,?,?,?,?,?
)

ON DUPLICATE KEY UPDATE
Visits = Visits+?, 
Clicks = Clicks+?, 
Conversions = Conversions+?, 
Cost = Cost+?, 
Revenue = Revenue+?,
Impressions = Impressions+?
`
