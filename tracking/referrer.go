package tracking

import (
	"Service/gracequit"
	"database/sql"
	"time"
)

// AdReferrerStatis表的支持工作
// 使用方式：tracking.Ref.AddClick(k, 1)

// ReferrerStatisKey AdReferrerStatis表里面的Unique Key部分
type ReferrerStatisKey struct {
	UserID     int64
	Timestamp  int64
	CampaignID int64
	Referrer   string
}

var referrerStatisSQL = `INSERT INTO AdReferrerStatis
(UserID,
Timestamp,
CampaignID,
Referrer,

Visits,
Clicks,
Conversions,
Cost,
Revenue,
Impressions)
VALUES
(?,?,?,?,?,?,?,?,?,?)
ON DUPLICATE KEY UPDATE
Visits = Visits+?,
Clicks = Clicks+?,
Conversions = Conversions+?,
Cost = Cost+?,
Revenue = Revenue+?,
Impressions = Impressions+?`

// Ref 默认的AdIPStatis汇总存储
var Ref gatherSaver

// InitRefGatherSaver 初始化tracking.Ref
func InitRefGatherSaver(g *gracequit.GraceQuit, db *sql.DB, saveInterval time.Duration) {
	Ref = newGatherSaver(g, referrerStatisSQL, saveInterval)
	Ref.Start(db)
}
