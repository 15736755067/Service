package blacklist

import (
	"Service/db"
	"Service/log"
	"Service/util/ipcmp"
	"database/sql"
	"encoding/json"
	"strings"
)

// dbgetter 默认的拿数据库的东西
// 方便测试的地方替换这个接口
var dbgetter = func() *sql.DB {
	return db.GetDB("DB")
}

// BotBlacklistConfig 对应UserBotBlacklist表
type BotBlacklistConfig struct {
	Id        int64
	UserId    int64
	IpRange   []ipcmp.IPCompare
	UserAgent []string
	Enabled   bool
}

// BuildBlacklistConfig 直接构建一个BotBlacklistConfig，方便测试用
func BuildBlacklistConfig(id, userID int64, ipRange, userAgent string, enabled int) (BotBlacklistConfig, error) {
	var c BotBlacklistConfig
	c.Id = id
	c.UserId = userID

	var err error
	c.IpRange, err = parseIpRanges(ipRange)
	if err != nil {
		log.Errorf("[blacklist][DBGetUserBlacklists]parseIpRanges user:%v UserBotBlacklist.id:%v ipRange:%s failed:%v, this rule will be ignored", userID, c.Id, ipRange, err)
		return c, err
	}

	c.UserAgent, err = parseUserAgents(userAgent)
	if err != nil {
		log.Errorf("[blacklist][DBGetUserBlacklists]parseUserAgents user:%v UserBotBlacklist.id:%v userAgent:%s failed:%v, this rule will be ignored", userID, c.Id, userAgent, err)
		return c, err
	}

	c.Enabled = enabled != 0
	return c, nil
}

func DBInsertUserBlacklist(name string, userId int64, ipRange []string, userAgent []string) error {
	d := dbgetter()
	sql := `INSERT INTO UserBotBlacklist(name,userId,ipRange,userAgent,enabled) VALUES(?,?,?,?,1)`
	ir := ""
	for i, ip := range ipRange {
		if i == 0 {
			ir = "[\"" + ip + "\""
		} else {
			ir += ",\"" + ip + "\""
		}
		if i == len(ipRange)-1 {
			ir += "]"
		}
	}
	ua := ""
	for i, u := range userAgent {
		if i == 0 {
			ua = "[\"" + u + "\""
		} else {
			ua += ",\"" + u + "\""
		}
		if i == len(userAgent)-1 {
			ua += "]"
		}
	}
	if ua == "" {
		ua = "[]"
	}
	_, err := d.Exec(sql, name, userId, ir, ua)
	if err != nil {
		return err
	}
	return nil
}

// DBGetUserBlacklists 加载用户的blacklist列表
// 如果某条有错误，不影响其它的配置
func DBGetUserBlacklists(userId int64) []BotBlacklistConfig {
	d := dbgetter()
	sqlQuery := `SELECT id, ipRange, userAgent, enabled FROM UserBotBlacklist WHERE userId=? and enabled=1 and deleted=0`
	rows, err := d.Query(sqlQuery, userId)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("[blacklist][DBGetUserBlacklists]Query %s with userId:%v failed:%v", sqlQuery, userId, err)
		return nil
	}
	defer rows.Close()

	var arr []BotBlacklistConfig

	var id int64
	var ipRange string
	var userAgent string
	var enabled int

	for rows.Next() {
		err := rows.Scan(&id,
			&ipRange,
			&userAgent,
			&enabled,
		)

		if err == sql.ErrNoRows {
			// 如果没有数据返回，当作正常来处理
			break
		}

		if err != nil {
			log.Errorf("[blacklist][DBGetUserBlacklists]Scan failed:%v", err)
			return nil
		}

		c, err := BuildBlacklistConfig(id, userId, ipRange, userAgent, enabled)
		if err != nil {
			// Build失败的要忽略
			log.Errorf("[blacklist]BuildBlacklistConfig failed with ipRange(%s)&userAgent(%s) for user(%d)\n", ipRange, userAgent, userId)
			continue
		}

		arr = append(arr, c)
	}
	return arr
}

func parseIpRanges(s string) ([]ipcmp.IPCompare, error) {
	// ["1.2.3.4", "1.2.3.4-1.2.3.7", "5.6.7.8"]
	var ips []string
	err := json.Unmarshal([]byte(s), &ips)
	if err != nil {
		return nil, err
	}

	var cmps []ipcmp.IPCompare

	for _, s := range ips {
		ips := strings.Split(s, "-")
		ip1 := ips[0]
		if len(ips) == 1 {
			cmp, err := ipcmp.NewIPCompare(ip1)
			if err != nil {
				return nil, err
			}

			cmps = append(cmps, cmp)
			continue
		}

		ip2 := ips[1]
		cmp, err := ipcmp.NewIPCompareRange(ip1, ip2)
		if err != nil {
			return nil, err
		}
		cmps = append(cmps, cmp)
	}
	return cmps, nil
}

func parseUserAgents(s string) ([]string, error) {
	s = strings.TrimSpace(s)

	// 允许是空的
	if len(s) == 0 {
		return nil, nil
	}

	var useragents []string
	err := json.Unmarshal([]byte(s), &useragents)
	return useragents, err
}
