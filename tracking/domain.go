package tracking

import (
	"Service/gracequit"
	"database/sql"
	"time"
)

// AdReferrerDomainStatis表的支持工作
// 使用方式：tracking.Domain.AddClick(k, 1)

// ReferrerDomainStatisKey AdReferrerDomainStatis表里面的Unique Key部分
type ReferrerDomainStatisKey struct {
	UserID         int64
	Timestamp      int64
	CampaignID     int64
	ReferrerDomain string
}

var referrerDomainStatisSQL = `INSERT INTO AdReferrerDomainStatis
(UserID,
Timestamp,
CampaignID,
ReferrerDomain,

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

// Domain 默认的AdIPStatis汇总存储
var Domain gatherSaver

// InitDomainGatherSaver 初始化tracking.Domain
func InitDomainGatherSaver(g *gracequit.GraceQuit, db *sql.DB, saveInterval time.Duration) {
	Domain = newGatherSaver(g, referrerDomainStatisSQL, saveInterval)
	Domain.Start(db)
}
