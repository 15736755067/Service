package db

import (
	"fmt"
	"sync"

	"Service/config"
	"Service/log"
	"github.com/go-redis/redis"
	//"gopkg.in/redis.v5"
)

var redisMux sync.RWMutex
var redisClients map[string]*redis.Client

func init() {
	redisClients = make(map[string]*redis.Client)
}

func GetRedisClient(title string) *redis.Client {
	redisMux.RLock()
	if client, ok := redisClients[title]; ok {
		redisMux.RUnlock()
		return client
	}
	redisMux.RUnlock()

	host := config.String(title, "host")
	password := config.String(title, "password")
	port := config.Int(title, "port")
	poolSize := config.Int(title, "pool")
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 6379
	}
	if poolSize == 0 {
		poolSize = 100
	}

	redisMux.Lock()
	defer redisMux.Unlock()
	if client, ok := redisClients[title]; ok {
		return client
	}

	client := newRedisClient(host, password, port, poolSize)
	if client == nil {
		log.Errorf("[GetRedisClient]New redis client %s failed:client is nil\n", title)
		return nil
	}

	log.Debugf("[GetRedisClient] new %s client= %v ok", title, client)
	redisClients[title] = client
	return client
}

func newRedisClient(host, password string, port, poolSize int) *redis.Client {
	addr := host + ":" + fmt.Sprintf("%d", port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
		PoolSize: poolSize,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Errorf("[redisClient][NewRedisClient] %s fail: %v", addr, err)
		return nil
	}

	return client
}
