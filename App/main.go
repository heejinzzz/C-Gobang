package main

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"math"
	"strconv"
)

var (
	a fyne.App
	w fyne.Window
)

type ShootPos [2]int

var (
	SessionID                 binding.String
	UserID                    binding.Int
	Username                  binding.String
	WinCount                  binding.Int
	LoseCount                 binding.Int
	CoinAsset                 binding.Int
	FightScore                binding.Int
	RegisterTime              binding.String
	AcceptGameWaitingProgress binding.Float

	ShootCountDown                   binding.Int
	OpponentShootCountDown           binding.Int
	ShootCountDownLabel              *widget.Label
	OpponentShootCountDownLabel      *widget.Label
	ShootCountDownContext            context.Context
	OpponentShootCountDownContext    context.Context
	ShootCountDownCancelFunc         context.CancelFunc
	OpponentShootCountDownCancelFunc context.CancelFunc

	Checkerboard      [][]int
	CheckerboardCells [][]*checkerboardCell

	ShootConfirmBtn *widget.Button
	ShootCancelBtn  *widget.Button
	ShootConfirmBox *fyne.Container

	OpponentNewShoot []ShootPos

	IsShootRound bool
)

func main() {
	a = app.NewWithID("com.heejinzzz.C-Gobang")
	a.Settings().SetTheme(newMyTheme())
	w = a.NewWindow("C-Gobang")

	SessionID = binding.BindPreferenceString("SessionID", a.Preferences())
	UserID = binding.NewInt()
	Username = binding.NewString()
	WinCount = binding.NewInt()
	LoseCount = binding.NewInt()
	CoinAsset = binding.NewInt()
	FightScore = binding.NewInt()
	RegisterTime = binding.NewString()
	AcceptGameWaitingProgress = binding.NewFloat()

	AcceptGameBar = widget.NewProgressBarWithData(AcceptGameWaitingProgress)
	AcceptGameBar.Max, AcceptGameBar.Min = AcceptGameWaitingSeconds, 0
	AcceptGameBar.TextFormatter = func() string {
		return "倒计时：" + strconv.Itoa(int(math.Round(AcceptGameBar.Value)))
	}
	FindNewGameDialog = dialog.NewInformation("null", "null", w)
	AcceptGameDialog = dialog.NewInformation("null", "null", w)
	WaitOpponentAcceptDialog = dialog.NewInformation("等待对方确认对局", "正在等待对方接受对局。。。", w)
	WaitOpponentAcceptDialog.SetDismissText("确定")

	ShootCountDown = binding.NewInt()
	OpponentShootCountDown = binding.NewInt()
	ShootCountDownLabel = widget.NewLabelWithData(binding.IntToString(ShootCountDown))
	OpponentShootCountDownLabel = widget.NewLabelWithData(binding.IntToString(OpponentShootCountDown))

	ShootConfirmBox = newShootConfirmBox()

	err := loginWithSessionID()
	if err != nil {
		panic(err)
	}

	w.Resize(fyne.NewSize(500, 900))
	w.ShowAndRun()
}
