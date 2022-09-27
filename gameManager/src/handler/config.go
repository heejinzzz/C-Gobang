package handler

import (
	"sync"
	"time"
)

type GameRoom map[int32]chan int32

type Game struct {
	ID           int32
	Type         int32
	StartTime    string
	EndTime      string
	BlackPlayer  int32
	WhitePlayer  int32
	Winner       int32
	Loser        int32
	Checkerboard [][]int
	Step         int
}

type ShootPos [2]int32

var (
	JuniorGameRoom = GameRoom{}
	MediumGameRoom = GameRoom{}
	HighGameRoom   = GameRoom{}

	JuniorGameRoomLock sync.Mutex
	MediumGameRoomLock sync.Mutex
	HighGameRoomLock   sync.Mutex

	WaitingPlayers     = map[int32]chan int32{}
	WaitingPlayersLock sync.Mutex

	Games     = map[int32]*Game{}
	GamesLock sync.Mutex

	GameWaitingReplyTimers     = map[int32]*time.Timer{}
	GameWaitingReplyTimersLock sync.Mutex

	ShootChannels     = map[int32][]chan ShootPos{}
	ShootChannelsLock sync.Mutex

	CurrentGameIDLock sync.Mutex
)

const (
	checkerboardRowNumber = 15
	checkerboardColNumber = 15

	playerAcceptGameWaitingWindow = 12 * time.Second
	playerReplyWaitingWindow      = 60 * time.Second

	UserInfoExpirationTime = 2 * time.Hour
)
