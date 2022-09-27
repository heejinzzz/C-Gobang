package main

import (
	pb2 "C-Gobang/App/GameManagerProto"
	pb1 "C-Gobang/App/UserManagerProto"
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

var (
	FindNewGameDialog        dialog.Dialog
	AcceptGameDialog         dialog.Dialog
	WaitOpponentAcceptDialog dialog.Dialog

	AcceptGameBar *widget.ProgressBar
)

func newFindNewGameDialog(gameType GameType) dialog.Dialog {
	bar := widget.NewProgressBarInfinite()
	label := widget.NewLabel("正在寻找对局。。。")
	label.Alignment = fyne.TextAlignCenter
	var title string
	if gameType == juniorGameType {
		title = "初级场"
	} else if gameType == mediumGameType {
		title = "中级场"
	} else {
		title = "高级场"
	}
	dlg := dialog.NewCustom(title, "取消", container.NewVBox(label, bar), w)
	dlg.SetOnClosed(func() {
		userID, err := UserID.Get()
		if err != nil {
			panic(err)
		}
		_, err = gameManagerClient.CancelRequireGame(context.Background(), &pb2.CancelRequireGameRequest{
			UserID:   int32(userID),
			GameType: int32(gameType),
		})
		if err != nil {
			panic(err)
		}
	})
	return dlg
}

func newAcceptGameDialog(gameType GameType, opponentID int32) dialog.Dialog {
	label := widget.NewLabel("匹配到对手，请确认对局：")
	label.Alignment = fyne.TextAlignCenter
	icon := widget.NewIcon(userAvatar)
	res, err := userManagerClient.GetUserInfo(context.Background(), &pb1.GetUserInfoRequest{UserID: opponentID})
	if err != nil {
		panic(err)
	}
	opponentName := widget.NewLabel(res.Username)
	opponentName.Alignment = fyne.TextAlignCenter
	opponentFightScore := widget.NewLabel("战力：" + strconv.Itoa(int(res.FightScore)))
	opponentFightScore.Alignment = fyne.TextAlignCenter
	ctn := container.NewVBox(label, container.NewCenter(icon), opponentName, opponentFightScore, AcceptGameBar)
	return dialog.NewCustomConfirm("确认对局", "接受", "拒绝", ctn, func(b bool) {
		if b {
			WaitOpponentAcceptDialog.Show()
			go func() {
				userID, err := UserID.Get()
				if err != nil {
					panic(err)
				}
				acceptRes, err := gameManagerClient.AcceptNewGame(context.Background(), &pb2.AcceptNewGameRequest{
					UserID:     int32(userID),
					GameType:   int32(gameType),
					OpponentID: opponentID,
				})
				if err != nil {
					panic(err)
				}
				WaitOpponentAcceptDialog.Hide()
				if acceptRes.Message != "" {
					dlg := dialog.NewInformation("对局取消", acceptRes.Message, w)
					dlg.SetDismissText("确定")
					dlg.Show()
					return
				}
				// 进入游戏界面
				Checkerboard = make([][]int, checkerboardRowNumber)
				OpponentNewShoot = []ShootPos{}
				for i := range Checkerboard {
					Checkerboard[i] = make([]int, checkerboardColNumber)
				}
				w.SetContent(newGamePage(acceptRes.GameID, gameType, opponentID, acceptRes.Black))
				if acceptRes.Black {
					dlg := dialog.NewInformation("对局开始", "对局开始，你执黑子先行！", w)
					dlg.SetDismissText("确定")
					dlg.Show()
				} else {
					dlg := dialog.NewInformation("对局开始", "对局开始，你执白子后行！", w)
					dlg.SetDismissText("确定")
					dlg.Show()
				}
			}()
		}
	}, w)
}
