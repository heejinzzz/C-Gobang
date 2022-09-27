package handler

import (
	"context"
	"strconv"
	"time"
	"userManager/errorRecorder"
	"userManager/mysqlClient"
	"userManager/redisClient"
	"userManager/sessionIdGenerator"
	"userManager/userLogRecorder"

	log "go-micro.dev/v4/logger"

	pb "userManager/proto"
)

type UserManager struct{}

func (e *UserManager) UserRegister(ctx context.Context, req *pb.UserRegisterRequest) (*pb.UserRegisterResponse, error) {
	log.Infof("Received UserManager.UserRegister request: %v", req)
	username, password := req.Username, req.Password
	slaveDB := mysqlClient.SlaveDB
	rows, err := slaveDB.Query("select id from user where username = ?", username)
	if err != nil {
		errorRecorder.RecordError("[userManager][Mysql SlaveDB query failed][" + err.Error() + "]")
		panic(err)
	}
	if rows.Next() {
		return &pb.UserRegisterResponse{Message: "用户名已存在，请更换用户名重试！"}, nil
	}
	redisDB := redisClient.Client
	for !redisDB.SetNX(context.Background(), "userTable", 1, -1).Val() {
		time.Sleep(100 * time.Millisecond)
	}
	var maxId int
	row := slaveDB.QueryRow("select ifnull(max_id, 0) from (select max(id) max_id from user) u")
	err = row.Scan(&maxId)
	if err != nil {
		errorRecorder.RecordError("[userManager][get current max userId failed][" + err.Error() + "]")
		panic(err)
	}
	userID := maxId + 1
	registerTime := time.Now().Format("2006-01-02 15:04:05")
	masterDB := mysqlClient.MasterDB
	_, err = masterDB.Exec("insert into user values(" + strconv.Itoa(userID) + ", '" + username + "', '" + password + "', 0, 0, 1000, 1000, '" + registerTime + "')")
	if err != nil {
		errorRecorder.RecordError("[userManager][insert new user into User Table failed][" + err.Error() + "]")
		redisDB.Del(context.Background(), "userTable")
		return &pb.UserRegisterResponse{Message: "注册失败，请重试!"}, nil
	}
	redisDB.Del(context.Background(), "userTable")
	sessionID := sessionIdGenerator.NewSessionId(sessionIDLength)
	rsp := &pb.UserRegisterResponse{
		Message:   "注册成功！",
		UserID:    int32(userID),
		SessionID: sessionID,
	}
	userLogRecorder.RecordUserLog("[user register][UserID:" + strconv.Itoa(userID) + "; Username:" + username + "; Password:" + password + "]")
	redisDB.Set(context.Background(), "UsernameToUserID-"+username, userID, UsernameToUserIDExpirationTime)
	redisDB.Set(context.Background(), "UsernameToPassword-"+username, password, UsernameToPasswordExpirationTime)
	redisDB.Set(context.Background(), "SessionIDToUserID-"+sessionID, userID, SessionIDToUserIDExpirationTime)
	redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(userID), map[string]interface{}{
		"userID":        userID,
		"username":      username,
		"win_count":     0,
		"lose_count":    0,
		"coin_asset":    1000,
		"fight_score":   1000,
		"register_time": registerTime,
	})
	redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(userID), UserInfoExpirationTime)
	return rsp, nil
}

