package blacklist

import (
	"strconv"
	"strings"
	"sync"

	"Service/log"
	"Service/util/ipcmp"
)

var lock sync.RWMutex
var blacklists = make(map[int64]*UserBlacklists)

// UserBlacklists 用户所有的BotBlacklistConfig
type UserBlacklists struct {
	lists []BotBlacklistConfig
}

// UserReqAllowed 判断是否允许这次调用
// remoteAddr (ip地址，支持192.168.0.155:52323这种格式)
// userAgent User Agent
func UserReqAllowed(userId int64, remoteAddr, userAgent string) bool {
	b := GetUserBlacklist(userId)
	if b == nil {
		return true
	}

	allowed, err := b.Allowed(remoteAddr, userAgent)
	if err != nil {
		log.Errorf("UserReqAllowed userId:%v remoteAddr:%v userAgent:%v call Allowed failed:%v",
			userId, remoteAddr, userAgent, err)
		return true
	}
	return allowed
}

// Allowed 返回一个IP和和个UserAgent的用户是否允许访问campaign url
func (l *UserBlacklists) Allowed(remoteIP, userAgent string) (bool, error) {
	ipInt, err := ipcmp.IPToInt64(remoteIP)
	if err != nil {
		return true, err
	}

	for _, l := range l.lists {
		log.Debug("UserBlacklists", remoteIP, userAgent, l.IpRange)
		if !l.Allowed(ipInt, userAgent) {
			return false, nil
		}
	}

	return true, nil
}

// GetUserBlacklist 获取一个用户的BotBlacklistConfig
func GetUserBlacklist(userID int64) *UserBlacklists {
	lock.RLock()
	defer lock.RUnlock()

	return blacklists[userID]
}

// SetUserBlacklist 设置一个用户的过滤项
func SetUserBlacklist(userID int64, p *UserBlacklists) {
	lock.Lock()
	defer lock.Unlock()

	blacklists[userID] = p
}

// ReloadUserBlacklist 重新加载一个用户的blacklist
func ReloadUserBlacklist(userIDs string) {
	ids := strings.Split(userIDs, ",")
	for _, id := range ids {
		uId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Errorf("[ReloadUserBlacklist]userIDs(%s) contains invalid integer\n", userIDs)
			return
		}
		lists := DBGetUserBlacklists(uId)
		p := &UserBlacklists{
			lists: lists,
		}
		SetUserBlacklist(uId, p)
	}
	return
}

// UserAgentAllowed 返回一个User Agent是否允许访问
// 如果非空，则包含的都禁止访问
// 如果为空，则全部禁止访问
func (b *BotBlacklistConfig) UserAgentAllowed(ua string) bool {
	if len(b.UserAgent) == 0 {
		return false
	}

	for _, banned := range b.UserAgent {
		if strings.Contains(ua, banned) {
			return false
		}
	}
	return true
}

// IPAllowed 返回一个IP是否允许访问
func (b *BotBlacklistConfig) IPAllowed(ip ipcmp.IP_INT) bool {
	for _, r := range b.IpRange {
		if r.IpIntIn(ip) {
			return false
		}
	}
	return true
}

// Allowed 返回一个IP和一个user agent是否允许访问
func (b *BotBlacklistConfig) Allowed(ip ipcmp.IP_INT, ua string) bool {
	ipAllowed := b.IPAllowed(ip)
	if ipAllowed {
		return true
	}
	return b.UserAgentAllowed(ua)
}
