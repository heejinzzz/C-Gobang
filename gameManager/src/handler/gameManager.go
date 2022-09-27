package handler

import (
	"context"
	"gameManager/errorRecorder"
	"gameManager/gameLogRecorder"
	"gameManager/mysqlClient"
	pb "gameManager/proto"
	"gameManager/redisClient"
	"log"
	"strconv"
	"time"
)

type GameManager struct{}

func (*GameManager) RequireNewGame(ctx context.Context, req *pb.RequireNewGameRequest) (*pb.RequireNewGameResponse, error) {
	log.Println("Received GameManager.RequireNewGame request:", req)
	WaitingPlayersLock.Lock()
	WaitingPlayers[req.UserID] = nil
	WaitingPlayersLock.Unlock()
	if req.GameType == 0 {
		JuniorGameRoomLock.Lock()
		JuniorGameRoom[req.UserID] = nil
		delete(JuniorGameRoom, req.UserID)
		if len(JuniorGameRoom) == 0 {
			ch := make(chan int32, 1)
			JuniorGameRoom[req.UserID] = ch
			JuniorGameRoomLock.Unlock()
			opponentID := <-ch
			return &pb.RequireNewGameResponse{Message: "找到对局!", OpponentID: opponentID}, nil
		}
		for k := range JuniorGameRoom {
			opponentID := k
			JuniorGameRoom[k] <- req.UserID
			delete(JuniorGameRoom, k)
			JuniorGameRoomLock.Unlock()
			return &pb.RequireNewGameResponse{Message: "找到对局！", OpponentID: opponentID}, nil
		}
	}
	if req.GameType == 1 {
		MediumGameRoomLock.Lock()
		MediumGameRoom[req.UserID] = nil
		delete(MediumGameRoom, req.UserID)
		if len(MediumGameRoom) == 0 {
			ch := make(chan int32, 1)
			MediumGameRoom[req.UserID] = ch
			MediumGameRoomLock.Unlock()
			opponentID := <-ch
			return &pb.RequireNewGameResponse{Message: "找到对局!", OpponentID: opponentID}, nil
		}
		for k := range MediumGameRoom {
			opponentID := k
			MediumGameRoom[k] <- req.UserID
			delete(MediumGameRoom, k)
			MediumGameRoomLock.Unlock()
			return &pb.RequireNewGameResponse{Message: "找到对局！", OpponentID: opponentID}, nil
		}
	}
	HighGameRoomLock.Lock()
	HighGameRoom[req.UserID] = nil
	delete(HighGameRoom, req.UserID)
	if len(HighGameRoom) == 0 {
		ch := make(chan int32, 1)
		HighGameRoom[req.UserID] = ch
		HighGameRoomLock.Unlock()
		opponentID := <-ch
		return &pb.RequireNewGameResponse{Message: "找到对局!", OpponentID: opponentID}, nil
	}
	for k := range HighGameRoom {
		opponentID := k
		HighGameRoom[k] <- req.UserID
		delete(HighGameRoom, k)
		HighGameRoomLock.Unlock()
		return &pb.RequireNewGameResponse{Message: "找到对局！", OpponentID: opponentID}, nil
	}
	return &pb.RequireNewGameResponse{Message: "对局系统错误！"}, nil
}

func (*GameManager) CancelRequireGame(ctx context.Context, req *pb.CancelRequireGameRequest) (*pb.CancelRequireGameResponse, error) {
	log.Println("Received GameManager.CancelRequireGame request:", req)
	if req.GameType == 0 {
		JuniorGameRoomLock.Lock()
		delete(JuniorGameRoom, req.UserID)
		JuniorGameRoomLock.Unlock()
		return &pb.CancelRequireGameResponse{Message: "取消对局成功"}, nil
	}
	if req.GameType == 1 {
		MediumGameRoomLock.Lock()
		delete(MediumGameRoom, req.UserID)
		MediumGameRoomLock.Unlock()
		return &pb.CancelRequireGameResponse{Message: "取消对局成功"}, nil
	}
	HighGameRoomLock.Lock()
	delete(HighGameRoom, req.UserID)
	HighGameRoomLock.Unlock()
	return &pb.CancelRequireGameResponse{Message: "取消对局成功"}, nil
}

