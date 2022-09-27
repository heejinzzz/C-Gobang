package handler

import (
	"context"
	"strconv"
	"strings"
	"time"
	"userManager/errorRecorder"
	"userManager/mysqlClient"
	"userManager/redisClient"
)

func RefreshTopPlayers() {
	var (
		username   string
		winCount   int
		loseCount  int
		coinAsset  int
		fightScore int
	)
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), topPlayersRefreshHour, 0, 0, 0, now.Location())
		if now.Hour() >= topPlayersRefreshHour {
			next = next.Add(24 * time.Hour)
		}
		timer := time.NewTimer(next.Sub(now))
		<-timer.C
		redisDB := redisClient.Client
		slaveDB := mysqlClient.SlaveDB
		redisDB.Del(context.Background(), "TopPlayers")
		rows, err := slaveDB.Query("select username, win_count, lose_count, coin_asset, fight_score from user order by fight_score desc limit 0, ?", topPlayersNumber)
		if err != nil {
			errorRecorder.RecordError("[userManager][get top players from SlaveDB failed][" + err.Error() + "]")
			panic(err)
		}
		for rows.Next() {
			err = rows.Scan(&username, &winCount, &loseCount, &coinAsset, &fightScore)
			if err != nil {
				errorRecorder.RecordError("[userManager][get top players from SlaveDB failed][" + err.Error() + "]")
				panic(err)
			}
			userInfo := strings.Join([]string{username, strconv.Itoa(winCount), strconv.Itoa(loseCount), strconv.Itoa(coinAsset), strconv.Itoa(fightScore)}, "|")
			redisDB.RPush(context.Background(), "TopPlayers", userInfo)
		}
		time.Sleep(time.Hour)
	}
}
