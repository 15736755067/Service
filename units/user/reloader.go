package user

import (
	"Service/db"
	"Service/log"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"Service/units/blacklist"
	"Service/units/campaign"
	"Service/units/flow"
	"Service/units/lander"
	"Service/units/offer"
	"Service/units/path"
	"Service/units/rule"

	"github.com/go-redis/redis"
	//"gopkg.in/redis.v5"
)

// 关于用户campaign改变通知的实现方案
// 分成两个阶段：
// 1. 服务器启动阶段，也就是加载所有用户信息期间。
// 在这段时间内，如果用户信息又有新的改动，应该在加载完之后，重新加载所有有更新的用户
// 2. 服务器正常运行时间
// 在这段时间，收到一个更新一个即可。
// 关于redis断线重连：
// 目前试下来，redis客户端是能够支持断线重新连接的

var subscribe = "channel_campaign_changed_users"

var ssMux sync.RWMutex
var subscribedElements = make(map[string]bool)

const buffsize = 10000

func SubscribeAllElements() {
	ssMux.Lock()
	defer ssMux.Unlock()

	subscribedElements = map[string]bool{
		ElementUser:             true,
		ElementCampaign:         true,
		ElementFlow:             true,
		ElementRule:             true,
		ElementPath:             true,
		ElementLander:           true,
		ElementOffer:            true,
		ElementTrafficSource:    true,
		ElementAffiliateNetwork: true,
		ElementFFRule:           true,
	}
}

func ClearSubscribeElements() {
	ssMux.Lock()
	defer ssMux.Unlock()
	subscribedElements = make(map[string]bool)
}

func UnsubscribeElement(element string) {
	ssMux.Lock()
	defer ssMux.Unlock()
	subscribedElements[element] = false
}

func SubscribeElement(element string) {
	ssMux.Lock()
	defer ssMux.Unlock()
	subscribedElements[element] = true
}

func IsElementSubscribed(element string) bool {
	ssMux.RLock()
	defer ssMux.RUnlock()
	return subscribedElements[element]
}

var botblacklist = "channel_blacklist_changed_users"

// CollectorCampChangedUsers 收集服务器启动期间改变了Campaign的用户
type CollectorCampChangedUsers struct {
	Users          []string // 收集到的需要修改campaign的用户
	BlacklistUsers []string // 所有blacklist有改变的用户
	pubsub         *redis.PubSub

	ChanStop				chan bool
	ChanUser				chan string
	ChanBlacklist		chan string
}

func NewCollectorCampChangeUsers() *CollectorCampChangedUsers {
	return &CollectorCampChangedUsers {
		ChanUser: make(chan string, buffsize),
		ChanBlacklist: make(chan string, buffsize),

		ChanStop: make(chan bool),
	}
}

// Stop 停止收集
func (c *CollectorCampChangedUsers) Close() {
	c.ChanStop <- true

	c.pubsub.Close()
	close(c.ChanUser)
	close(c.ChanBlacklist)
}

//Update handle user's update information
func (c *CollectorCampChangedUsers) Update() {
	log.Debug("user.reloader.CollectorCampChangedUsers.Update begin") 
	if c == nil {
		log.Error("user reloader CollectorCampChangedUsers pointer is nil and exit Update")
		return
	}

	go func(c *CollectorCampChangedUsers) {
		defer func() {
			if x := recover(); x != nil {
				log.Error("user.reloader.CollectorCampChangedUsers.Update  panics:", x)
			}
		}()

		log.Debug("user.reloader.CollectorCampChangedUsers.Update start go into for select")
		for {
			select {
				case msg := <- c.ChanBlacklist:
					blacklist.ReloadUserBlacklist(msg)
				case msg := <- c.ChanUser:
					log.Debugf("user.reloader.CollectorCampChangedUsers.Update user channel receive: %s\n", msg)
					if err := ReloadUserInfo(msg); err != nil {
						log.Errorf("user.reloader.CollectorCampChangedUsers.Update ReloadUserInfo failed with err(%s)\n", err.Error())
					}
				case  <- c.ChanStop:
					close(c.ChanStop)
					return
			}
		}
	}(c)
}

