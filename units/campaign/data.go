package campaign

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"Service/common"
	"Service/db"
	"Service/log"
	"Service/units/ffrule"
	"Service/units/flow"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

var campaignDBCache common.DBCache

func EnableCampaignsDBCache() {
	log.Info("EnableCampaignsDBCache Begin")
	flow.EnableFlowDBCache()
	ffrule.EnableRuleDBCache()
	enableTSDBCache()
	enableUserCampaignsDBCache()

	origin := dbGetAvailableCampaigns()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}

	campaignDBCache.Init(data)
	log.Info("EnableCampaignsDBCache End")
}

func DisableCampaignsDBCache() {
	campaignDBCache.Clear()
	userCampaignsDBCache.Clear()
	disableTSDBCache()
	ffrule.DisableRuleDBCache()
	flow.DisableFlowDBCache()
	log.Info("DisableCampaignsDBCache")
}

var userCampaignsDBCache common.DBSliceCache

func enableUserCampaignsDBCache() {
	origin := dbGetAllUserAvailableCampaigns()
	data := make(map[int64][]common.HasID, len(origin))
	for userId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[userId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[userId][i] = s[i]
		}
	}
	userCampaignsDBCache.Init(data)
}

func disableUserCampaignsDBCache() {
	userCampaignsDBCache.Clear()
}

