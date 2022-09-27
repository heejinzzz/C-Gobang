package main

import (
	pb2 "C-Gobang/App/GameManagerProto"
	pb1 "C-Gobang/App/UserManagerProto"
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"time"
)

func newGamePage(gameID int32, gameType GameType, opponentID int32, isBlack bool) *fyne.Container {
	header := newGameHeader(opponentID, !isBlack)
	footer := newGameFooter(isBlack)
	checkerboard := newCheckerboard(gameID, gameType, isBlack)
	ctn := container.NewBorder(header, footer, nil, nil, container.NewBorder(nil, ShootConfirmBox, nil, nil, checkerboard))
	ShootConfirmBox.Hide()
	ShootCountDownContext, ShootCountDownCancelFunc = context.WithCancel(context.Background())
	OpponentShootCountDownContext, OpponentShootCountDownCancelFunc = context.WithCancel(context.Background())
	if isBlack {
		IsShootRound = true
		go shootCountDownGoroutine()
	} else {
		IsShootRound = false
		go waitFirstShoot(gameID)
		go opponentShootCountDownGoroutine(gameID, isBlack)
	}
	return ctn
}

func newGameHeader(opponentID int32, isBlack bool) *widget.Card {
	opponent, err := userManagerClient.GetUserInfo(context.Background(), &pb1.GetUserInfoRequest{UserID: opponentID})
	if err != nil {
		panic(err)
	}
	name := widget.NewLabel(opponent.Username)
	fightScore := widget.NewLabel("战力：" + strconv.Itoa(int(opponent.FightScore)))
	avatar := canvas.NewImageFromResource(userAvatar)
	avatar.SetMinSize(fyne.NewSize(80, 80))
	statusIcon := widget.NewIcon(blackCellDark)
	if !isBlack {
		statusIcon = widget.NewIcon(whiteCellDark)
	}
	nameBox := container.NewHBox(name, statusIcon)
	ctn := container.NewBorder(nil, nil, avatar, OpponentShootCountDownLabel, container.NewVBox(nameBox, fightScore))
	return widget.NewCard("", "", ctn)
}

func newGameFooter(isBlack bool) *widget.Card {
	name, err := Username.Get()
	if err != nil {
		panic(err)
	}
	nameLabel := widget.NewLabel(name)
	fightScore, err := FightScore.Get()
	if err != nil {
		panic(err)
	}
	fightScoreLabel := widget.NewLabel("战力：" + strconv.Itoa(fightScore))
	avatar := canvas.NewImageFromResource(userAvatar)
	avatar.SetMinSize(fyne.NewSize(80, 80))
	statusIcon := widget.NewIcon(blackCellDark)
	if !isBlack {
		statusIcon = widget.NewIcon(whiteCellDark)
	}
	nameBox := container.NewHBox(nameLabel, statusIcon)
	ctn := container.NewBorder(nil, nil, avatar, ShootCountDownLabel, container.NewVBox(nameBox, fightScoreLabel))
	return widget.NewCard("", "", ctn)
}

func newCheckerboard(gameID int32, gameType GameType, isBlack bool) *fyne.Container {
	CheckerboardCells = make([][]*checkerboardCell, checkerboardRowNumber)
	for i := range CheckerboardCells {
		CheckerboardCells[i] = make([]*checkerboardCell, checkerboardColNumber)
	}
	for i := 0; i < checkerboardRowNumber; i++ {
		for j := 0; j < checkerboardColNumber; j++ {
			cellImg := canvas.NewImageFromResource(spaceCellDark)
			if (i+j)%2 == 1 {
				cellImg = canvas.NewImageFromResource(spaceCellLight)
			}
			cell := newCheckerboardCell(cellImg)
			CheckerboardCells[i][j] = cell
			cell.SetOnTap(newOnTapFunc(gameID, gameType, i, j, isBlack))
		}
	}
	ctn := container.New(newCheckerboardLayout(checkerboardColNumber))
	for _, row := range CheckerboardCells {
		for _, v := range row {
			ctn.Add(v)
		}
	}
	return ctn
}

