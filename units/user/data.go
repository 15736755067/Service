package user

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

// 获取所有User，包含被删除以及被停止服务的User
func DBGetAllUsers() []UserConfig {
	return nil
}

// 获取未被删除、未停止服务的User
//no cache
func dbGetAvailableUsers() []UserConfig {
	d := dbgetter()
	sql := "SELECT id, idText, rootDomainRedirect, status FROM User WHERE deleted=0"
	rows, err := d.Query(sql)

	if err != nil {
		log.Errorf("[user][DBGetAvailableUsers]Query: %s failed:%v", sql, err)
		return nil
	}

	if rows == nil {
		log.Errorf("user.data.DBGetAvailableUsers.Query: %s has no row ", sql)
		return nil
	}

	var arr []UserConfig
	for rows.Next() {
		var c UserConfig
		if err := rows.Scan(&c.Id, &c.IdText, &c.RootDomainRedirect, &c.Status); err != nil {
			log.Errorf("[user][DBGetAvailableUsers] scan failed:%v and continue", err)
			//rows.Close()
			//return nil
			continue
		}

		arr = append(arr, c)
	}
	rows.Close()

	for i, c := range arr {
		if c.Id > 0 {
			arr[i].Domains = DBGetUserDomains(c.Id)
		}
	}

	return arr
}

//no cache
func dbGetUserInfo(userId int64) (c UserConfig) {
	d := dbgetter()
	sql := "SELECT id, idText, rootDomainRedirect, status FROM User WHERE id=?"
	row := d.QueryRow(sql, userId)

	if err := row.Scan(&c.Id, &c.IdText, &c.RootDomainRedirect, &c.Status); err != nil {
		log.Errorf("[user][DBGetUserInfo] scan failed:%v", err)
		return
	}

	c.Domains = DBGetUserDomains(userId)
	return c
}

//no cache
func dbGetAllUserDomains() (userDomains map[int64][]UserDomain) {
	d := dbgetter()
	sql := "SELECT userId, domain, main, verified, customize FROM UserDomain WHERE deleted=0"
	rows, err := d.Query(sql)

	if err != nil {
		log.Errorf("[user][dbGetAllUserDomains]Query: %s failed:%v", sql, err)
		return nil
	}
	if rows == nil {
		log.Infof("user.data.dbGetAllUserDomains().Query() %s has no rows\n", sql)
		return nil
	}
	defer rows.Close()

	userDomains = make(map[int64][]UserDomain)
	var ud UserDomain
	var userId int64
	var main, verified, customize int

	for rows.Next() {
		if err := rows.Scan(&userId, &ud.Domain, &main, &verified, &customize); err != nil {
			log.Errorf("[user][dbGetAllUserDomains]scan failed:%v and continue", err)
			//return
			continue
		}

		ud.Main = (main == 1)
		ud.Verified = (verified == 1)
		ud.Customize = (customize == 1)
		userDomains[userId] = append(userDomains[userId], ud)
	}

	return
}

//with cache
func DBGetUserDomains(userId int64) (domains []UserDomain) {
	if userId <= 0 {
		return
	}

	log.Debugf("[DBGetUserDomains]userId%d domains(%v)\n", userId, domains)
	if userDomainsDBCache.Enabled() {
		v := userDomainsDBCache.Get(userId)
		if v == nil {
			log.Errorf("[DBGetUserDomains]User(%d) does not exist in userDomainsDBCache", userId)
			return
		}

		domains = make([]UserDomain, len(v))
		for i, _ := range v {
			domains[i] = v[i].(UserDomain)
		}

		return
	}

	d := dbgetter()
	sql := "SELECT domain, main, verified, customize FROM UserDomain WHERE deleted=0 AND userId = ?"
	rows, err := d.Query(sql, userId)
	if err != nil {
		log.Errorf("user.data.DBGetUserDomains.Query: %s for user%d failed:%v", sql, userId, err)
		return nil
	}

	if rows == nil {
		log.Infof("user.data.DBGetUserDomains.Query: %s for user%d has no rows", sql, userId )
		return nil
	}
	defer rows.Close()

	var ud UserDomain
	var main, verified, customize int
	for rows.Next() {
		if err := rows.Scan(&ud.Domain, &main, &verified, &customize); err != nil {
			log.Errorf("user.data.DBGetUserDomains.row.scan user%d failed:%v and continue", userId, err)
			//return
			continue
		}

		ud.Main = (main == 1)
		ud.Verified = (verified == 1)
		ud.Customize = (customize == 1)

		domains = append(domains, ud)
	}

	return
}

var userDomainsDBCache common.DBSliceCache

func EnableUserDomainsDBCache() {
	log.Info("EnableUserDomainsDBCache Begin")

	origin := dbGetAllUserDomains()
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

	userDomainsDBCache.Init(data)
	log.Info("EnableUserDomainsDBCache End")
}

func DisableUserDomainsDBCache() {
	userDomainsDBCache.Clear()
	log.Info("DisableUserDomainsDBCache")
}
