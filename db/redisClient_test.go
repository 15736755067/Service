package db

import (
	"Service/log"
	"testing"

	"gopkg.in/redis.v5"
)

func TestRedis(t *testing.T) {
	log.Init("console", `{"level":6}`, true)

	redisHost := "localhost"
	redisPort := 6379
	client := NewRedisClient(redisHost, redisPort)
	client.Set("adbund", 0, 0)
	val, err := client.Get("test").Result()
	t.Log(val, err)

	IncrByXX := redis.NewScript(`
		if redis.call("GET", KEYS[1]) ~= false then
			 redis.call("INCRBY", KEYS[1], ARGV[1])
		end

		if redis.call("HEXISTS", KEYS[2],"Field") == 1 then
			 redis.call("HINCRBY", KEYS[2],"Field",1)
		else
			 redis.call("HSET", KEYS[2],"Field",1)
		end
		return nil

	`)
	n, err := IncrByXX.Run(client, []string{"adbund", "adbund_test"}, 2).Result()
	t.Log(n, err)
}
