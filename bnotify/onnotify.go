// 用于接收系统的后台通知
// 用于控制系统内部的一些行为
// bnotify:background notification
package bnotify

import (
	"Service/db"
	"Service/log"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"Service/util/pprof"
	"github.com/go-redis/redis"
)

var msgChannel = "channel_system_notification"
var notifications = make([]string, 0)

// Start 启动收集协程
func Start() {
	cli := db.GetRedisClient("MSGQUEUE")
	log.Infof("bnotify running with MSGQUEUE redis:%v...", cli)

	// redis.S
	pubsub := cli.Subscribe(msgChannel)
	if pubsub == nil {
		panic("bnotify PSubscribe  failed")
	}
	/*
	if err != nil {
		panic(fmt.Sprintf("PSubscribe %v failed:%v", msgChannel, err))
	}
	*/

	go func(pubsub *redis.PubSub ) {
		defer func() {
			if x := recover(); x != nil {
				log.Error(x, string(debug.Stack()))
			}
		}()

		for {
			received, err := pubsub.ReceiveMessage()
			if err != nil {
				log.Errorf("bnotify receive from %v failed:%v", msgChannel, err)
				continue
			}

			log.Infof("bnotify received %s from channel %s\n", received.Payload, received.Channel)
			// 直接加载这个用户相关信息即可
			if err := doAction(received.Payload); err != nil {
				log.Errorf("bNotify.DoAction failed with err(%s)\n", err.Error())
			}
		}
	}(pubsub)
}

const (
	ActionEnable  = "enable"
	ActionDisable = "disable"
	ActionSet     = "set"
)

const (
	EntityPprof    = "pprof"
	EntityExpvar   = "expvar"
	EntityLogLevel = "loglevel"
)

// 执行后台系统命令
// command:action.entity[.value]
func doAction(command string) (err error) {
	s := strings.Split(command, ".")
	var action, entity string
	var value int64
	switch len(s) {
	case 2: // action.entity
		action = s[0]
		entity = s[1]
	case 3: // action.entity.value
		action = s[0]
		entity = s[1]
		value, err = strconv.ParseInt(s[2], 10, 64)
		if err != nil {
			return fmt.Errorf("command(%s).value is an invalid integer", command)
		}
	default:
		return fmt.Errorf("command(%s) is invalid", command)
	}

	switch action {
	case ActionEnable:
		switch entity {
		case EntityPprof:
			pprof.Activate()
		default:
			return fmt.Errorf("command(%s).entity is invalid", command)
		}
	case ActionDisable:
		switch entity {
		case EntityPprof:
			pprof.Inactivate()
		default:
			return fmt.Errorf("command(%s).entity is invalid", command)
		}
	case ActionSet:
		switch entity {
		case EntityLogLevel:
			log.SetLevel(int(value))
		default:
			return fmt.Errorf("command(%s).entity is invalid", command)
		}
	default:
		return fmt.Errorf("command(%s).action is invalid", command)
	}
	return nil
}
