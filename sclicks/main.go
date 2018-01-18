package main

import (
	"database/sql"
	"flag"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"Service/config"
	"Service/db"
	"Service/log"
	"Service/request"

	"gopkg.in/redis.v5"
)

func main() {
	defer func() {
		log.Alertf("%d Quit main()\n", os.Getpid())
		log.Alert(string(debug.Stack()))
	}()
	flag.Parse()

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
	localRedis := db.GetRedisClient("LOCALREQCACHE")
	if localRedis == nil {
		log.Error("Connect LOCALREQCACHE redis failed.")
		return
	}
	log.Debugf("Connect LOCALREQCACHE redis success: localRedis:%p", localRedis)

	remoteDB := db.GetDB("REMOTEREQCACHE")
	if remoteDB == nil {
		log.Error("Connect REMOTEREQCACHE db failed.")
		return
	}
	log.Debugf("Connect REMOTEREQCACHE redis success: remoteDB:%p", remoteDB)

	request.InitRemoteCacheStmt()
	defer request.CloseRemoteCacheStmt()

	if err := SyncLocalClicksToRemote(localRedis, remoteDB); err != nil {
		panic(err.Error())
	}
}

func SyncLocalClicksToRemote(localRedis *redis.Client, remoteDB *sql.DB) (err error) {
	keys, err := localRedis.Keys("*").Result()
	if err != nil {
		return err
	}
	log.Infof("[SyncLocalClicksToRemote]%d clicks got!\n", len(keys))

	str := ""
	i := int64(0)
	wait := sync.WaitGroup{}
	for _, key := range keys {
		str, err = localRedis.Get(key).Result()
		if err != nil {
			log.Errorf("[SyncLocalClicksToRemote]localRedis.Get(%s) failed with err(%s)\n", key, err.Error())
			continue
		}
		req := request.CacheStr2Req(str)
		if req == nil {
			log.Errorf("[SyncLocalClicksToRemote]request.CacheStr2Req(%s) failed for key(%s)\n", str, key)
			continue
		}

		if req.OfferId() > 0 || (req.CampaignId() > 0 && req.FlowId() == 0) {
			wait.Add(1)
			go func(req request.Request) {
				if !req.CacheSave(time.Duration(-1), config.ClickCacheTime) {
					log.Errorf("[SyncLocalClicksToRemote]req.CacheSave failed for key(%s)\n", key)
				} else {
					log.Debugf("[SyncLocalClicksToRemote]req.CacheSave finished for key(%s)\n", key)
				}
				atomic.AddInt64(&i, 1)
				wait.Done()
			}(req)
		}
	}
	wait.Wait()
	log.Infof("[SyncLocalClicksToRemote]%d clicks synced!\n", i)
	return nil
}