func (*GameManager) AcceptNewGame(ctx context.Context, req *pb.AcceptNewGameRequest) (*pb.AcceptNewGameResponse, error) {
	log.Println("Received GameManager.AcceptNewGame request:", req)
	WaitingPlayersLock.Lock()
	if WaitingPlayers[req.UserID] != nil {
		WaitingPlayersLock.Unlock()
		return &pb.AcceptNewGameResponse{Message: "确认超时，请重新寻找对局！"}, nil
	}
	if WaitingPlayers[req.OpponentID] == nil {
		ch := make(chan int32, 1)
		WaitingPlayers[req.UserID] = ch
		WaitingPlayersLock.Unlock()
		timer := time.NewTimer(playerAcceptGameWaitingWindow)
		select {
		case gameID := <-ch:
			isBlack := false
			if Games[gameID].BlackPlayer == req.UserID {
				isBlack = true
			}
			return &pb.AcceptNewGameResponse{GameID: gameID, Black: isBlack}, nil
		case <-timer.C:
			WaitingPlayersLock.Lock()
			WaitingPlayers[req.UserID] = nil
			delete(WaitingPlayers, req.UserID)
			WaitingPlayers[req.OpponentID] = make(chan int32, 1)
			WaitingPlayersLock.Unlock()
			return &pb.AcceptNewGameResponse{Message: "对方拒绝或放弃了对局！"}, nil
		}
	}
	gameID := createNewGame(req.UserID, req.OpponentID, req.GameType)
	WaitingPlayers[req.OpponentID] <- gameID
	delete(WaitingPlayers, req.OpponentID)
	WaitingPlayersLock.Unlock()
	timer := time.NewTimer(playerReplyWaitingWindow)
	GameWaitingReplyTimersLock.Lock()
	GameWaitingReplyTimers[gameID] = timer
	GameWaitingReplyTimersLock.Unlock()
	channel1 := make(chan ShootPos, 1)
	channel2 := make(chan ShootPos, 1)
	ShootChannelsLock.Lock()
	ShootChannels[gameID] = []chan ShootPos{channel1, channel2}
	ShootChannelsLock.Unlock()
	go waitToDeleteGame(gameID, timer)
	isBlack := false
	if Games[gameID].BlackPlayer == req.UserID {
		isBlack = true
	}
	gameLogRecorder.RecordGameLog(gameID, "[Game Start]"+"[GameType:"+strconv.Itoa(int(req.GameType))+"; StartTime:"+Games[gameID].StartTime+"; BlackPlayer:"+strconv.Itoa(int(Games[gameID].BlackPlayer))+"; WhitePlayer:"+strconv.Itoa(int(Games[gameID].WhitePlayer)))
	return &pb.AcceptNewGameResponse{GameID: gameID, Black: isBlack}, nil
}