func newOnTapFunc(gameID int32, gameType GameType, i, j int, isBlack bool) func() {
	return func() {
		if !IsShootRound || Checkerboard[i][j] != 0 {
			return
		}
		IsShootRound = false
		if len(OpponentNewShoot) > 0 {
			pos := OpponentNewShoot[0]
			if (pos[0]+pos[1])%2 == 0 {
				if isBlack {
					CheckerboardCells[pos[0]][pos[1]].Img.Resource = whiteCellDark
				} else {
					CheckerboardCells[pos[0]][pos[1]].Img.Resource = blackCellDark
				}
			} else {
				if isBlack {
					CheckerboardCells[pos[0]][pos[1]].Img.Resource = whiteCellLight
				} else {
					CheckerboardCells[pos[0]][pos[1]].Img.Resource = blackCellLight
				}
			}
			CheckerboardCells[pos[0]][pos[1]].Refresh()
			OpponentNewShoot = []ShootPos{}
		}
		if (i+j)%2 == 0 {
			if isBlack {
				CheckerboardCells[i][j].Img.Resource = confirmBlackCellDark
			} else {
				CheckerboardCells[i][j].Img.Resource = confirmWhiteCellDark
			}
		} else {
			if isBlack {
				CheckerboardCells[i][j].Img.Resource = confirmBlackCellLight
			} else {
				CheckerboardCells[i][j].Img.Resource = confirmWhiteCellLight
			}
		}
		CheckerboardCells[i][j].Refresh()
		ShootConfirmBtn.OnTapped = func() {
			ShootConfirmBox.Hide()
			if (i+j)%2 == 0 {
				if isBlack {
					CheckerboardCells[i][j].Img.Resource = blackCellDark
				} else {
					CheckerboardCells[i][j].Img.Resource = whiteCellDark
				}
			} else {
				if isBlack {
					CheckerboardCells[i][j].Img.Resource = blackCellLight
				} else {
					CheckerboardCells[i][j].Img.Resource = whiteCellLight
				}
			}
			CheckerboardCells[i][j].Refresh()
			go shoot(gameID, gameType, i, j, isBlack)
		}
		ShootCancelBtn.OnTapped = func() {
			ShootConfirmBox.Hide()
			if (i+j)%2 == 0 {
				CheckerboardCells[i][j].Img.Resource = spaceCellDark
			} else {
				CheckerboardCells[i][j].Img.Resource = spaceCellLight
			}
			CheckerboardCells[i][j].Refresh()
			IsShootRound = true
		}
		ShootConfirmBox.Show()
	}
}

func shoot(gameID int32, gameType GameType, row, col int, isBlack bool) {
	Checkerboard[row][col] = 1
	if !isBlack {
		Checkerboard[row][col] = 2
	}
	userID, err := UserID.Get()
	if err != nil {
		panic(err)
	}
	ShootCountDownCancelFunc()
	OpponentShootCountDownContext, OpponentShootCountDownCancelFunc = context.WithCancel(context.Background())
	go opponentShootCountDownGoroutine(gameID, isBlack)
	res, err := gameManagerClient.Shoot(context.Background(), &pb2.ShootRequest{
		GameID: gameID,
		UserID: int32(userID),
		Black:  isBlack,
		Row:    int32(row),
		Col:    int32(col),
	})
	if err != nil {
		panic(err)
	}
	OpponentShootCountDownCancelFunc()
	if res.Message == "对局已不存在！" {
		err = fillInfo(int32(userID))
		if err != nil {
			panic(err)
		}
		dlg := dialog.NewInformation("对局结束", "落子超时，对局已经结束！", w)
		dlg.SetDismissText("确定")
		dlg.SetOnClosed(func() {
			w.SetContent(newMainPage())
		})
		dlg.Show()
		return
	}
	if res.Result == 1 {
		err = fillInfo(int32(userID))
		if err != nil {
			panic(err)
		}
		coinIncreaseAmount := 10
		if gameType == mediumGameType {
			coinIncreaseAmount = 20
		} else if gameType == highGameType {
			coinIncreaseAmount = 50
		}
		dlg := dialog.NewInformation("胜利", "恭喜获胜！获得："+strconv.Itoa(coinIncreaseAmount)+"棋币。", w)
		dlg.SetDismissText("确定")
		dlg.SetOnClosed(func() {
			w.SetContent(newMainPage())
		})
		dlg.Show()
		return
	}
	i, j := int(res.Row), int(res.Col)
	Checkerboard[i][j] = 1
	if isBlack {
		Checkerboard[i][j] = 2
	}
	if (i+j)%2 == 0 {
		if isBlack {
			CheckerboardCells[i][j].Img.Resource = newWhiteCellDark
		} else {
			CheckerboardCells[i][j].Img.Resource = newBlackCellDark
		}
	} else {
		if isBlack {
			CheckerboardCells[i][j].Img.Resource = newWhiteCellLight
		} else {
			CheckerboardCells[i][j].Img.Resource = newBlackCellLight
		}
	}
	CheckerboardCells[i][j].Refresh()
	OpponentNewShoot = append(OpponentNewShoot, ShootPos{i, j})
	if res.Result == -1 {
		err = fillInfo(int32(userID))
		if err != nil {
			panic(err)
		}
		coinReduceAmount := 10
		if gameType == mediumGameType {
			coinReduceAmount = 20
		} else if gameType == highGameType {
			coinReduceAmount = 50
		}
		dlg := dialog.NewInformation("落败", "很遗憾，你落败了！失去："+strconv.Itoa(coinReduceAmount)+"棋币。", w)
		dlg.SetDismissText("确定")
		dlg.SetOnClosed(func() {
			w.SetContent(newMainPage())
		})
		dlg.Show()
		return
	}
	IsShootRound = true
	ShootCountDownContext, ShootCountDownCancelFunc = context.WithCancel(context.Background())
	go shootCountDownGoroutine()
}

