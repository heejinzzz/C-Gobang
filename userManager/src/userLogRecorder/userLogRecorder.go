package userLogRecorder

import (
	"github.com/Shopify/sarama"
	"log"
	"time"
)

var kafkaClusterAddrs = []string{"C-Gobang-kafka100:9092", "C-Gobang-kafka101:9092", "C-Gobang-kafka102:9092"}

const userLogTopic = "userLogs"

var userLogRecorder sarama.AsyncProducer

func init() {
	config := sarama.NewConfig()
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true

	var err error
	userLogRecorder, err = sarama.NewAsyncProducer(kafkaClusterAddrs, config)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range userLogRecorder.Errors() {
			log.Println(msg)
			panic(err)
		}
	}()
}

func RecordUserLog(logStatement string) {
	timePrefix := time.Now().Format("[2006-01-02 15:04:05]")
	producerMsg := sarama.ProducerMessage{
		Topic: userLogTopic,
		Value: sarama.StringEncoder(timePrefix + logStatement),
	}
	userLogRecorder.Input() <- &producerMsg
}
