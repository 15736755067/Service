package path

import (
	"Service/common"
	"Service/db"
	"Service/log"
	"Service/units/lander"
	"Service/units/offer"
	"database/sql"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

//no cache
func dbGetAvailablePaths() []PathConfig {
	d := dbgetter()
	sql := "SELECT id, userId, redirectMode, directLink, status FROM Path WHERE deleted=0"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[path][dbGetAvailablePaths]Query: %s failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	var c PathConfig
	var arr []PathConfig
	for rows.Next() {
		if err := rows.Scan(&c.Id, &c.UserId, &c.RedirectMode, &c.DirectLink, &c.Status); err != nil {
			log.Errorf("[path][dbGetAvailablePaths] scan failed:%v", err)
			return nil
		}
		arr = append(arr, c)
	}
	return arr
}

//func DBGetUserPaths(userId int64) []PathConfig {
//	d := dbgetter()
//	sql := "SELECT id, userId, redirectMode, directLink, status FROM Path WHERE userId=? AND deleted=0"
//	rows, err := d.Query(sql)
//	if err != nil {
//		log.Errorf("[path][DBGetUserPaths]Query: %s failed:%v", sql, err)
//		return nil
//	}
//	defer rows.Close()

//	var c PathConfig
//	var arr []PathConfig
//	for rows.Next() {
//		if err := rows.Scan(&c.Id, &c.UserId, &c.RedirectMode, &c.DirectLink, &c.Status); err != nil {
//			log.Errorf("[path][DBGetAvailablePaths] scan failed:%v", err)
//			return nil
//		}
//		arr = append(arr, c)
//	}
//	return arr
//}

// with cache
func DBGetPath(pathId int64) (c PathConfig) {
	if pathDBCache.Enabled() {
		v := pathDBCache.Get(pathId)
		if v == nil {
			log.Errorf("[DBGetPath]Path(%d) does not exist in pathDBCache", pathId)
			return
		}
		return v.(PathConfig)
	}

	d := dbgetter()
	sql := "SELECT id, userId, redirectMode, directLink, status FROM Path WHERE id=?"
	row := d.QueryRow(sql, pathId)

	if err := row.Scan(&c.Id, &c.UserId, &c.RedirectMode, &c.DirectLink, &c.Status); err != nil {
		log.Errorf("[path][DBGetPath] pathId:%v scan failed:%v", pathId, err)
		return
	}
	return
}

//with cache
func DBGetPathLanders(pathId int64) (landers []PathLander) {
	if pathLanderDBCache.Enabled() {
		v := pathLanderDBCache.Get(pathId)
		if v == nil {
			log.Errorf("[DBGetPathLanders]Path(%d) does not exist in pathLanderDBCache", pathId)
			return
		}
		landers = make([]PathLander, len(v))
		for i, _ := range v {
			landers[i] = v[i].(PathLander)
		}
		return
	}

	d := dbgetter()
	sql := "SELECT landerId, weight FROM Lander2Path WHERE pathId=? AND deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql, pathId)
	if err != nil {
		log.Errorf("[path][DBGetPathLanders]Query sql:%v with pathId:%v failed:%v", sql, pathId, err)
		return nil
	}
	defer rows.Close()

	var pl PathLander
	for rows.Next() {
		err := rows.Scan(&pl.LanderId, &pl.Weight)
		if err != nil {
			log.Errorf("[path][DBGetPathLanders]Query Scan from sql:%v with pathId:%v failed:%v", sql, pathId, err)
			return nil
		}
		landers = append(landers, pl)
	}
	return
}

//no cache
func dbGetAvailablePathLanders() (pathLanders map[int64][]PathLander) {
	d := dbgetter()
	sql := "SELECT pathId, landerId, weight FROM Lander2Path WHERE deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[path][dbGetAvailablePathLanders]Query sql:%v failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	pathLanders = make(map[int64][]PathLander)
	pathId := int64(0)
	var pl PathLander
	for rows.Next() {
		err := rows.Scan(&pathId, &pl.LanderId, &pl.Weight)
		if err != nil {
			log.Errorf("[path][dbGetAvailablePathLanders]Query Scan from sql:%v failed:%v", sql, err)
			return nil
		}
		pathLanders[pathId] = append(pathLanders[pathId], pl)
	}
	return
}

//with cache
func DBGetPathOffers(pathId int64) (offers []PathOffer) {
	if pathOfferDBCache.Enabled() {
		v := pathOfferDBCache.Get(pathId)
		if v == nil {
			log.Errorf("[DBGetPathOffers]Path(%d) does not exist in pathOfferDBCache", pathId)
			return
		}
		offers = make([]PathOffer, len(v))
		for i, _ := range v {
			offers[i] = v[i].(PathOffer)
		}
		return
	}

	d := dbgetter()
	sql := "SELECT offerId, weight FROM Offer2Path WHERE pathId=? AND deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql, pathId)
	if err != nil {
		log.Errorf("[path][DBGetPathOffers]Query sql:%v with pathId:%v failed:%v", sql, pathId, err)
		return nil
	}
	defer rows.Close()

	var po PathOffer
	for rows.Next() {
		err := rows.Scan(&po.OfferId, &po.Weight)
		if err != nil {
			log.Errorf("[path][DBGetPathOffers]Query Scan from sql:%v with pathId:%v failed:%v", sql, pathId, err)
			return nil
		}
		offers = append(offers, po)
	}
	return
}

//no cache
func dbGetAvailablePathOffers() (pathOffers map[int64][]PathOffer) {
	d := dbgetter()
	sql := "SELECT pathId, offerId, weight FROM Offer2Path WHERE deleted=0 ORDER BY `order` ASC"
	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("[path][dbGetAvailablePathOffers]Query sql:%v failed:%v", sql, err)
		return nil
	}
	defer rows.Close()

	pathOffers = make(map[int64][]PathOffer)
	pathId := int64(0)
	var po PathOffer
	for rows.Next() {
		err := rows.Scan(&pathId, &po.OfferId, &po.Weight)
		if err != nil {
			log.Errorf("[path][dbGetAvailablePathOffers]Query Scan from sql:%v failed:%v", sql, err)
			return nil
		}

		pathOffers[pathId] = append(pathOffers[pathId], po)
	}
	return
}

