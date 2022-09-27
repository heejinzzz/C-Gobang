package errorRecorder

import (
	"github.com/Shopify/sarama"
	"log"
	"time"
)

var kafkaClusterAddrs = []string{"C-Gobang-kafka100:9092", "C-Gobang-kafka101:9092", "C-Gobang-kafka102:9092"}

const errorLogTopic = "errorLogs"

var errorRecorder sarama.AsyncProducer

func init() {
	config := sarama.NewConfig()
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true

	var err error
	errorRecorder, err = sarama.NewAsyncProducer(kafkaClusterAddrs, config)
	if err != nil {
		panic(err)
	}

	go func() {
		for msg := range errorRecorder.Errors() {
			log.Println(msg)
			panic(err)
		}
	}()
}

func RecordError(errorMsg string) {
	timePrefix := time.Now().Format("[2006-01-02 15:04:05]")
	producerMsg := sarama.ProducerMessage{
		Topic: errorLogTopic,
		Value: sarama.StringEncoder(timePrefix + errorMsg),
	}
	errorRecorder.Input() <- &producerMsg
}