func (*GameManager) Shoot(ctx context.Context, req *pb.ShootRequest) (*pb.ShootResponse, error) {
	log.Println("Received GameManager.Shoot request:", req)
	gameID, isBlack, shootRow, shootCol := req.GameID, req.Black, req.Row, req.Col
	GamesLock.Lock()
	if Games[gameID] == nil {
		GamesLock.Unlock()
		return &pb.ShootResponse{Message: "对局已不存在！"}, nil
	}
	GameWaitingReplyTimersLock.Lock()
	timer := GameWaitingReplyTimers[gameID]
	timer.Reset(playerReplyWaitingWindow)
	GameWaitingReplyTimersLock.Unlock()
	game := Games[gameID]
	if isBlack {
		gameLogRecorder.RecordGameLog(req.GameID, "[Step "+strconv.Itoa(game.Step)+"][Black Shoot][Row:"+strconv.Itoa(int(shootRow))+"; Col:"+strconv.Itoa(int(shootCol))+"]")
	} else {
		gameLogRecorder.RecordGameLog(req.GameID, "[Step "+strconv.Itoa(game.Step)+"][White Shoot][Row:"+strconv.Itoa(int(shootRow))+"; Col:"+strconv.Itoa(int(shootCol))+"]")
	}
	game.Step++
	status := 1
	if !isBlack {
		status = 2
	}
	game.Checkerboard[shootRow][shootCol] = status
	result := checkGameEnd(game.Checkerboard, int(shootRow), int(shootCol))
	if result == status {
		gameType, blackPlayer, whitePlayer, winner, loser, startTime := Games[gameID].Type, Games[gameID].BlackPlayer, Games[gameID].WhitePlayer, Games[gameID].BlackPlayer, Games[gameID].WhitePlayer, Games[gameID].StartTime
		Games[gameID] = nil
		delete(Games, gameID)
		GamesLock.Unlock()
		if !isBlack {
			winner, loser = loser, winner
		}
		endTime := time.Now().Format("2006-01-02 15:04:05")
		masterDB := mysqlClient.MasterDB
		_, err := masterDB.Exec("insert into game values(?, ?, ?, ?, ?, ?, ?, ?)", gameID, gameType, startTime, endTime, blackPlayer, whitePlayer, winner, loser)
		if err != nil {
			errorRecorder.RecordError("[gameManager][insert new game into MasterDB failed][" + err.Error() + "]")
			panic(err)
		}
		time.Sleep(100 * time.Millisecond)
		slaveDB := mysqlClient.SlaveDB
		redisDB := redisClient.Client
		var (
			userID       int
			username     string
			winCount     int
			loseCount    int
			coinAsset    int
			fightScore   int
			registerTime string
		)
		row1 := slaveDB.QueryRow("select id, username, win_count, lose_count, coin_asset, fight_score, register_time from user where id = ?", winner)
		err = row1.Scan(&userID, &username, &winCount, &loseCount, &coinAsset, &fightScore, &registerTime)
		if err != nil {
			errorRecorder.RecordError("[gameManager][get UserInfo of userID:" + strconv.Itoa(int(winner)) + " in SlaveDB failed][" + err.Error() + "]")
			panic(err)
		}
		redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(winner)), map[string]interface{}{
			"userID":        userID,
			"username":      username,
			"win_count":     winCount,
			"lose_count":    loseCount,
			"coin_asset":    coinAsset,
			"fight_score":   fightScore,
			"register_time": registerTime,
		})
		redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(int(winner)), UserInfoExpirationTime)
		row2 := slaveDB.QueryRow("select id, username, win_count, lose_count, coin_asset, fight_score, register_time from user where id = ?", loser)
		err = row2.Scan(&userID, &username, &winCount, &loseCount, &coinAsset, &fightScore, &registerTime)
		if err != nil {
			errorRecorder.RecordError("[gameManager][get UserInfo of userID:" + strconv.Itoa(int(loser)) + " in SlaveDB failed][" + err.Error() + "]")
			panic(err)
		}
		redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(loser)), map[string]interface{}{
			"userID":        userID,
			"username":      username,
			"win_count":     winCount,
			"lose_count":    loseCount,
			"coin_asset":    coinAsset,
			"fight_score":   fightScore,
			"register_time": registerTime,
		})
		redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(int(loser)), UserInfoExpirationTime)
		ShootChannelsLock.Lock()
		if isBlack {
			ShootChannels[gameID][0] <- [2]int32{checkerboardRowNumber + shootRow, checkerboardColNumber + shootCol}
		} else {
			ShootChannels[gameID][1] <- [2]int32{checkerboardRowNumber + shootRow, checkerboardColNumber + shootCol}
		}
		ShootChannels[gameID] = nil
		delete(ShootChannels, gameID)
		ShootChannelsLock.Unlock()
		gameLogRecorder.RecordGameLog(gameID, "[Game End][Type:"+strconv.Itoa(int(gameType))+"; StartTime:"+startTime+"; EndTime"+endTime+"; BlackPlayer:"+strconv.Itoa(int(blackPlayer))+"; WhitePlayer:"+strconv.Itoa(int(whitePlayer))+"; Winner:"+strconv.Itoa(int(winner))+"; Loser:"+strconv.Itoa(int(loser))+"]")
		return &pb.ShootResponse{Message: "恭喜获胜！", Row: -1, Col: -1, Result: 1}, nil
	}
	GamesLock.Unlock()
	ShootChannelsLock.Lock()
	if isBlack {
		ShootChannels[gameID][0] <- [2]int32{shootRow, shootCol}
	} else {
		ShootChannels[gameID][1] <- [2]int32{shootRow, shootCol}
	}
	ShootChannelsLock.Unlock()
	ch := ShootChannels[gameID][1]
	if !isBlack {
		ch = ShootChannels[gameID][0]
	}
	pos := <-ch
	if pos[0] >= checkerboardRowNumber && pos[1] >= checkerboardColNumber {
		return &pb.ShootResponse{Message: "很遗憾，你落败了！", Row: pos[0] - checkerboardRowNumber, Col: pos[1] - checkerboardColNumber, Result: -1}, nil
	}
	return &pb.ShootResponse{Message: "继续对局", Row: pos[0], Col: pos[1], Result: 0}, nil
}