var pathDBCache common.DBCache

func EnablePathDBCache() {
	log.Info("EnablePathDBCache Begin")
	lander.EnableLanderDBCache()
	offer.EnableOfferDBCache()

	enablePathLanderDBCache()
	enablePathOfferDBCache()

	origin := dbGetAvailablePaths()
	data := make([]common.HasID, len(origin))
	for i, _ := range origin {
		data[i] = origin[i]
	}
	pathDBCache.Init(data)
	log.Info("EnablePathDBCache End")
}
func DisablePathDBCache() {
	pathDBCache.Clear()
	disablePathOfferDBCache()
	disablePathLanderDBCache()

	offer.DisableOfferDBCache()
	lander.DisableLanderDBCache()
	log.Info("DisablePathDBCache")
}

var pathLanderDBCache common.DBSliceCache

func enablePathLanderDBCache() {
	origin := dbGetAvailablePathLanders()
	data := make(map[int64][]common.HasID, len(origin))
	for pathId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[pathId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[pathId][i] = s[i]
		}
	}
	pathLanderDBCache.Init(data)
}
func disablePathLanderDBCache() {
	pathLanderDBCache.Clear()
}

var pathOfferDBCache common.DBSliceCache

func enablePathOfferDBCache() {
	origin := dbGetAvailablePathOffers()
	data := make(map[int64][]common.HasID, len(origin))
	for pathId, s := range origin {
		if len(s) == 0 {
			continue
		}
		data[pathId] = make([]common.HasID, len(s))
		for i, _ := range s {
			data[pathId][i] = s[i]
		}
	}
	pathOfferDBCache.Init(data)
}
func disablePathOfferDBCache() {
	pathOfferDBCache.Clear()
}
