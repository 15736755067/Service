package main

import (
	"Service/db"
	"flag"
	"fmt"
	"net/http"

	_ "Service/util/pprof"

	"Service/bnotify"
	"Service/common"
	"Service/config"
	"Service/gracequit"
	"Service/log"
	"Service/request"
	"Service/servehttp"
	"Service/tracking"
	"Service/units"
	_ "Service/units/blacklist"
	"Service/units/user"
	"Service/util/ip"
	"time"
)

func main() {
	help := flag.Bool("help", false, "show help")
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

	// 启动保存协程
	gracequit.StartGoroutine(func(c gracequit.StopSigChan) {
		tracking.Saving(db.GetDB("DB"), c)
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

	// Postback不需要ip库的支持
	//ip2location.Open(config.String("IP", "path"))

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

	collector := user.NewCollectorCampChangeUsers()
	collector.Start()
	//collector := new(user.CollectorCampChangedUsers)
	
	if err := units.Init(); err != nil {
		panic(err.Error())
	}
	//collector.Stop()
/*
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


	http.HandleFunc("/status", Status)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/robots.txt", robots)
	http.HandleFunc(config.String("DEFAULT", "s2spostback"), OnS2SPostback)
	http.HandleFunc(config.String("DEFAULT", "conversionUpload"), OnUploadConversions)
	http.HandleFunc(config.String("DEFAULT", "conversionpixelurl"), OnConversionPixel)
	http.HandleFunc(config.String("DEFAULT", "conversionscripturl"), OnConversionScript)

	log.Error(StartServe())

	log.Infof("stopping background goroutines...")
	gracequit.StopAll()
	log.Infof("background goroutines stopped")
}

func StartServe() error {
	reqServer := &http.Server{Addr: ":" + config.GetPostbackPort(), Handler: http.DefaultServeMux}
	log.Info("Start listening postback at", config.GetPostbackPort())
	return servehttp.Serve(reqServer) // reqServer.ListenAndServe()
}

func Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "It works!")
}

func OnS2SPostback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	units.OnS2SPostback(w, r)
}
func OnUploadConversions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	units.OnUploadConversions(w, r)
}

const base64GifPixel = "R0lGODlhAQABAIAAAP///wAAACwAAAAAAQABAAACAkQBADs="

func OnConversionPixel(w http.ResponseWriter, r *http.Request) {
	//TODO 去除重复的clickId和transactionId的conversions
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)

	units.OnConversionPixel(w, r)

	/* 备忘
	w.Header().Set("Content-Type", "image/gif")
	output, _ := base64.StdEncoding.DecodeString(base64GifPixel)
	w.Write(output)
	*/
}

func OnConversionScript(w http.ResponseWriter, r *http.Request) {
	//TODO 去除重复的clickId和transactionId的conversions
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	units.OnConversionScript(w, r)

}

var robotsTxt = []byte(`User-agent: *
Disallow: /`)

func robots(w http.ResponseWriter, r *http.Request) {
	log.Infof("robots.txt requested from:%v UserAgent:%v", ip.GetIP(r), r.UserAgent())
	w.Write(robotsTxt)
}