func (*GameManager) RequireFirstShoot(ctx context.Context, req *pb.RequireFirstShootRequest) (*pb.RequireFirstShootResponse, error) {
	gameID := req.GameID
	if Games[gameID] == nil {
		return &pb.RequireFirstShootResponse{Message: "对局已不存在！"}, nil
	}
	ch := ShootChannels[gameID][0]
	pos := <-ch
	return &pb.RequireFirstShootResponse{Message: "继续对局", Row: pos[0], Col: pos[1]}, nil
}

func (*GameManager) InformTimeOut(ctx context.Context, req *pb.InformTimeOutRequest) (*pb.InformTimeOutResponse, error) {
	log.Println("Received GameManager.InformTimeOut request:", req)
	GamesLock.Lock()
	if Games[req.GameID] == nil {
		GamesLock.Unlock()
		return &pb.InformTimeOutResponse{Message: "对局已不存在！"}, nil
	}
	gameLogRecorder.RecordGameLog(req.GameID, "[User Inform TimeOut][UserID:"+strconv.Itoa(int(req.UserID))+"]")
	gameID, gameType, blackPlayer, whitePlayer, winner, loser, startTime := req.GameID, Games[req.GameID].Type, Games[req.GameID].BlackPlayer, Games[req.GameID].WhitePlayer, Games[req.GameID].BlackPlayer, Games[req.GameID].WhitePlayer, Games[req.GameID].StartTime
	if !req.Black {
		winner, loser = loser, winner
	}
	Games[gameID] = nil
	delete(Games, gameID)
	GamesLock.Unlock()
	ShootChannelsLock.Lock()
	ShootChannels[gameID] = nil
	delete(ShootChannels, gameID)
	ShootChannelsLock.Unlock()
	endTime := time.Now().Format("2006-01-02 15:04:05")
	masterDB := mysqlClient.MasterDB
	_, err := masterDB.Exec("insert into game values(?, ?, ?, ?, ?, ?, ?, ?)", gameID, gameType, startTime, endTime, blackPlayer, whitePlayer, winner, loser)
	if err != nil {
		errorRecorder.RecordError("[gameManager][insert new game into MasterDB failed][" + err.Error() + "]")
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)
	slaveDB := mysqlClient.SlaveDB
	redisDB := redisClient.Client
	var (
		userID       int
		username     string
		winCount     int
		loseCount    int
		coinAsset    int
		fightScore   int
		registerTime string
	)
	row1 := slaveDB.QueryRow("select id, username, win_count, lose_count, coin_asset, fight_score, register_time from user where id = ?", winner)
	err = row1.Scan(&userID, &username, &winCount, &loseCount, &coinAsset, &fightScore, &registerTime)
	if err != nil {
		errorRecorder.RecordError("[gameManager][get UserInfo of userID:" + strconv.Itoa(int(winner)) + " in SlaveDB failed][" + err.Error() + "]")
		panic(err)
	}
	redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(winner)), map[string]interface{}{
		"userID":        userID,
		"username":      username,
		"win_count":     winCount,
		"lose_count":    loseCount,
		"coin_asset":    coinAsset,
		"fight_score":   fightScore,
		"register_time": registerTime,
	})
	redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(int(winner)), UserInfoExpirationTime)
	row2 := slaveDB.QueryRow("select id, username, win_count, lose_count, coin_asset, fight_score, register_time from user where id = ?", loser)
	err = row2.Scan(&userID, &username, &winCount, &loseCount, &coinAsset, &fightScore, &registerTime)
	if err != nil {
		errorRecorder.RecordError("[gameManager][get UserInfo of userID:" + strconv.Itoa(int(loser)) + " in SlaveDB failed][" + err.Error() + "]")
		panic(err)
	}
	redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(loser)), map[string]interface{}{
		"userID":        userID,
		"username":      username,
		"win_count":     winCount,
		"lose_count":    loseCount,
		"coin_asset":    coinAsset,
		"fight_score":   fightScore,
		"register_time": registerTime,
	})
	redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(int(loser)), UserInfoExpirationTime)
	gameLogRecorder.RecordGameLog(gameID, "[Game End][Type:"+strconv.Itoa(int(gameType))+"; StartTime:"+startTime+"; EndTime"+endTime+"; BlackPlayer:"+strconv.Itoa(int(blackPlayer))+"; WhitePlayer:"+strconv.Itoa(int(whitePlayer))+"; Winner:"+strconv.Itoa(int(winner))+"; Loser:"+strconv.Itoa(int(loser))+"]")
	return &pb.InformTimeOutResponse{Message: "对方未在规定时间内落子，你已获胜！"}, nil
}

