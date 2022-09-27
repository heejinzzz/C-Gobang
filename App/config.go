package main

type GameType int32

const (
	ServerIP = "127.0.0.1"
	userManagerServerAddr = ServerIP + ":10050"
	gameManagerServerAddr = ServerIP + ":10051"

	getFreeCoinAmount = 20

	juniorGameType GameType = 0
	mediumGameType GameType = 1
	highGameType   GameType = 2

	AcceptGameWaitingSeconds = 8

	checkerboardRowNumber = 15
	checkerboardColNumber = 15

	shootWaitingSeconds = 30
)
