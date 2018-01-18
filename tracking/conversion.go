package tracking

import (
	"Service/log"
	"database/sql"
	"sync"
)

// Conversion 对应数据库里面的AdConversionsStatis字段
type Conversion struct {
	UserID               int64
	PostbackTimestamp    int64
	VisitTimestamp       int64
	ExternalID           string
	ClickID              string
	TransactionID        string
	Revenue              float64
	Cost                 float64
	CampaignName         string
	CampaignID           int64
	LanderName           string
	LanderID             int64
	OfferName            string
	OfferID              int64
	Country              string
	CountryCode          string
	TrafficSourceName    string
	TrafficSourceID      int64
	AffiliateNetworkName string
	AffiliateNetworkID   int64
	Device               string
	OS                   string
	OSVersion            string
	Brand                string
	Model                string
	Browser              string
	BrowserVersion       string
	ISP                  string
	MobileCarrier        string
	ConnectionType       string
	VisitorIP            string
	VisitorReferrer      string
	V1                   string
	V2                   string
	V3                   string
	V4                   string
	V5                   string
	V6                   string
	V7                   string
	V8                   string
	V9                   string
	V10                  string
}

var conversionsChan chan *Conversion

func init() {
	// 理论上讲Conversion的数据量会比较小，用不了那么大
	conversionsChan = make(chan *Conversion, 1024)
}

// SaveConversion 异步存储一条Conversion
func SaveConversion(c *Conversion) {
	conversionsChan <- c
}

// SavingConversions 负责对Conversion的存储
func SavingConversions(db *sql.DB, stop chan struct{}) {
	var conversions []*Conversion

	// 把chan里面所有需要存储的Conversion都拿出来
	// 如果没有，则直接返回
	collect := func() {
		for {
			select {
			case conversion := <-conversionsChan:
				conversions = append(conversions, conversion)
			default:
				return
			}
		}
	}

	for {
		select {
		case conversion := <-conversionsChan:
			// 拿出一条，则拿出所有
			conversions = append(conversions, conversion)

			// 把剩下的一并拿出来，统一存储
			collect()
			saveConversions(db, conversions)

			// 保存完毕，清空列表
			conversions = conversions[0:0]
		case <-stop:
			// 如果仍然有新的数据，也先写一下数据库
			collect()
			saveConversions(db, conversions)
			// 保存完毕，清空列表
			conversions = conversions[0:0]
			return
		}
	}
}

func saveConversions(db *sql.DB, conversions []*Conversion) {
	if len(conversions) == 0 {
		return
	}

	stmt, err := db.Prepare(insertConversionSQL)
	if err != nil {
		log.Errorf("[tracking][SavingConversions] Prepare[%s] failed:%v", insertConversionSQL, err)
		return
	}
	defer stmt.Close()

	w := sync.WaitGroup{}
	for _, c := range conversions {
		w.Add(1)
		go func(c *Conversion) {
			defer func() {
				w.Done()
				if x := recover(); x != nil {
					log.Error("[tracking][saveConversions] insert panic:", x)
				}
			}()
			_, err := stmt.Exec(
				c.UserID,
				c.PostbackTimestamp,
				c.VisitTimestamp,
				c.ExternalID,
				c.ClickID,
				c.TransactionID,
				c.Revenue,
				c.Cost,
				c.CampaignName,
				c.CampaignID,
				c.LanderName,
				c.LanderID,
				c.OfferName,
				c.OfferID,
				c.Country,
				c.CountryCode,
				c.TrafficSourceName,
				c.TrafficSourceID,
				c.AffiliateNetworkName,
				c.AffiliateNetworkID,
				c.Device,
				c.OS,
				c.OSVersion,
				c.Brand,
				c.Model,
				c.Browser,
				c.BrowserVersion,
				c.ISP,
				c.MobileCarrier,
				c.ConnectionType,
				c.VisitorIP,
				c.VisitorReferrer,
				c.V1,
				c.V2,
				c.V3,
				c.V4,
				c.V5,
				c.V6,
				c.V7,
				c.V8,
				c.V9,
				c.V10,
			)

			if err != nil {
				log.Errorf("[tracking][SavingConversions] Insert failed:%v", err)
			}
		}(c)
	}
	w.Wait()
}

var insertConversionSQL = `INSERT INTO AdConversionsStatis
(
UserID,
PostbackTimestamp,
VisitTimestamp,
ExternalID,
ClickID,
TransactionID,
Revenue,
Cost,
CampaignName,
CampaignID,
LanderName,
LanderID,
OfferName,
OfferID,
Country,
CountryCode,
TrafficSourceName,
TrafficSourceID,
AffiliateNetworkName,
AffiliateNetworkID,
Device,
OS,
OSVersion,
Brand,
Model,
Browser,
BrowserVersion,
ISP,
MobileCarrier,
ConnectionType,
VisitorIP,
VisitorReferrer,
V1,
V2,
V3,
V4,
V5,
V6,
V7,
V8,
V9,
V10)
VALUES
(
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?,
?)
`