func createNewGame(player1, player2, gameType int32) int32 {
	redisDB := redisClient.Client
	currentGameID := 1
	CurrentGameIDLock.Lock()
	if redisDB.Exists(context.Background(), "CurrentGameID").Val() == 1 {
		var err error
		currentGameID, err = strconv.Atoi(redisDB.Get(context.Background(), "CurrentGameID").Val())
		if err != nil {
			errorRecorder.RecordError("[gameManager][get current game id from RedisDB failed][" + err.Error() + "]")
			CurrentGameIDLock.Unlock()
			panic(err)
		}
	}
	gameID := currentGameID + 1
	redisDB.Set(context.Background(), "CurrentGameID", gameID, -1)
	CurrentGameIDLock.Unlock()
	startTime := time.Now().Format("2006-01-02 15:04:05")
	blackPlayer, whitePlayer := player1, player2
	if time.Now().UnixNano()%2 == 1 {
		blackPlayer, whitePlayer = player2, player1
	}
	checkerboard := make([][]int, checkerboardRowNumber)
	for i := range checkerboard {
		checkerboard[i] = make([]int, checkerboardColNumber)
	}
	game := &Game{
		ID:           int32(gameID),
		Type:         gameType,
		StartTime:    startTime,
		BlackPlayer:  blackPlayer,
		WhitePlayer:  whitePlayer,
		Checkerboard: checkerboard,
	}
	GamesLock.Lock()
	Games[int32(gameID)] = game
	GamesLock.Unlock()
	return int32(gameID)
}

func checkGameEnd(checkerboard [][]int, i, j int) int {
	status := checkerboard[i][j]
	m, n := len(checkerboard), len(checkerboard[0])
	a, b, count := i-1, j, 1
	for a >= 0 && checkerboard[a][b] == status {
		count++
		a--
	}
	a, b = i+1, j
	for a < m && checkerboard[a][b] == status {
		count++
		a++
	}
	if count >= 5 {
		return status
	}
	a, b, count = i, j-1, 1
	for b >= 0 && checkerboard[a][b] == status {
		count++
		b--
	}
	a, b = i, j+1
	for b < n && checkerboard[a][b] == status {
		count++
		b++
	}
	if count >= 5 {
		return status
	}
	a, b, count = i-1, j-1, 1
	for a >= 0 && b >= 0 && checkerboard[a][b] == status {
		count++
		a--
		b--
	}
	a, b = i+1, j+1
	for a < m && b < n && checkerboard[a][b] == status {
		count++
		a++
		b++
	}
	if count >= 5 {
		return status
	}
	a, b, count = i-1, j+1, 1
	for a >= 0 && b < n && checkerboard[a][b] == status {
		count++
		a--
		b++
	}
	a, b = i+1, j-1
	for a < m && b >= 0 && checkerboard[a][b] == status {
		count++
		a++
		b--
	}
	if count >= 5 {
		return status
	}
	return 0
}

func waitToDeleteGame(gameID int32, timer *time.Timer) {
	<-timer.C
	GamesLock.Lock()
	if Games[gameID] == nil {
		GamesLock.Unlock()
		return
	}
	Games[gameID] = nil
	delete(Games, gameID)
	GamesLock.Unlock()
	ShootChannelsLock.Lock()
	ShootChannels[gameID] = nil
	delete(ShootChannels, gameID)
	ShootChannelsLock.Unlock()
	GameWaitingReplyTimersLock.Lock()
	GameWaitingReplyTimers[gameID] = nil
	delete(GameWaitingReplyTimers, gameID)
	GameWaitingReplyTimersLock.Unlock()
	gameLogRecorder.RecordGameLog(gameID, "[Game End With Both Players TimeOut]")
}
