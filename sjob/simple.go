//
//
package main

import (
	"flag"
	"os"
	"runtime/debug"
	"time"

	"Service/config"
	"Service/db"
	"Service/log"
	"Service/util/cron"
)

func main() {
	defer func() {
		log.Alertf("%d Quit main()\n", os.Getpid())
		log.Alert(string(debug.Stack()))
	}()
	flag.Parse()

	// 按照UTC时间来执行定时任务
	c := cron.NewWithLocation(time.Now().UTC().Location())
	if c == nil {
		panic("c is nil")
	}

	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}

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

	// redis 要能够连接
	redisClient := db.GetRedisClient("MSGQUEUE")
	if redisClient == nil {
		log.Errorf("Connect redis server failed.")
		return
	}
	log.Debugf("Connect redis success: redisClient:%p", redisClient)

	var err error

	//TODO 将到期的custom plan，自动转为basic plan，而不是修改用户status
	err = c.AddFunc("@midnight", userStatusOverdueUpdate)
	if err != nil {
		log.Errorf("AddFunc(@midnight,userStatusOverdueUpdate):%s\n", err.Error())
		return
	}
	// 变成events免费策略，不再需要检查用户的透支时间
	//	err = c.AddFunc("@every 30s", userStatusOverageUpdate)
	//	if err != nil {
	//		log.Errorf("AddFunc(@every 30s,userStatusOverageUpdate):%s\n", err.Error())
	//		return
	//	}
	err = c.AddFunc("@midnight", clearOutdatedClickDetails)
	if err != nil {
		log.Errorf("AddFunc(@midnight,clearOutdatedClickDetails):%s\n", err.Error())
		return
	}

	c.Start()
	log.Info("cron started")
	ch := make(chan int)
	<-ch
}