func (*UserManager) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	log.Infof("Received UserManager.UserLogin request: %v", req)
	redisDB := redisClient.Client
	slaveDB := mysqlClient.SlaveDB
	if redisDB.Exists(context.Background(), "UsernameToPassword-"+req.Username).Val() == 1 {
		password := redisDB.Get(context.Background(), "UsernameToPassword-"+req.Username).Val()
		if password != req.Password {
			return &pb.UserLoginResponse{Message: "密码错误！"}, nil
		} else if redisDB.Exists(context.Background(), "UsernameToUserID-"+req.Username).Val() == 1 {
			userID, err := strconv.Atoi(redisDB.Get(context.Background(), "UsernameToUserID-"+req.Username).Val())
			if err != nil {
				errorRecorder.RecordError("[userManager][convert userID from string to int failed][" + err.Error() + "]")
				panic(err)
			}
			sessionID := sessionIdGenerator.NewSessionId(sessionIDLength)
			rsp := &pb.UserLoginResponse{
				Message:   "登录成功！",
				UserID:    int32(userID),
				SessionID: sessionID,
			}
			redisDB.Set(context.Background(), "SessionIDToUserID-"+sessionID, userID, SessionIDToUserIDExpirationTime)
			userLogRecorder.RecordUserLog("[user login with username and password][UserID:" + strconv.Itoa(userID) + "; Username:" + req.Username + "; Password: " + req.Password + "]")
			return rsp, nil
		} else {
			var userID int
			row := slaveDB.QueryRow("select id from user where username = ?", req.Username)
			err := row.Scan(&userID)
			if err != nil {
				errorRecorder.RecordError("[userManager][get userID from username in SlaveDB failed][" + err.Error() + "]")
				panic(err)
			}
			sessionID := sessionIdGenerator.NewSessionId(sessionIDLength)
			rsp := &pb.UserLoginResponse{
				Message:   "登录成功！",
				UserID:    int32(userID),
				SessionID: sessionID,
			}
			redisDB.Set(context.Background(), "SessionIDToUserID-"+sessionID, userID, SessionIDToUserIDExpirationTime)
			redisDB.Set(context.Background(), "UsernameToUserID-"+req.Username, userID, UsernameToUserIDExpirationTime)
			userLogRecorder.RecordUserLog("[user login with username and password][UserID:" + strconv.Itoa(userID) + "; Username:" + req.Username + "; Password: " + req.Password + "]")
			return rsp, nil
		}
	}
	var password string
	row := slaveDB.QueryRow("select password from user where username = ?", req.Username)
	err := row.Scan(&password)
	if err != nil {
		return &pb.UserLoginResponse{Message: "用户名不存在！"}, nil
	}
	redisDB.Set(context.Background(), "UsernameToPassword-"+req.Username, password, UsernameToPasswordExpirationTime)
	if password != req.Password {
		return &pb.UserLoginResponse{Message: "输入的密码错误！"}, nil
	}
	var userID int
	if redisDB.Exists(context.Background(), "UsernameToUserID-"+req.Username).Val() == 1 {
		userID, err = strconv.Atoi(redisDB.Get(context.Background(), "UsernameToUserID-"+req.Username).Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert userID from string to int failed][" + err.Error() + "]")
			panic(err)
		}
	} else {
		rowForUserID := slaveDB.QueryRow("select id from user where username = ?", req.Username)
		err = rowForUserID.Scan(&userID)
		if err != nil {
			errorRecorder.RecordError("[userManager][get userID from username in SlaveDB failed][" + err.Error() + "]")
			panic(err)
		}
		redisDB.Set(context.Background(), "UsernameToUserID-"+req.Username, userID, UsernameToUserIDExpirationTime)
	}
	sessionID := sessionIdGenerator.NewSessionId(sessionIDLength)
	redisDB.Set(context.Background(), "SessionIDToUserID-"+sessionID, userID, SessionIDToUserIDExpirationTime)
	rsp := &pb.UserLoginResponse{
		Message:   "登录成功！",
		UserID:    int32(userID),
		SessionID: sessionID,
	}
	userLogRecorder.RecordUserLog("[user login with username and password][UserID:" + strconv.Itoa(userID) + "; Username:" + req.Username + "; Password: " + req.Password + "]")
	return rsp, nil
}

func (*UserManager) UserLoginWithSessionID(ctx context.Context, req *pb.UserLoginWithSessionIDRequest) (*pb.UserLoginWithSessionIDResponse, error) {
	log.Infof("Received UserManager.UserLoginWithSessionID request: %v", req)
	redisDB := redisClient.Client
	if redisDB.Exists(context.Background(), "SessionIDToUserID-"+req.SessionID).Val() != 1 {
		return &pb.UserLoginWithSessionIDResponse{Message: "登录信息已过期，请重新登录！"}, nil
	}
	userID, err := strconv.Atoi(redisDB.Get(context.Background(), "SessionIDToUserID-"+req.SessionID).Val())
	if err != nil {
		errorRecorder.RecordError("[userManager][convert userID from string to int failed][" + err.Error() + "]")
		panic(err)
	}
	rsp := &pb.UserLoginWithSessionIDResponse{
		Message: "登录成功！",
		UserID:  int32(userID),
	}
	userLogRecorder.RecordUserLog("[user login with sessionID][UserID:" + strconv.Itoa(userID) + "; SessionID:" + req.SessionID + "]")
	return rsp, nil
}

