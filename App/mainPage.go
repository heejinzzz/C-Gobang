package main

import (
	pb2 "C-Gobang/App/GameManagerProto"
	pb1 "C-Gobang/App/UserManagerProto"
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
	"time"
)

func newMainPage() *fyne.Container {
	top := canvas.NewImageFromResource(header)
	top.SetMinSize(fyne.NewSize(200, 140))
	footer := newFooter()
	sideBar := newSideBar()
	tabs := newMainPageContent()
	return container.NewBorder(top, footer, sideBar, nil, tabs)
}

func newSideBar() *widget.Card {
	avatar := canvas.NewImageFromResource(userAvatar)
	avatar.SetMinSize(fyne.NewSize(130, 130))
	name := widget.NewLabelWithData(Username)
	name.Alignment = fyne.TextAlignCenter
	ctn1 := container.NewVBox(avatar, name)

	fightScoreIcon := widget.NewIcon(fireIcon)
	fightScoreName := widget.NewLabel("战力")
	fightScore := widget.NewLabelWithData(binding.IntToString(FightScore))
	ctn2 := container.NewVBox(container.NewHBox(fightScoreIcon, fightScoreName), fightScore)

	coinAssetIcon := widget.NewIcon(coinIcon)
	coinAssetName := widget.NewLabel("棋币")
	coinAsset := widget.NewLabelWithData(binding.IntToString(CoinAsset))
	getCoinButton := widget.NewButtonWithIcon("领取每日\n免费棋币", getCoinIcon, func() {
		userID, err := UserID.Get()
		if err != nil {
			panic(err)
		}
		rsp, err := userManagerClient.UserCoinAssetChange(context.Background(), &pb1.UserCoinAssetChangeRequest{
			UserID:       int32(userID),
			ChangeAmount: getFreeCoinAmount,
		})
		if err != nil {
			dlg := dialog.NewInformation("领取失败", "连接服务器失败。Error:"+err.Error(), w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if rsp.Message != "棋币余额修改成功！" {
			dlg := dialog.NewInformation("领取失败", rsp.Message, w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		dlg := dialog.NewInformation("领取成功", "成功领取 20 棋币！", w)
		dlg.SetDismissText("确定")
		dlg.Show()
		err = fillInfo(int32(userID))
		if err != nil {
			panic(err)
		}
	})
	getCoinButton.Alignment = widget.ButtonAlignCenter
	ctn3 := container.NewVBox(container.NewHBox(coinAssetIcon, coinAssetName), coinAsset, container.NewCenter(getCoinButton))

	card := widget.NewCard("", "", container.NewVBox(ctn1, widget.NewSeparator(), ctn2, widget.NewSeparator(), ctn3))
	return card
}

func newMainPageContent() *container.AppTabs {
	tabs := container.NewAppTabs(
		newStartGameTabItem(),
		newLeaderboardTabItem(),
		newProfileTabItem(),
	)
	return tabs
}

func newStartGameTabItem() *container.TabItem {
	entry1 := newCheckerboardCell(canvas.NewImageFromResource(juniorGameEntryCover))
	entry1.SetOnTap(newGameEntryFunc(juniorGameType))
	entry2 := newCheckerboardCell(canvas.NewImageFromResource(mediumGameEntryCover))
	entry2.SetOnTap(newGameEntryFunc(mediumGameType))
	entry3 := newCheckerboardCell(canvas.NewImageFromResource(highGameEntryCover))
	entry3.SetOnTap(newGameEntryFunc(highGameType))
	ctn := container.NewGridWithRows(3, entry1, entry2, entry3)
	return container.NewTabItemWithIcon("开始游戏", startIcon, ctn)
}

func newGameEntryFunc(gameType GameType) func() {
	return func() {
		FindNewGameDialog = newFindNewGameDialog(gameType)
		FindNewGameDialog.Show()
		go popUpAcceptGameDialog(gameType)
	}
}

func popUpAcceptGameDialog(gameType GameType) {
	userID, err := UserID.Get()
	if err != nil {
		panic(err)
	}
	res, err := gameManagerClient.RequireNewGame(context.Background(), &pb2.RequireNewGameRequest{
		UserID:   int32(userID),
		GameType: int32(gameType),
	})
	if err != nil {
		panic(err)
	}
	FindNewGameDialog.Hide()
	go AcceptGameCountDown()
	AcceptGameDialog = newAcceptGameDialog(gameType, res.OpponentID)
	AcceptGameDialog.Show()
}

func AcceptGameCountDown() {
	err := AcceptGameWaitingProgress.Set(8)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(time.Second)
		timeRemained, err := AcceptGameWaitingProgress.Get()
		if err != nil {
			panic(err)
		}
		if timeRemained <= 0.01 {
			AcceptGameDialog.Hide()
			return
		}
		timeRemained -= 1
		err = AcceptGameWaitingProgress.Set(timeRemained)
		if err != nil {
			panic(err)
		}
	}
}

func newLeaderboardTabItem() *container.TabItem {
	l1 := widget.NewLabel("排名")
	l1.Wrapping = fyne.TextWrapBreak
	l2 := widget.NewLabel("玩家")
	l2.Wrapping = fyne.TextWrapBreak
	l3 := widget.NewLabel("胜场")
	l3.Wrapping = fyne.TextWrapBreak
	l4 := widget.NewLabel("败场")
	l4.Wrapping = fyne.TextWrapBreak
	l5 := widget.NewLabel("战力")
	l5.Wrapping = fyne.TextWrapBreak
	gridCtn1 := container.NewGridWithColumns(5, l1, l2, l3, l4, l5)
	rsp, err := userManagerClient.GetTopPlayers(context.Background(), &pb1.GetTopPlayersRequest{})
	if err != nil {
		label := widget.NewLabel("无法连接到服务器")
		label.Alignment = fyne.TextAlignCenter
		return container.NewTabItem("巅峰战力榜", label)
	}
	players := []*widget.Label{}
	for i, player := range rsp.TopPlayersInfo {
		playerInfo := strings.Split(player, "|")
		playerName, playerWinCount, playerLoseCount, playerFightScore := playerInfo[0], playerInfo[1], playerInfo[2], playerInfo[4]
		lb1 := widget.NewLabel(strconv.Itoa(i + 1))
		lb2 := widget.NewLabel(playerName)
		lb3 := widget.NewLabel(playerWinCount)
		lb4 := widget.NewLabel(playerLoseCount)
		lb5 := widget.NewLabel(playerFightScore)
		lb1.Wrapping, lb2.Wrapping, lb3.Wrapping, lb4.Wrapping, lb5.Wrapping = fyne.TextWrapBreak, fyne.TextWrapBreak, fyne.TextWrapBreak, fyne.TextWrapBreak, fyne.TextWrapBreak
		players = append(players, lb1, lb2, lb3, lb4, lb5)
	}
	gridCtn2 := container.NewGridWithColumns(5)
	for _, p := range players {
		gridCtn2.Add(p)
	}
	explanationLink := widget.NewHyperlink("查看巅峰战力榜说明", nil)
	explanationLink.OnTapped = func() {
		explanation := widget.NewLabel("1. 每天上午 6:00 更新巅峰战力榜。\n2. 战力说明：战力值代表了玩家的五子棋实力强弱。每次对局结束时，获胜方增加的战力值为落败方战力值的 1%，而落败方的战力值减少 20。\n3. 战力值上限为 100000，下限为 -1000。")
		explanation.Wrapping = fyne.TextWrapBreak
		dialog.NewCustom("巅峰战力榜说明", "确定", explanation, w).Show()
	}
	icon := widget.NewIcon(questionMarkIcon)
	box := container.NewHBox(icon, explanationLink)
	headCtn := container.NewVBox(box, widget.NewSeparator(), gridCtn1, widget.NewSeparator())
	ctn := container.NewBorder(headCtn, nil, nil, nil, container.NewScroll(gridCtn2))
	if len(players) == 0 {
		msg := widget.NewLabel("暂无数据")
		msg.Alignment = fyne.TextAlignCenter
		ctn.Add(msg)
	}
	card := widget.NewCard("C-Gobang 巅峰战力榜", "", ctn)
	return container.NewTabItemWithIcon("巅峰战力榜", leaderboardIcon, card)
}

func newProfileTabItem() *container.TabItem {
	form := widget.NewForm(
		widget.NewFormItem("用户ID：", widget.NewLabelWithData(binding.IntToString(UserID))),
		widget.NewFormItem("用户名：", widget.NewLabelWithData(Username)),
		widget.NewFormItem("注册时间：", widget.NewLabelWithData(RegisterTime)),
		widget.NewFormItem("胜场数：", widget.NewLabelWithData(binding.IntToString(WinCount))),
		widget.NewFormItem("败场数：", widget.NewLabelWithData(binding.IntToString(LoseCount))),
		widget.NewFormItem("战力值：", widget.NewLabelWithData(binding.IntToString(FightScore))),
		widget.NewFormItem("棋币余额：", widget.NewLabelWithData(binding.IntToString(CoinAsset))),
	)
	logoutButton := widget.NewButtonWithIcon("退出登录", theme.LogoutIcon(), func() {
		err := SessionID.Set("")
		if err != nil {
			panic(err)
		}
		w.SetContent(newLoginPage())
	})
	logoutButton.Importance = widget.HighImportance
	logoutButton.Alignment = widget.ButtonAlignCenter
	ctn := container.NewVBox(form, container.NewCenter(logoutButton))
	tabItem := container.NewTabItemWithIcon("个人信息", profileIcon, container.NewScroll(ctn))
	return tabItem
}
