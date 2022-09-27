package collector

import (
	"bufio"
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"logCollector/errorRecorder"
	"os"
	"strings"
	"time"
)

const userLogTopic = "userLogs"
const UserLogSaveDir = LogSaveDir + "/userLogs"

type userLogCollectorHandler struct{}

func (userLogCollectorHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (userLogCollectorHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (handler userLogCollectorHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		splits1 := strings.Split(string(msg.Value), "[")
		if len(splits1) < 2 {
			return errors.New("[log format error][user_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		splits2 := strings.Split(splits1[1], " ")
		if len(splits2) != 2 {
			return errors.New("[log format error][user_log \"" + string(msg.Value) + "\" is not in correct format]")
		}
		splits3 := strings.Split(splits2[0], "-")
		logFileName := "userLog-" + strings.Join(splits3, "") + ".log"
		for !redisClient.SetNX(context.Background(), logFileName, 1, -1).Val() {
			time.Sleep(100 * time.Millisecond)
		}
		err := writeLogIntoFile(string(msg.Value), UserLogSaveDir+"/"+logFileName)
		redisClient.Del(context.Background(), logFileName)
		if err != nil {
			return errors.New("[write log into file error][" + err.Error() + "]")
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func writeLogIntoFile(logStatement string, filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(logStatement + "\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	return err
}

func UserLogCollect() {
	defer Wg.Done()

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(kafkaClusterAddrs, "userLogCollector", config)
	if err != nil {
		errorRecorder.RecordError("[logCollector][create userLogCollector consumerGroup failed][" + err.Error() + "]")
		panic(err)
	}

	go func() {
		for err := range consumerGroup.Errors() {
			errorRecorder.RecordError("[logCollector][userLogCollector consume error][" + err.Error() + "]")
			panic(err)
		}
	}()

	handler := userLogCollectorHandler{}

	err = consumerGroup.Consume(context.Background(), []string{userLogTopic}, handler)
	if err != nil {
		errorRecorder.RecordError("[logCollector][userLogCollector error][" + err.Error() + "]")
		panic(err)
	}
}
