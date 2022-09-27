package collector

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"logCollector/errorRecorder"
	"strings"
	"time"
)

const errorLogTopic = "errorLogs"
const ErrorLogSaveDir = LogSaveDir + "/errorLogs"

type errorLogCollectorHandler struct{}

func (errorLogCollectorHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (errorLogCollectorHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (errorLogCollectorHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		splits1 := strings.Split(string(msg.Value), "[")
		if len(splits1) < 2 {
			return errors.New("[log format error][error_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		splits2 := strings.Split(splits1[1], " ")
		if len(splits2) != 2 {
			return errors.New("[log format error][error_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		splits3 := strings.Split(splits2[0], "-")
		logFileName := "errorLog-" + strings.Join(splits3, "") + ".log"
		for !redisClient.SetNX(context.Background(), logFileName, 1, -1).Val() {
			time.Sleep(100 * time.Millisecond)
		}
		err := writeLogIntoFile(string(msg.Value), ErrorLogSaveDir+"/"+logFileName)
		redisClient.Del(context.Background(), logFileName)
		if err != nil {
			return errors.New("[write log into file error][" + err.Error() + "]")
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func ErrorLogCollect() {
	defer Wg.Done()

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(kafkaClusterAddrs, "errorLogCollector", config)
	if err != nil {
		errorRecorder.RecordError("[logCollector][create errorLogCollector consumerGroup failed][" + err.Error() + "]")
		panic(err)
	}

	go func() {
		for err := range consumerGroup.Errors() {
			errorRecorder.RecordError("[logCollector][errorLogCollector consume error][" + err.Error() + "]")
			panic(err)
		}
	}()

	handler := errorLogCollectorHandler{}

	err = consumerGroup.Consume(context.Background(), []string{errorLogTopic}, handler)
	if err != nil {
		errorRecorder.RecordError("[logCollector][errorLogCollector error][" + err.Error() + "]")
		panic(err)
	}
}
