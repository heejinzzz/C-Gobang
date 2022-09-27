package main

import (
	"logCollector/collector"
	"os"
)

func init() {
	// 创建保存日志文件的文件夹
	err := os.MkdirAll(collector.UserLogSaveDir, 0777)
	if err != nil {
		panic("mkdir " + collector.UserLogSaveDir + " fail. Error: " + err.Error())
	}
	err = os.MkdirAll(collector.ErrorLogSaveDir, 0777)
	if err != nil {
		panic("mkdir " + collector.ErrorLogSaveDir + " fail. Error: " + err.Error())
	}
	err = os.MkdirAll(collector.GameLogSaveDir, 0777)
	if err != nil {
		panic("mkdir " + collector.GameLogSaveDir + " fail. Error: " + err.Error())
	}
}

func main() {
	collector.Wg.Add(3)
	go collector.UserLogCollect()
	go collector.ErrorLogCollect()
	go collector.GameLogCollect()
	collector.Wg.Wait()
}
