package main

import (
	"google.golang.org/grpc"
	"net"
	"userManager/handler"
	pb "userManager/proto"
)

const ServerAddr = "0.0.0.0:10050"

func main() {
	go handler.RefreshTopPlayers()

	lis, err := net.Listen("tcp", ServerAddr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserManagerServer(grpcServer, new(handler.UserManager))
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
