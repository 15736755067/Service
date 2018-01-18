package offer

import (
	"Service/common"
	"Service/db"
	"Service/log"
	"database/sql"
	"time"
)

//func DBGetAllOffers() []OfferConfig {
//	return nil
//}

func dbGetAvailableTimezones() []Timezone {
	d := dbgetter()
	sql := `SELECT id, name, region, utcShift FROM Timezones`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[offer][dbGetAvailableTimezones]Query %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var c Timezone
	var arr []Timezone
	for rows.Next() {
		err := rows.Scan(&c.Id,
			&c.Name,
			&c.Region,
			&c.UtcShift,
		)

		if err != nil {
			log.Errorf("[offer][dbGetAvailableTimezones]Scan %s failed:%v", sql, err)
			return nil
		}

		arr = append(arr, c)
	}
	return arr
}

//no cache
func dbGetAvailableOffers() []OfferConfig {
	d := dbgetter()
	sql := `SELECT id, name, userId, url, AffiliateNetworkId,AffiliateNetworkName, postbackUrl, payoutMode, payoutValue,capEnabled,dailyCap,capTimezoneId,redirectOfferId FROM Offer WHERE deleted=0`
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[offer][DBGetAvailableOffers]Query %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var c OfferConfig
	var arr []OfferConfig
	for rows.Next() {
		err := rows.Scan(&c.Id,
			&c.Name,
			&c.UserId,
			&c.Url,
			&c.AffiliateNetworkId,
			&c.AffiliateNetworkName,
			&c.PostbackUrl,
			&c.PayoutMode,
			&c.PayoutValue,
			&c.CapEnabled,
			&c.DailyCap,
			&c.CapTimezoneId,
			&c.RedirectOfferId,
		)

		if err != nil {
			log.Errorf("[offer][DBGetAvailableOffers]Scan %s failed:%v", sql, err)
			return nil
		}

		arr = append(arr, c)
	}
	return arr
}

//func DBGetUserOffers(userId int64) []OfferConfig {
//	d := dbgetter()
//	sql := `SELECT id, name, userId, url, AffiliateNetworkId, AffiliateNetworkName, postbackUrl, payoutMode, payoutValue FROM Offer WHERE userId=? and deleted=0`
//	rows, err := d.Query(sql, userId)
//	if err != nil {
//		log.Errorf("[offer][DBGetAvailableOffers]Query %s with userId:%v failed:%v", sql, userId, err)
//		return nil
//	}
//	defer rows.Close()

//	var c OfferConfig
//	var arr []OfferConfig
//	for rows.Next() {
//		err := rows.Scan(&c.Id,
//			&c.Name,
//			&c.UserId,
//			&c.Url,
//			&c.AffiliateNetworkId,
//			&c.AffiliateNetworkName,
//			&c.PostbackUrl,
//			&c.PayoutMode,
//			&c.PayoutValue,
//		)

//		if err != nil {
//			log.Errorf("[offer][DBGetUserOffers]Scan failed:%v", err)
//			return nil
//		}

//		arr = append(arr, c)
//	}
//	return arr
//}

//with cache
func DBGetOffer(offerId int64) (c OfferConfig) {
	if offerDBCache.Enabled() {
		v := offerDBCache.Get(offerId)
		if v == nil {
			log.Errorf("[DBGetOffer]Offer(%d) does not exist in offerDBCache\n", offerId)
			return
		}
		return v.(OfferConfig)
	}

	d := dbgetter()
	sql := `SELECT id, name, userId, url, AffiliateNetworkId, AffiliateNetworkName, postbackUrl, payoutMode, payoutValue,capEnabled,dailyCap,capTimezoneId,redirectOfferId FROM Offer WHERE id=? and deleted=0`
	row := d.QueryRow(sql, offerId)
	err := row.Scan(&c.Id,
		&c.Name,
		&c.UserId,
		&c.Url,
		&c.AffiliateNetworkId,
		&c.AffiliateNetworkName,
		&c.PostbackUrl,
		&c.PayoutMode,
		&c.PayoutValue,
		&c.CapEnabled,
		&c.DailyCap,
		&c.CapTimezoneId,
		&c.RedirectOfferId,
	)

	if err != nil {
		log.Errorf("[offer][DBGetOffer] offerId:%d Scan failed:%v", offerId, err)
		return c
	}

	return c
}

func GetLocation(timezonesId int64) *time.Location {
	tz := GetTimezone(timezonesId)
	if tz.Id == 0 {
		return nil
	}
	loc, _ := time.LoadLocation(tz.Region)
	return loc
}
func GetTimezone(timezonesId int64) Timezone {
	var c Timezone
	if offerDBCache.Enabled() {
		v := timezonesDBCache.Get(timezonesId)
		if v == nil {
			log.Errorf("[GetLocation]timezone(%d) does not exist in GetLocation\n", timezonesId)
			return c
		}
		return v.(Timezone)
	}

	d := dbgetter()
	sql := `SELECT id, name, region, utcShift FROM Timezones WHERE id=? `
	row := d.QueryRow(sql, timezonesId)

	err := row.Scan(&c.Id,
		&c.Name,
		&c.Region,
		&c.UtcShift,
	)

	if err != nil {
		log.Errorf("[offer][GetTimezone]Scan %s failed:%v", sql, err)
		return c
	}

	return c
}

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

var offerDBCache common.DBCache
var timezonesDBCache common.DBCache

func EnableTimezonesDBCache() {
	log.Info("EnableTimezonesDBCache Begin")
	origin := dbGetAvailableTimezones()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	timezonesDBCache.Init(data)
	log.Info("EnableTimezonesDBCache End")
}

func DisableTimezonesDBCache() {
	timezonesDBCache.Clear()
	log.Info("DisableTimezonesDBCache")
}

func EnableOfferDBCache() {
	log.Info("EnableOfferDBCache Begin")
	origin := dbGetAvailableOffers()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	offerDBCache.Init(data)
	log.Info("EnableOfferDBCache End")
	EnableTimezonesDBCache()
}

func DisableOfferDBCache() {
	offerDBCache.Clear()
	log.Info("DisableOfferDBCache")
	DisableTimezonesDBCache()
}
