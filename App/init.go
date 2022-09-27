package main

import (
	pb2 "C-Gobang/App/GameManagerProto"
	pb1 "C-Gobang/App/UserManagerProto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	userManagerClient pb1.UserManagerClient
	gameManagerClient pb2.GameManagerClient
)

func init() {
	conn1, err := grpc.Dial(userManagerServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	userManagerClient = pb1.NewUserManagerClient(conn1)

	conn2, err := grpc.Dial(gameManagerServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	gameManagerClient = pb2.NewGameManagerClient(conn2)
}
