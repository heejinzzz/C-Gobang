package main

import (
	pb "C-Gobang/App/UserManagerProto"
	"context"
	"fyne.io/fyne/v2/dialog"
)

func fillInfo(userID int32) error {
	rsp, err := userManagerClient.GetUserInfo(context.Background(), &pb.GetUserInfoRequest{UserID: userID})
	if err != nil {
		return err
	}
	err = UserID.Set(int(rsp.UserID))
	if err != nil {
		return err
	}
	err = Username.Set(rsp.Username)
	if err != nil {
		return err
	}
	err = WinCount.Set(int(rsp.WinCount))
	if err != nil {
		return err
	}
	err = LoseCount.Set(int(rsp.LoseCount))
	if err != nil {
		return err
	}
	err = CoinAsset.Set(int(rsp.CoinAsset))
	if err != nil {
		return err
	}
	err = FightScore.Set(int(rsp.FightScore))
	if err != nil {
		return err
	}
	err = RegisterTime.Set(rsp.RegisterTime)
	if err != nil {
		return err
	}
	return nil
}

func loginWithSessionID() error {
	sessionID, err := SessionID.Get()
	if err != nil {
		return err
	}
	if sessionID == "" {
		w.SetContent(newLoginPage())
		return nil
	}
	rsp, err := userManagerClient.UserLoginWithSessionID(context.Background(), &pb.UserLoginWithSessionIDRequest{SessionID: sessionID})
	if err != nil {
		return err
	}
	if rsp.Message != "登录成功！" {
		w.SetContent(newLoginPage())
		dlg := dialog.NewInformation("登录失败", rsp.Message, w)
		dlg.SetDismissText("确定")
		dlg.Show()
		return nil
	}
	err = fillInfo(rsp.UserID)
	if err != nil {
		return err
	}
	// 进入主页
	w.SetContent(newMainPage())
	return nil
}
