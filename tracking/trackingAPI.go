package tracking

import (
	"crypto/md5"
	"fmt"
	"time"
)

// AdStatisKey AdStatic表里面用来当作unique key的所有的字段
// Timestamp不用给，里面会赋值
type AdStatisKey struct {
	UserID               int64
	CampaignID           int64
	CampaignName         string
	FlowID               int64
	FlowName             string
	LanderID             int64
	LanderName           string
	OfferID              int64
	OfferName            string
	AffiliateNetworkID   int64
	AffiliateNetworkName string
	TrafficSourceID      int64
	TrafficSourceName    string
	Language             string
	Model                string
	Country              string
	City                 string
	Region               string
	ISP                  string
	MobileCarrier        string
	Domain               string
	DeviceType           string
	Brand                string
	OS                   string
	OSVersion            string
	Browser              string
	BrowserVersion       string
	ConnectionType       string
	Timestamp            int64
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

	TSCampaignId string
	TSWebsiteId string
}

func statisKeyMD5(k *AdStatisKey) string {
	h := md5.New()
	// io.WriteString(h, txt)
	fmt.Fprintf(h, "%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v",
		k.UserID,
		k.CampaignID,
		k.FlowID,
		k.LanderID,
		k.OfferID,
		k.AffiliateNetworkID,
		k.TrafficSourceID,
		k.Language,
		k.Model,
		k.Country,
		k.City,
		k.Region,
		k.ISP,
		k.MobileCarrier,
		k.Domain,
		k.DeviceType,
		k.Brand,
		k.OS,
		k.OSVersion,
		k.Browser,
		k.BrowserVersion,
		k.ConnectionType,
		k.Timestamp,
		k.V1,
		k.V2,
		k.V3,
		k.V4,
		k.V5,
		k.V6,
		k.V7,
		k.V8,
		k.V9,
		k.V10,
		k.TSCampaignId,
		k.TSWebsiteId,
	)
	return fmt.Sprintf("%X", h.Sum(nil))
}

// AddVisit 增加visit统计信息
func AddVisit(key AdStatisKey, count int) {
	f := func(d *adStatisValues) {
		d.Visits += count
	}
	addTrackEvent(key, f)
}

// AddClick 增加click统计信息
func AddClick(key AdStatisKey, count int) {
	f := func(d *adStatisValues) {
		d.Clicks += count
	}
	addTrackEvent(key, f)
}

// AddConversion 增加conversion统计信息
func AddConversion(key AdStatisKey, count int) {
	f := func(d *adStatisValues) {
		d.Conversions += count
	}
	addTrackEvent(key, f)
}

// AddImpression 增加impression统计信息
func AddImpression(key AdStatisKey, count int) {
	f := func(d *adStatisValues) {
		d.Impressions += count
	}
	addTrackEvent(key, f)
}

// AddCost 增加cost统计信息
func AddCost(key AdStatisKey, count float64) {
	f := func(d *adStatisValues) {
		d.Cost += int64(count * MILLION)
	}
	addTrackEvent(key, f)
}

// AddPayout 增加payout统计信息
func AddPayout(key AdStatisKey, count float64) {
	f := func(d *adStatisValues) {
		d.Revenue += int64(count * MILLION)
	}
	addTrackEvent(key, f)
}

// 统计信息的时间粒度
// 以前是一个小时，现在改成5分钟
const timestampGranularity = 3600000 / 12

func addTrackEvent(key AdStatisKey, action func(d *adStatisValues)) {
	gatherChan <- events{
		keyMd5:    statisKeyMD5(&key),
		keyFields: key,
		action:    action,
	}
}

// Timestamp 精确到毫秒的、以5分钟为单位的时间截
func Timestamp() int64 {
	currentMillisecond := time.Now().UnixNano() / int64(time.Millisecond)
	return currentMillisecond - (currentMillisecond % timestampGranularity)
}
