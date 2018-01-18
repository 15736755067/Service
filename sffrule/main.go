//
//
package main

import (
	"flag"
	"os"
	"runtime/debug"

	"Service/config"
	"Service/db"
	"Service/log"
	"Service/units/ffrule"
	"Service/units/user"
)

func main() {
	defer func() {
		log.Alertf("%d Quit main()\n", os.Getpid())
		log.Alert(string(debug.Stack()))
	}()
	flag.Parse()

	// 按照UTC时间来执行定时任务
	logAdapter := config.String("LOG", "adapter")
	logConfig := config.String("LOG", "jsonconfig")
	logAsync := config.Bool("LOG", "async")
	if logAdapter == "" {
		logAdapter = "console"
	}
	if logConfig == "" {
		logConfig = `{"level":7}`
	}
	log.Init(logAdapter, logConfig, logAsync)
	defer func() {
		log.Flush()
	}()

	// 只用监控FFRule的变化即可
	user.SubscribeElement(user.ElementFFRule)
	collector := new(user.CollectorCampChangedUsers)
	collector.Start()

	if err := ffrule.InitAllRules(); err != nil {
		panic(err.Error())
	}
	collector.Stop()
	reloader := new(user.Reloader)
	go reloader.Running()

	log.Infof("collected users:%+v", collector.Users)
	// redis 要能够连接
	redisClient := db.GetRedisClient("MSGQUEUE")
	if redisClient == nil {
		log.Errorf("Connect MSGQUEUE redis server failed.")
		return
	}
	log.Debugf("Connect MSGQUEUE redis success: redisClient:%p", redisClient)

	redisClient = db.GetRedisClient("LOCALFFCACHE")
	if redisClient == nil {
		log.Errorf("Connect LOCALFFCACHE redis server failed.")
		return
	}
	log.Debugf("Connect LOCALFFCACHE redis success: redisClient:%p", redisClient)

	if err := ffrule.Start(ffrule.ModeConsumer); err != nil {
		panic(err.Error())
	}

	ch := make(chan int)
	<-ch
}