func waitFirstShoot(gameID int32) {
	res, err := gameManagerClient.RequireFirstShoot(context.Background(), &pb2.RequireFirstShootRequest{GameID: gameID})
	if err != nil {
		panic(err)
	}
	OpponentShootCountDownCancelFunc()
	if res.Message == "对局已不存在！" {
		dlg := dialog.NewInformation("对局结束", "落子超时，对局已经结束！", w)
		dlg.SetDismissText("确定")
		dlg.SetOnClosed(func() {
			w.SetContent(newMainPage())
		})
		dlg.Show()
		return
	}
	i, j := int(res.Row), int(res.Col)
	Checkerboard[i][j] = 1
	if (i+j)%2 == 0 {
		CheckerboardCells[i][j].Img.Resource = newBlackCellDark
	} else {
		CheckerboardCells[i][j].Img.Resource = newBlackCellLight
	}
	CheckerboardCells[i][j].Refresh()
	OpponentNewShoot = append(OpponentNewShoot, ShootPos{i, j})
	IsShootRound = true
	ShootCountDownContext, ShootCountDownCancelFunc = context.WithCancel(context.Background())
	go shootCountDownGoroutine()
}

func shootCountDownGoroutine() {
	OpponentShootCountDownLabel.Hide()
	err := ShootCountDown.Set(shootWaitingSeconds)
	if err != nil {
		panic(err)
	}
	ShootCountDownLabel.Show()
	for {
		select {
		case <-ShootCountDownContext.Done():
			return
		default:
			time.Sleep(time.Second)
			t, err := ShootCountDown.Get()
			if err != nil {
				panic(err)
			}
			if t == 0 {
				dlg := dialog.NewInformation("落败", "你未在规定时间内落子，视为认输！", w)
				dlg.SetDismissText("确定")
				dlg.SetOnClosed(func() {
					w.SetContent(newMainPage())
				})
				dlg.Show()
				go func() {
					time.Sleep(4 * time.Second)
					userID, err := UserID.Get()
					if err != nil {
						panic(err)
					}
					err = fillInfo(int32(userID))
					if err != nil {
						panic(err)
					}
				}()
				return
			}
			err = ShootCountDown.Set(t - 1)
			if err != nil {
				panic(err)
			}
		}
	}
}

func opponentShootCountDownGoroutine(gameID int32, isBlack bool) {
	ShootCountDownLabel.Hide()
	err := OpponentShootCountDown.Set(shootWaitingSeconds)
	if err != nil {
		panic(err)
	}
	OpponentShootCountDownLabel.Show()
	for {
		select {
		case <-OpponentShootCountDownContext.Done():
			return
		default:
			time.Sleep(time.Second)
			t, err := OpponentShootCountDown.Get()
			if err != nil {
				panic(err)
			}
			if t == 0 {
				time.Sleep(2 * time.Second)
				select {
				case <-OpponentShootCountDownContext.Done():
					return
				default:
					userID, err := UserID.Get()
					if err != nil {
						panic(err)
					}
					res, err := gameManagerClient.InformTimeOut(context.Background(), &pb2.InformTimeOutRequest{
						GameID: gameID,
						UserID: int32(userID),
						Black:  isBlack,
					})
					if err != nil {
						panic(err)
					}
					dlg := dialog.NewInformation("对局结束", res.Message, w)
					dlg.SetDismissText("确定")
					dlg.SetOnClosed(func() {
						w.SetContent(newMainPage())
					})
					dlg.Show()
					err = fillInfo(int32(userID))
					if err != nil {
						panic(err)
					}
					return
				}
			}
			err = OpponentShootCountDown.Set(t - 1)
			if err != nil {
				panic(err)
			}
		}
	}
}

func newShootConfirmBox() *fyne.Container {
	label := widget.NewLabel("确定落子？")
	label.Alignment = fyne.TextAlignCenter
	ShootConfirmBtn = widget.NewButtonWithIcon("确定", theme.ConfirmIcon(), func() {})
	ShootConfirmBtn.Importance = widget.HighImportance
	ShootCancelBtn = widget.NewButtonWithIcon("取消", theme.CancelIcon(), func() {})
	gridCtn := container.NewGridWithColumns(2, ShootConfirmBtn, ShootCancelBtn)
	return container.NewVBox(label, gridCtn)
}
