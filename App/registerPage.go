package main

import (
	pb "C-Gobang/App/UserManagerProto"
	"context"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newRegisterPage() *fyne.Container {
	top := canvas.NewImageFromResource(header)
	top.SetMinSize(fyne.NewSize(200, 140))
	footer := newFooter()
	card := newRegisterCard()
	ctn := container.NewBorder(top, footer, nil, nil, card)
	return ctn
}

func newRegisterCard() *widget.Card {
	toolbar := widget.NewToolbar(widget.NewToolbarAction(returnIcon, func() {
		w.SetContent(newLoginPage())
	}))
	returnHint := widget.NewHyperlink("返回登录页面", nil)
	returnHint.OnTapped = func() {
		w.SetContent(newLoginPage())
	}
	returnBox := container.NewHBox(toolbar, returnHint)
	usernameEntry := widget.NewEntry()
	usernameEntry.PlaceHolder = "设置用户名"
	usernameEntry.Validator = func(s string) error {
		if len(s) < 6 {
			return errors.New("用户名过短")
		}
		if len(s) > 24 {
			return errors.New("用户名过长")
		}
		return nil
	}
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "设置密码"
	passwordEntry.Validator = func(s string) error {
		if len(s) < 6 {
			return errors.New("密码过短")
		}
		if len(s) > 12 {
			return errors.New("密码过长")
		}
		return nil
	}
	passwordConfirmEntry := widget.NewPasswordEntry()
	passwordConfirmEntry.PlaceHolder = "再输入一次密码"
	passwordConfirmEntry.Validator = func(s string) error {
		if s != passwordEntry.Text {
			return errors.New("确认密码与设置密码不一致")
		}
		return nil
	}
	form := widget.NewForm(
		widget.NewFormItem("用户名：", usernameEntry),
		widget.NewFormItem("密码：", passwordEntry),
		widget.NewFormItem("确认密码：", passwordConfirmEntry),
	)
	check := widget.NewCheck("我已阅读并同意", func(b bool) {})
	link := widget.NewHyperlink("《用户协议》", nil)
	link.OnTapped = func() {
		dialog.NewInformation("用户协议", "暂无", w).Show()
	}
	ctn1 := container.NewHBox(check, link)
	registerButton := widget.NewButtonWithIcon("注册", theme.DocumentIcon(), func() {
		if usernameEntry.Text == "" {
			dlg := dialog.NewInformation("注册失败", "请输入用户名", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if len(usernameEntry.Text) < 6 {
			dlg := dialog.NewInformation("注册失败", "用户名过短", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if len(usernameEntry.Text) > 24 {
			dlg := dialog.NewInformation("注册失败", "用户名过长", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if passwordEntry.Text == "" {
			dlg := dialog.NewInformation("注册失败", "请输入密码", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if len(passwordEntry.Text) < 6 {
			dlg := dialog.NewInformation("注册失败", "密码设置过短", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if len(passwordEntry.Text) > 12 {
			dlg := dialog.NewInformation("注册失败", "密码设置过长", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if passwordConfirmEntry.Text != passwordEntry.Text {
			dlg := dialog.NewInformation("注册失败", "确认密码与设置密码不一致", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if !check.Checked {
			dlg := dialog.NewInformation("注册失败", "请勾选”我已阅读并同意《用户协议》“", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		rsp, err := userManagerClient.UserRegister(context.Background(), &pb.UserRegisterRequest{
			Username: usernameEntry.Text,
			Password: passwordEntry.Text,
		})
		if err != nil {
			dlg := dialog.NewInformation("注册失败", "注册失败，请重试", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		if rsp.Message != "注册成功！" {
			dlg := dialog.NewInformation("注册失败", rsp.Message, w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		_ = SessionID.Set(rsp.SessionID)
		err = fillInfo(rsp.UserID)
		if err != nil {
			dlg := dialog.NewInformation("注册失败", "注册失败，请重试", w)
			dlg.SetDismissText("确定")
			dlg.Show()
			return
		}
		// 进入主页
		w.SetContent(newMainPage())
	})
	registerButton.Alignment = widget.ButtonAlignCenter
	registerButton.Importance = widget.HighImportance
	ctn2 := container.NewVBox(ctn1, registerButton)
	ctn3 := container.NewBorder(nil, ctn2, nil, nil, container.NewVBox(form, widget.NewSeparator(), returnBox))
	card := widget.NewCard("注册新用户", "", ctn3)
	return card
}
