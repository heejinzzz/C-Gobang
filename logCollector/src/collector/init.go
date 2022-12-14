package collector

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v9"
	"logCollector/errorRecorder"
	"sync"
)

const LogSaveDir = "/data/logs"

var kafkaClusterAddrs = []string{"C-Gobang-kafka100:9092", "C-Gobang-kafka101:9092", "C-Gobang-kafka102:9092"}
var redisClusterAddrs = []string{
	"172.31.0.10:6379",
	"172.31.0.11:6379",
	"172.31.0.12:6379",
	"172.31.0.13:6379",
	"172.31.0.14:6379",
	"172.31.0.15:6379",
}

var redisClient *redis.ClusterClient

var Wg sync.WaitGroup

func init() {
	// 创建 kafka topics
	createTopics()

	// 连接 redis 集群
	opt := redis.ClusterOptions{
		Addrs: redisClusterAddrs,
	}
	redisClient = redis.NewClusterClient(&opt)

	res := redisClient.Ping(context.Background()).Val()
	if res != "PONG" {
		errorRecorder.RecordError("[logCollector][connect redis cluster fail]")
		panic("connect redis cluster fail")
	}
}

func createTopics() {
	admin, err := sarama.NewClusterAdmin(kafkaClusterAddrs, nil)
	if err != nil {
		panic(err)
	}
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 3,
	}
	err = admin.CreateTopic(errorLogTopic, topicDetail, false)
	if err != nil {
		panic(err)
	}
	err = admin.CreateTopic(userLogTopic, topicDetail, false)
	if err != nil {
		panic(err)
	}
	err = admin.CreateTopic(gameLogTopic, topicDetail, false)
	if err != nil {
		panic(err)
	}
}
