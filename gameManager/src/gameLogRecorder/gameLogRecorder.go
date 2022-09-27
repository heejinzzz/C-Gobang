package gameLogRecorder

import (
	"github.com/Shopify/sarama"
	"log"
	"strconv"
)

var kafkaClusterAddrs = []string{"C-Gobang-kafka100:9092", "C-Gobang-kafka101:9092", "C-Gobang-kafka102:9092"}

const gameLogTopic = "gameLogs"

var gameLogRecorder sarama.AsyncProducer

func init() {
	config := sarama.NewConfig()
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true

	var err error
	gameLogRecorder, err = sarama.NewAsyncProducer(kafkaClusterAddrs, config)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range gameLogRecorder.Errors() {
			log.Println(msg)
			panic(err)
		}
	}()
}

func RecordGameLog(gameID int32, logStatement string) {
	prefix := "[" + strconv.Itoa(int(gameID)) + "]"
	producerMessage := sarama.ProducerMessage{
		Topic: gameLogTopic,
		Value: sarama.StringEncoder(prefix + logStatement),
	}
	gameLogRecorder.Input() <- &producerMessage
}