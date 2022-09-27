package handler

import "time"

const (
	UsernameToUserIDExpirationTime   = 7 * 24 * time.Hour
	UsernameToPasswordExpirationTime = 7 * 24 * time.Hour
	SessionIDToUserIDExpirationTime  = 7 * 24 * time.Hour
	UserInfoExpirationTime           = 2 * time.Hour
	//UsernameToUserIDExpirationTime   = time.Minute
	//UsernameToPasswordExpirationTime = time.Minute
	//SessionIDToUserIDExpirationTime  = time.Minute
	//UserInfoExpirationTime           = time.Minute

	sessionIDLength = 18

	topPlayersNumber      = 10
	topPlayersRefreshHour = 6

	getFreeCoinCycleTime = 24 * time.Hour
)