func (*UserManager) UserCoinAssetChange(ctx context.Context, req *pb.UserCoinAssetChangeRequest) (*pb.UserCoinAssetChangeResponse, error) {
	log.Infof("Received UserManager.UserCoinAssetChange request: %v", req)
	redisDB := redisClient.Client
	masterDB := mysqlClient.MasterDB
	if !redisDB.SetNX(context.Background(), "GetFreeCoin-"+strconv.Itoa(int(req.UserID)), 1, getFreeCoinCycleTime).Val() {
		return &pb.UserCoinAssetChangeResponse{Message: "你已经领取过每日免费棋币，请24小时后再来领取！"}, nil
	}
	var coinAsset int
	if redisDB.Exists(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID))).Val() == 1 {
		var err error
		coinAsset, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "coin_asset").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert CoinAsset from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "coin_asset", coinAsset+int(req.ChangeAmount))
		_, err = masterDB.Exec("update user set coin_asset = ? where id = ?", coinAsset+int(req.ChangeAmount), req.UserID)
		if err != nil {
			errorRecorder.RecordError("[userManager][change CoinAsset of userID:" + strconv.Itoa(int(req.UserID)) + " in MasterDB failed][" + err.Error() + "]")
			panic(err)
		}
	} else {
		_, err := masterDB.Exec("update user set coin_asset = coin_asset + ? where id = ?", int(req.ChangeAmount), req.UserID)
		if err != nil {
			errorRecorder.RecordError("[userManager][change CoinAsset of userID:" + strconv.Itoa(int(req.UserID)) + " in MasterDB failed][" + err.Error() + "]")
			panic(err)
		}
	}
	userLogRecorder.RecordUserLog("[user CoinAsset change][UserID:" + strconv.Itoa(int(req.UserID)) + "; ChangeAmount:" + strconv.Itoa(int(req.ChangeAmount)) + "]")
	return &pb.UserCoinAssetChangeResponse{Message: "棋币余额修改成功！"}, nil
}

func (*UserManager) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	log.Infof("Received UserManager.GetUserInfo request: %v", req)
	redisDB := redisClient.Client
	slaveDB := mysqlClient.SlaveDB
	var (
		userID       int
		username     string
		winCount     int
		loseCount    int
		coinAsset    int
		fightScore   int
		registerTime string
	)
	if redisDB.Exists(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID))).Val() == 1 {
		var err error
		userID, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "userID").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert userID from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		username = redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "username").Val()
		winCount, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "win_count").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert WinCount from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		loseCount, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "lose_count").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert LoseCount from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		coinAsset, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "coin_asset").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert CoinAsset from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		fightScore, err = strconv.Atoi(redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "fight_score").Val())
		if err != nil {
			errorRecorder.RecordError("[userManager][convert FightScore from string to int failed][" + err.Error() + "]")
			panic(err)
		}
		registerTime = redisDB.HGet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), "register_time").Val()
	} else {
		row := slaveDB.QueryRow("select id, username, win_count, lose_count, coin_asset, fight_score, register_time from user where id = ?", req.UserID)
		err := row.Scan(&userID, &username, &winCount, &loseCount, &coinAsset, &fightScore, &registerTime)
		if err != nil {
			errorRecorder.RecordError("[userManager][get UserInfo of userID:" + strconv.Itoa(int(req.UserID)) + " in SlaveDB failed][" + err.Error() + "]")
			panic(err)
		}
		redisDB.HSet(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), map[string]interface{}{
			"userID":        userID,
			"username":      username,
			"win_count":     winCount,
			"lose_count":    loseCount,
			"coin_asset":    coinAsset,
			"fight_score":   fightScore,
			"register_time": registerTime,
		})
		redisDB.Expire(context.Background(), "UserInfo-"+strconv.Itoa(int(req.UserID)), UserInfoExpirationTime)
	}
	rsp := &pb.GetUserInfoResponse{
		UserID:       int32(userID),
		Username:     username,
		WinCount:     int32(winCount),
		LoseCount:    int32(loseCount),
		CoinAsset:    int32(coinAsset),
		FightScore:   int32(fightScore),
		RegisterTime: registerTime,
	}
	return rsp, nil
}

func (*UserManager) GetTopPlayers(ctx context.Context, req *pb.GetTopPlayersRequest) (*pb.GetTopPlayersResponse, error) {
	log.Infof("Received UserManager.GetTopPlayers request: %v", req)
	redisDB := redisClient.Client
	if redisDB.Exists(context.Background(), "TopPlayers").Val() != 1 {
		return &pb.GetTopPlayersResponse{Message: "暂无数据!"}, nil
	}
	topPlayers := redisDB.LRange(context.Background(), "TopPlayers", 0, -1).Val()
	rsp := &pb.GetTopPlayersResponse{
		Message:        "获取成功！",
		TopPlayersInfo: topPlayers,
	}
	return rsp, nil
}
