package lander

import (
	"Service/common"
	"Service/db"
	"Service/log"
	"database/sql"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

//func DBGetAllLanders() []LanderConfig {
//	return nil
//}

//no cache
func dbGetAvailableLanders() []LanderConfig {
	d := dbgetter()
	sql := "SELECT id, name, userId, url, numberOfOffers FROM Lander WHERE deleted=0"

	rows, err := d.Query(sql)
	if err != nil {
		log.Errorf("lander.data.DBGetAllLanders Query: %s failed:%v and return", sql, err)
		return nil
	}

	if rows == nil {
		log.Infof("lander.data.DBGetAllLanders Query: %s has no result and return", sql)
		return nil
	}
	defer rows.Close()

	var l LanderConfig
	var arr []LanderConfig

	for rows.Next() {
		if err := rows.Scan(&l.Id, &l.Name, &l.UserId, &l.Url, &l.NumberOfOffers); err != nil {
			log.Errorf("lander.data.DBGetAllLanders rows.Scan failed:%v", err)
			continue
			//return nil
		}

		arr = append(arr, l)
	}

	return arr
}

//func DBGetUserLanders(userId int64) []LanderConfig {
//	d := dbgetter()
//	sql := "SELECT id, name, userId, url, numberOfOffers FROM Lander WHERE userId=? and deleted=0"
//	rows, err := d.Query(sql, userId)
//	if err != nil {
//		log.Errorf("[lander][DBGetAllLanders]Query: %s failed:%v", sql, err)
//		return nil
//	}
//	defer rows.Close()

//	var l LanderConfig
//	var arr []LanderConfig
//	for rows.Next() {
//		if err := rows.Scan(&l.Id, &l.Name, &l.UserId, &l.Url, &l.NumberOfOffers); err != nil {
//			log.Errorf("[lander][DBGetAllLanders] scan failed:%v", err)
//			return nil
//		}
//		arr = append(arr, l)
//	}
//	return arr
//}

//with cache
func DBGetLander(landerId int64) (c LanderConfig) {
	if landerDBCache.Enabled() {
		v := landerDBCache.Get(landerId)
		if v == nil {
			log.Errorf("[DBGetLander]Lander(%d) does not exist in landerDBCache", landerId)
			return
		}
		return v.(LanderConfig)
	}

	d := dbgetter()
	sql := "SELECT id, name, userId, url, numberOfOffers FROM Lander WHERE id=? and deleted=0"
	row := d.QueryRow(sql, landerId)

	if err := row.Scan(&c.Id, &c.Name, &c.UserId, &c.Url, &c.NumberOfOffers); err != nil {
		log.Errorf("[lander][DBGetAllLanders] landerId:%v scan failed:%v", landerId, err)
		return c
	}

	return c
}

var landerDBCache common.DBCache

func EnableLanderDBCache() {
	log.Info("EnableLanderDBCache Begin")

	origin := dbGetAvailableLanders()
	if  origin == nil {
		log.Info("lander.data.EnableLanderDBCache.dbGetAvailableLanders() get nil")
		return
	}

	data := make([]common.HasID, len(origin))

	for i, _ := range origin {
		data[i] = origin[i]
	}

	landerDBCache.Init(data)
	log.Info("EnableLanderDBCache End")
}

func DisableLanderDBCache() {
	landerDBCache.Clear()
	log.Info("DisableLanderDBCache")
}