// Start 启动收集协程
func (c *CollectorCampChangedUsers) Start() {
	cli := db.GetRedisClient("MSGQUEUE")
	log.Infof("user CollectorCampChangedUsers: running with MSGQUEUE redis:%v...", cli)

	var err error
	c.pubsub = cli.Subscribe(subscribe, botblacklist)
	if err != nil {
		log.Errorf("collector: PSubscribe %v failed:%v", subscribe, err)
		return
	}

	go func(pubsub *redis.PubSub) {
		defer func() {
			if x := recover(); x != nil {
				log.Error("CollectorCampChangedUsers.Start panics:", x)
			}
		}()

		for {
			received, err := pubsub.ReceiveMessage()
			if err != nil {
				log.Warnf("collector: receive from %v failed:%v. stop collecting.", subscribe, err)
				return
			}

			log.Infof("collector: user:%v campaign changed", received.Payload)
			if received.Channel == botblacklist {
				//c.BlacklistUsers = append(c.BlacklistUsers, received.Payload)
				c.ChanBlacklist <- received.Payload
			} else {
				log.Debug("user.reloader.CollectorCampChangedUsers.start receive user payload=%v and send into channel\n", received.Payload)
				c.ChanUser <- received.Payload
				//c.Users = append(c.Users, received.Payload)
			}
		}
	}(c.pubsub)
}

const (
	ActionAdd    = "add"
	ActionDel    = "delete"
	ActionUpdate = "update"
)
const (
	ElementUser             = "user"
	ElementCampaign         = "campaign"
	ElementFlow             = "flow"
	ElementRule             = "rule"
	ElementPath             = "path"
	ElementLander           = "lander"
	ElementOffer            = "offer"
	ElementTrafficSource    = "trafficSource"
	ElementAffiliateNetwork = "affiliateNetwork"
	ElementFFRule           = "ffrule"
)

// Reloader 当用户的campaign信息有更新的时候，要重新加载这个用户的campaign信息
type Reloader struct {
}

// Running 在后台持续更新用户数据
// 应该在加载所有的用户信息之后，启动这个
// 防止加载过程中有更新
// 此协程不进行存盘操作，所以不需要gracestop
// Playload:userId.action.element.id(userId为0表示system)
func (r Reloader) Running() {
	defer func() {
		if x := recover(); x != nil {
			log.Error(x, string(debug.Stack()))
		}
	}()
	redis := db.GetRedisClient("MSGQUEUE")
	log.Infof("user reloader: running with MSGQUEUE redis:%v...", redis)

	// redis.S
	pubsub := redis.Subscribe(subscribe, botblacklist)
	if pubsub == nil {
		log.Errorf("user reloader: PSubscribe  failed\n")
		return
	}
	/*
	if err != nil {
		log.Errorf("reloader: PSubscribe %v failed:%v", subscribe, err)
		return
	}
	*/

	for {
		received, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Errorf("user reloader: receive from %v failed:%v", subscribe, err)
			continue
		}

		log.Infof("%s user reloader: user %v notification\n", received.Channel, received.Payload)
		// 直接加载这个用户相关信息即可
		if received.Channel == botblacklist {
			blacklist.ReloadUserBlacklist(received.Payload)
		} else {
			log.Debugf("user.reloader.CollectorCampChangedUsers.Update user_campaign_change info : %s", received.Payload)
			if err := ReloadUserInfo(received.Payload); err != nil {
				log.Errorf("[Reloader][Running]ReloadUserInfo failed with err(%s)\n", err.Error())
			}
		}
	}
}

// ReloadUser 重新加载用户信息
// info:userId.action.element.id
func ReloadUserInfo(info string) (err error) {
	s := strings.Split(info, ".")
	if len(s) < 4 {
		return fmt.Errorf("info(%s) is invalid", info)
	}
	if !IsElementSubscribed(s[2]) {
		return fmt.Errorf("info(%s).element is not subscribed yet", info)
	}
	id, err := strconv.ParseInt(s[3], 10, 64)
	if err != nil {
		return fmt.Errorf("info(%s).id is an invalid integer", info)
	}

	switch s[1] { // action
	case ActionAdd:
		fallthrough
	case ActionUpdate:
		switch s[2] { // element
		case ElementUser:
			return InitUser(id)
		case ElementCampaign:
			return campaign.InitCampaign(id)
		case ElementFlow:
			return flow.InitFlow(id)
		case ElementRule:
			return rule.InitRule(id)
		case ElementPath:
			return path.InitPath(id)
		case ElementLander:
			return lander.InitLander(id)
		case ElementOffer:
			return offer.InitOffer(id)
		case ElementTrafficSource:
			return campaign.InitTrafficSource(id)
		case ElementAffiliateNetwork:
			return offer.InitAffiliateNetwork(id)
		case ElementFFRule:
			return campaign.InitFFRuleCampaigns(id)
		default:
			return fmt.Errorf("info(%s).element is invalid", info)
		}
	case ActionDel:
		switch s[2] { // element
		case ElementFFRule:
			return campaign.DeleteFFRuleCampaigns(id)
		default:
			// 暂时屏蔽这一条，以防有bug时的问题扩大
			log.Infof("[ReloadUserInfo]Omit %s\n", info)
		}
	default:
		return fmt.Errorf("info(%s).action is invalid", info)
	}

	return nil
}