// no cache
func dbGetAllUserAvailableCampaigns() (userCampaigns map[int64][]CampaignConfig) {
	d := dbgetter()
	sql := "SELECT id, name, userId, hash, url, impPixelUrl, trafficSourceId, trafficSourceName, costModel, cpcValue, cpaValue, cpmValue, postbackUrl, pixelRedirectUrl, redirectMode, targetType, targetFlowId, targetUrl, status, country FROM TrackingCampaign WHERE deleted=0"
	rows, err := d.Query(sql)

	if err != nil {
		log.Errorf("[campaign][dbGetAllUserAvailableCampaigns]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	userCampaigns = make(map[int64][]CampaignConfig)

	for rows.Next() {
		var c CampaignConfig
		if err := rows.Scan(&c.Id, &c.Name, &c.UserId, &c.Hash, &c.Url, &c.ImpPixelUrl, &c.TrafficSourceId, &c.TrafficSourceName, &c.CostModel, &c.CPCValue, &c.CPAValue, &c.CPMValue, &c.PostbackUrl, &c.PixelRedirectUrl, &c.RedirectMode, &c.TargetType, &c.TargetFlowId, &c.TargetUrl, &c.Status, &c.Country); err != nil {
			log.Errorf("[campaign][dbGetAllUserAvailableCampaigns] scan failed:%v", err)
			return nil
		}

		c.TrafficSource, err = DBGetCampaignTrafficConfig(c.TrafficSourceId)
		if err != nil {
			log.Errorf("[campaign][dbGetAllUserAvailableCampaigns] DBGetCampaignTrafficConfig failed:%v", err)
		}

		userCampaigns[c.UserId] = append(userCampaigns[c.UserId], c)
	}

	return
}

var tsDBCache common.DBCache

func enableTSDBCache() {
	origin := dbGetAvailableCampaignTrafficConfigs()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	tsDBCache.Init(data)
}

func disableTSDBCache() {
	tsDBCache.Clear()
}

// no cache
func dbGetAvailableCampaignTrafficConfigs() (trafficSources []TrafficSourceConfig) {
	d := dbgetter()
	sql := "SELECT id, userId, name, postbackUrl, pixelRedirectUrl, impTracking, externalId, cost, params, campaignId, websiteId FROM TrafficSource"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[campaign][dbGetAvailableCampaignTrafficConfigs]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	arr := make([]TrafficSourceConfig, 0)
	var e, c, v, tsCampaign, tsWebsite string

	for rows.Next() {
		var trafficSource TrafficSourceConfig
		err = rows.Scan(&trafficSource.Id,
			&trafficSource.UserId,
			&trafficSource.Name,
			&trafficSource.PostbackURL,
			&trafficSource.PixelRedirectURL,
			&trafficSource.ImpTracking,
			&e, &c, &v, &tsCampaign, &tsWebsite)

		if err != nil {
			log.Errorf("Scan from sql:%v failed:%v", sql, err)
			return
		}

		if len(e) != 0 {
			err = json.Unmarshal([]byte(e), &trafficSource.ExternalId)
			if err != nil {
				log.Errorf("Unmarshal:[%s] to ExternalId failed:%v", e, err)
			}
		}

		if len(c) != 0 {
			err = json.Unmarshal([]byte(c), &trafficSource.Cost)
			if err != nil {
				log.Errorf("Unmarshal:[%s] to Cost failed:%v", c, err)
			}
		}

		if len(v) != 0 {
			err = json.Unmarshal([]byte(v), &trafficSource.Vars)
			if err != nil {
				log.Errorf("Unmarshal:[%s] to TrafficSourceParams failed:%v", v, err)
			}
		}

		if len(tsCampaign) != 0 {
			err = json.Unmarshal([]byte(tsCampaign), &trafficSource.TSCampaignId)
			if err != nil {
				log.Errorf("Unmarshal:[%s] to TSCampaignId failed:%v", tsCampaign, err)
			}
		}

		if len(tsWebsite) != 0 {
			err = json.Unmarshal([]byte(tsWebsite), &trafficSource.TSWebsitId)
			if err != nil {
				log.Errorf("Unmarshal:[%s] to TSWebsitId failed:%v", tsWebsite, err)
			}
		}

		arr = append(arr, trafficSource)
	}

	return arr
}

// with cache
func DBGetCampaignTrafficConfig(trafficSourceId int64) (
	trafficSource TrafficSourceConfig,
	err error,
) {
	if trafficSourceId == 0 {
		return
	}
	if tsDBCache.Enabled() {
		v := tsDBCache.Get(trafficSourceId)
		if v == nil {
			err = fmt.Errorf("traffic source data not exist in tsDBCache for id(%d)", trafficSourceId)
			return
		}
		trafficSource = v.(TrafficSourceConfig).GetCopy()
		return
	}

	d := dbgetter()
	sql := "SELECT id, userId, name, postbackUrl, pixelRedirectUrl, impTracking, externalId, cost, params, campaignId, websiteId FROM TrafficSource WHERE id=?"
	row := d.QueryRow(sql, trafficSourceId)
	var e, c, v, tsCampaign, tsWebsite string
	err = row.Scan(&trafficSource.Id,
		&trafficSource.UserId,
		&trafficSource.Name,
		&trafficSource.PostbackURL,
		&trafficSource.PixelRedirectURL,
		&trafficSource.ImpTracking,
		&e, &c, &v, &tsCampaign, &tsWebsite)

	if err != nil {
		log.Errorf("Scan from sql:%v failed:%v", sql, err)
		return
	}

	if len(e) != 0 {
		err = json.Unmarshal([]byte(e), &trafficSource.ExternalId)
		if err != nil {
			log.Errorf("Unmarshal:[%s] to ExternalId failed:%v", e, err)
		}
	}

	if len(c) != 0 {
		err = json.Unmarshal([]byte(c), &trafficSource.Cost)
		if err != nil {
			log.Errorf("Unmarshal:[%s] to Cost failed:%v", c, err)
		}
	}

	if len(v) != 0 {
		err = json.Unmarshal([]byte(v), &trafficSource.Vars)
		if err != nil {
			log.Errorf("Unmarshal:[%s] to TrafficSourceParams failed:%v", v, err)
		}
	}

	if len(tsCampaign) != 0 {
		err = json.Unmarshal([]byte(tsCampaign), &trafficSource.TSCampaignId)
		if err != nil {
			log.Errorf("Unmarshal:[%s] to TSCampaignId failed:%v", tsCampaign, err)
		}
	}

	if len(tsWebsite) != 0 {
		err = json.Unmarshal([]byte(tsWebsite), &trafficSource.TSWebsitId)
		if err != nil {
			log.Errorf("Unmarshal:[%s] to TSWebsitId failed:%v", tsWebsite, err)
		}
	}

	return
}

// no cache
func dbGetAvailableCampaigns() []CampaignConfig {
	d := dbgetter()
	sql := "SELECT id, name, userId, hash, url, impPixelUrl, trafficSourceId, trafficSourceName, costModel, cpcValue, cpaValue, cpmValue, postbackUrl, pixelRedirectUrl, redirectMode, targetType, targetFlowId, targetUrl, status, country FROM TrackingCampaign WHERE deleted=0"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[campaign][dbGetAvailableCampaigns]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var arr []CampaignConfig
	for rows.Next() {
		var c CampaignConfig
		if err := rows.Scan(&c.Id, &c.Name, &c.UserId, &c.Hash, &c.Url, &c.ImpPixelUrl, &c.TrafficSourceId, &c.TrafficSourceName, &c.CostModel, &c.CPCValue, &c.CPAValue, &c.CPMValue, &c.PostbackUrl, &c.PixelRedirectUrl, &c.RedirectMode, &c.TargetType, &c.TargetFlowId, &c.TargetUrl, &c.Status, &c.Country); err != nil {
			log.Errorf("[campaign][dbGetAvailableCampaigns] scan failed:%v", err)
			return nil
		}

		c.TrafficSource, err = DBGetCampaignTrafficConfig(c.TrafficSourceId)
		if err != nil {
			log.Errorf("[campaign][dbGetAvailableCampaigns] DBGetCampaignTrafficConfig failed:%v", err)
		}
		arr = append(arr, c)
	}

	return arr
}

// with cache
func DBGetUserCampaigns(userId int64) (arr []CampaignConfig) {
	if userCampaignsDBCache.Enabled() {
		v := userCampaignsDBCache.Get(userId)
		if v == nil {
			log.Errorf("campaign data not exist in userCampaignsDBCache for id(%d)", userId)
			return
		}
		arr = make([]CampaignConfig, len(v))
		for i, _ := range v {
			arr[i] = v[i].(CampaignConfig).GetCopy()
		}
		return
	}

	d := dbgetter()
	sql := "SELECT id, name, userId, hash, url, impPixelUrl, trafficSourceId, trafficSourceName, costModel, cpcValue, cpaValue, cpmValue, postbackUrl, pixelRedirectUrl, redirectMode, targetType, targetFlowId, targetUrl, status, country FROM TrackingCampaign WHERE userId=? AND deleted=0"
	rows, err := d.Query(sql, userId)
	if err != nil {
		log.Errorf("[campaign][DBGetUserCampaigns]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var c CampaignConfig
		if err := rows.Scan(&c.Id, &c.Name, &c.UserId, &c.Hash, &c.Url, &c.ImpPixelUrl, &c.TrafficSourceId, &c.TrafficSourceName, &c.CostModel, &c.CPCValue, &c.CPAValue, &c.CPMValue, &c.PostbackUrl, &c.PixelRedirectUrl, &c.RedirectMode, &c.TargetType, &c.TargetFlowId, &c.TargetUrl, &c.Status, &c.Country); err != nil {
			log.Errorf("[campaign][DBGetUserCampaigns] scan failed:%v", err)
			return nil
		}

		c.TrafficSource, err = DBGetCampaignTrafficConfig(c.TrafficSourceId)
		if err != nil {
			log.Errorf("[campaign][DBGetUserCampaigns] DBGetCampaignTrafficConfig failed:%v", err)
		}
		arr = append(arr, c)
	}
	return
}

// with cache
func DBGetCampaign(campaignId int64) (c CampaignConfig) {
	if campaignDBCache.Enabled() {
		v := campaignDBCache.Get(campaignId)
		if v == nil {
			log.Errorf("campaign data not exist in campaignDBCache for id(%d)", campaignId)
			return
		}
		return v.(CampaignConfig).GetCopy()
	}

	d := dbgetter()
	sql := "SELECT id, name, userId, hash, url, impPixelUrl, trafficSourceId, trafficSourceName, costModel, cpcValue, cpaValue, cpmValue, postbackUrl, pixelRedirectUrl, redirectMode, targetType, targetFlowId, targetUrl, status, country FROM TrackingCampaign WHERE id=?"
	row := d.QueryRow(sql, campaignId)

	if err := row.Scan(&c.Id, &c.Name, &c.UserId, &c.Hash, &c.Url, &c.ImpPixelUrl, &c.TrafficSourceId, &c.TrafficSourceName, &c.CostModel, &c.CPCValue, &c.CPAValue, &c.CPMValue, &c.PostbackUrl, &c.PixelRedirectUrl, &c.RedirectMode, &c.TargetType, &c.TargetFlowId, &c.TargetUrl, &c.Status, &c.Country); err != nil {
		log.Errorf("[campaign][DBGetCampaign] scan failed:%v", err)
		return
	}

	var err error
	c.TrafficSource, err = DBGetCampaignTrafficConfig(c.TrafficSourceId)
	if err != nil {
		log.Errorf("[campaign][DBGetCampaign] DBGetCampaignTrafficConfig failed:%v", err)
	}
	return c
}

// no cache(特殊，因为Load阶段不会使用)
func DBGetCampaignByHash(campaignHash string) (c CampaignConfig) {
	d := dbgetter()
	sql := "SELECT id, name, userId, hash, url, impPixelUrl, trafficSourceId, trafficSourceName, costModel, cpcValue, cpaValue, cpmValue, postbackUrl, pixelRedirectUrl, redirectMode, targetType, targetFlowId, targetUrl, status, country FROM TrackingCampaign WHERE hash=?"
	row := d.QueryRow(sql, campaignHash)

	if err := row.Scan(&c.Id, &c.Name, &c.UserId, &c.Hash, &c.Url, &c.ImpPixelUrl, &c.TrafficSourceId, &c.TrafficSourceName, &c.CostModel, &c.CPCValue, &c.CPAValue, &c.CPMValue, &c.PostbackUrl, &c.PixelRedirectUrl, &c.RedirectMode, &c.TargetType, &c.TargetFlowId, &c.TargetUrl, &c.Status, &c.Country); err != nil {
		log.Errorf("[campaign][DBGetCampaign] campaignHash:%s scan failed:%v", campaignHash, err)
		return
	}

	var err error
	c.TrafficSource, err = DBGetCampaignTrafficConfig(c.TrafficSourceId)
	if err != nil {
		log.Errorf("[campaign][DBGetCampaign] DBGetCampaignTrafficConfig:%v failed:%v", c.TrafficSourceId, err)
	}
	return c
}
