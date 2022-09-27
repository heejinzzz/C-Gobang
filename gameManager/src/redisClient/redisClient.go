package redisClient

import (
	"context"
	"gameManager/errorRecorder"
	"github.com/go-redis/redis/v9"
)

var redisClusterAddr = []string{
	"172.31.0.10:6379",
	"172.31.0.11:6379",
	"172.31.0.12:6379",
	"172.31.0.13:6379",
	"172.31.0.14:6379",
	"172.31.0.15:6379",
}

var Client *redis.ClusterClient

func init() {
	opt := redis.ClusterOptions{
		Addrs: redisClusterAddr,
	}
	Client = redis.NewClusterClient(&opt)

	res := Client.Ping(context.Background()).Val()
	if res != "PONG" {
		errorRecorder.RecordError("[userManager][connect redis cluster fail]")
		panic("connect redis cluster fail")
	}
}
