package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	_ "Service/util/pprof"

	"Service/bnotify"
	"Service/common"
	"Service/config"
	"Service/db"
	"Service/gracequit"
	"Service/log"
	"Service/request"
	"Service/servehttp"
	"Service/tracking"
	"Service/units"
	"Service/units/blacklist"
	"Service/units/user"
	"Service/util/ip"
	"Service/util/ip2location"
)

func main() {
	defer func() {
		log.Alertf("%d Quit main()\n", os.Getpid())
		log.Alert(string(debug.Stack()))
	}()

	help := flag.Bool("help", false, "show help")
	blacklistPath := flag.String("blacklist", "", "global blacklist.txt path. If it's empty, disable global blacklist. You can download blacklist here: https://myip.ms/files/blacklist/general/full_blacklist_database.zip")

	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if err := config.LoadConfig(true); err != nil {
		panic(err.Error())
	}

	common.Init(config.GetSvrUniqueStr(), "")

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

	common.WritePidFile()

	bnotify.Start()

	if len(*blacklistPath) != 0 {
		if err := blacklist.EnableBlacklist(*blacklistPath); err != nil {
			log.Errorf("EnableBlacklist with path:%v failed:%v", *blacklistPath, err)
		}
	}

	ip2location.Open(config.String("IP", "path"))

	// 启动保存协程
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		tracking.Saving(db.GetDB("DB"), c)
	})

	// 启动CampaignMap保存协程
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		tracking.CampMapSaving(db.GetDB("DB"), c)
	})

	// 启动Conversion保存
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		tracking.SavingConversions(db.GetDB("DB"), c)
	})

	// 启动汇总协程
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		secondsAdStatis := config.Int("TRACKING", "adstatis-interval")
		interval := time.Duration(secondsAdStatis) * time.Second
		if interval == 0 {
			log.Warnf("config: TRACKING:adstatis-interval not found. Using default interval: 10 minutes")
			interval = 10 * 60 * time.Second
		}
		tracking.Gathering(c, interval)
	})

	secondsIpReferrerDomain := config.Int("TRACKING", "ip-interval")
	interval := time.Duration(secondsIpReferrerDomain) * time.Second
	if interval == 0 {
		log.Warnf("config: TRACKING:ip-interval not found. Using default interval: 10 minutes")
		interval = 10 * 60 * time.Second
	}

	// 启动AdIPStatis表的汇总协程
	tracking.InitIPGatherSaver(&gracequit.G, db.GetDB("DB"), interval)

	// 启动AdReferrerStatis表的汇总协程
	tracking.InitRefGatherSaver(&gracequit.G, db.GetDB("DB"), interval)

	// 启动AdReferrerDomainStatis表的汇总协程
	tracking.InitDomainGatherSaver(&gracequit.G, db.GetDB("DB"), interval)

	request.InitRemoteCacheStmt(
		config.Bool("REMOTEREQCACHE", "asyncwrite"),
		config.Int("REMOTEREQCACHE", "asyncbuffer"))
	defer request.CloseRemoteCacheStmt()
	// 启动RemoteCache的异步保存
	gracequit.StartGoroutineN(func(c gracequit.StopSigChan) {
		request.AsyncingRemoteCache(c)
	}, config.Int("REMOTEREQCACHE", "asyncwriters"))
/*
	// redis 要能够连接
	redisClient := db.GetRedisClient("MSGQUEUE")
	if redisClient == nil {
		log.Errorf("Connect redis server failed.")
		return
	}
	log.Debugf("Connect redis success: redisClient:%p", redisClient)
*/
	user.SubscribeAllElements()

	//collector := new(user.CollectorCampChangedUsers)
	collector := user.NewCollectorCampChangeUsers()
	collector.Start()

	if err := units.Init(); err != nil {
		panic(err.Error())
	}

	/*
	collector.Stop()

	reloader := new(user.Reloader)
	go reloader.Running()

	log.Infof("collected users:%+v", collector.Users)
	//log.Debugf("redisClient:%p", db.GetRedisClient("MSGQUEUE"))

	for _, info := range collector.Users {
		user.ReloadUserInfo(info)
	}

	for _, uid := range collector.BlacklistUsers {
		blacklist.ReloadUserBlacklist(uid)
	}
*/


	collector.Update()
	defer collector.Close()

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/robots.txt", robots)
	http.HandleFunc("/dmr", units.OnDoubleMetaRefresh)
	http.HandleFunc("/status", Status)
	http.HandleFunc("/status/", Status)
	http.HandleFunc(config.String("DEFAULT", "lpofferrequrl"), units.OnLPOfferRequest)
	http.HandleFunc(config.String("DEFAULT", "lpclickurl"), units.OnLandingPageClick)
	http.HandleFunc(config.String("DEFAULT", "lpclickopturl"), units.OnLandingPageClick)
	http.HandleFunc(config.String("DEFAULT", "impressionurl"), units.OnImpression)

	reqServer := &http.Server{Addr: ":" + config.GetEnginePort(), Handler: http.DefaultServeMux}
	log.Info("Start listening request at", config.GetEnginePort())
	log.Error(servehttp.Serve(reqServer))
	log.Infof("http server stopped. stopping other goroutines...")
	// 只需要在HTTP服务器退出的时候，等待协程退出

	log.Infof("stopping background goroutines...")

	gracequit.StopAll()
	log.Infof("background goroutines stopped")
	log.Infof("%d Quit main()\n", os.Getpid())
}

func Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "It works!")
}

var robotsTxt = []byte(`User-agent: *
Disallow: /`)

func robots(w http.ResponseWriter, r *http.Request) {
	log.Infof("robots.txt requested from:%v UserAgent:%v", ip.GetIP(r), r.UserAgent())
	w.Write(robotsTxt)
}
