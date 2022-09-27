package collector

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"logCollector/errorRecorder"
	"strings"
	"time"
)

const gameLogTopic = "gameLogs"
const GameLogSaveDir = LogSaveDir + "/gameLogs"

type gameLogCollectorHandler struct{}

func (gameLogCollectorHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (gameLogCollectorHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler gameLogCollectorHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		splits1 := strings.Split(string(msg.Value), "]")
		if len(splits1) < 2 {
			return errors.New("[log format error][game_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		if splits1[0][0] != '[' {
			return errors.New("[log format error][game_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		logFileName := "gameLog-" + splits1[0][1:] + ".log"
		for !redisClient.SetNX(context.Background(), logFileName, 1, -1).Val() {
			time.Sleep(100 * time.Millisecond)
		}
		err := writeLogIntoFile(string(msg.Value), GameLogSaveDir+"/"+logFileName)
		redisClient.Del(context.Background(), logFileName)
		if err != nil {
			return errors.New("[write log into file error][" + err.Error() + "]")
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func GameLogCollect() {
	defer Wg.Done()

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(kafkaClusterAddrs, "gameLogCollector", config)
	if err != nil {
		errorRecorder.RecordError("[logCollector][create gameLogCollector consumerGroup failed][" + err.Error() + "]")
		panic(err)
	}

	go func() {
		for err := range consumerGroup.Errors() {
			errorRecorder.RecordError("[logCollector][gameLogCollector consume error][" + err.Error() + "]")
			panic(err)
		}
	}()

	handler := gameLogCollectorHandler{}

	err = consumerGroup.Consume(context.Background(), []string{gameLogTopic}, handler)
	if err != nil {
		errorRecorder.RecordError("[logCollector][gameLogCollector error][" + err.Error() + "]")
		panic(err)
	}
}
