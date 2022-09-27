package main

import (
	pb "C-Gobang/App/UserManagerProto"
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

func newLoginPage() *fyne.Container {
	top := canvas.NewImageFromResource(header)
	top.SetMinSize(fyne.NewSize(200, 140))
	footer := newFooter()
	card := newLoginCard()
	ctn := container.NewBorder(top, footer, nil, nil, card)
	return ctn
}

func newLoginCard() *widget.Card {
	usernameEntry := widget.NewEntry()
	usernameEntry.PlaceHolder = "输入用户名"
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "输入密码"
	form := widget.NewForm(
		widget.NewFormItem("用户名：", usernameEntry),
		widget.NewFormItem("密码：", passwordEntry),
	)

	loginButton := widget.NewButtonWithIcon("登录", theme.LoginIcon(), func() {
		if usernameEntry.Text == "" {
			dlg := dialog.NewInformation("登录失败", "请输入用户名", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if passwordEntry.Text == "" {
			dlg := dialog.NewInformation("登录失败", "请输入密码", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		rsp, err := userManagerClient.UserLogin(context.Background(), &pb.UserLoginRequest{
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
		})
		if err != nil {
			dlg := dialog.NewInformation("登录失败", "登录失败，请重试", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if rsp.Message != "登录成功！" {
			dlg := dialog.NewInformation("登录失败", rsp.Message, w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		err = SessionID.Set(rsp.SessionID)
		if err != nil {
			dialog.NewError(err, w).Show()
			return
		}
		err = fillInfo(rsp.UserID)
		if err != nil {
			dialog.NewError(err, w).Show()
			return
		}
		// 进入主页
		w.SetContent(newMainPage())
	})
	loginButton.Alignment = widget.ButtonAlignCenter
	loginButton.Importance = widget.HighImportance
	registerButton := widget.NewButtonWithIcon("注册", theme.DocumentIcon(), func() {
		w.SetContent(newRegisterPage())
	})
	registerButton.Alignment = widget.ButtonAlignCenter
	gridContainer := container.NewGridWithColumns(2, loginButton, registerButton)

	registerHint1 := canvas.NewText("还没有用户？", color.NRGBA{R: 48, G: 138, B: 255, A: 180})
	registerHint1.TextSize = 12
	registerHint1.Alignment = fyne.TextAlignTrailing

	registerHint2 := canvas.NewText("点击下方注册按钮立即注册新用户！", color.NRGBA{R: 48, G: 138, B: 255, A: 180})
	registerHint2.TextSize = 12
	registerHint2.Alignment = fyne.TextAlignTrailing

	card := widget.NewCard("登录", "", container.NewBorder(nil, gridContainer, nil, nil, container.NewVBox(form, widget.NewSeparator(), registerHint1, registerHint2)))
	return card
}
