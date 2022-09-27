package main

import (
	"gameManager/handler"
	pb "gameManager/proto"
	"google.golang.org/grpc"
	"net"
)

const serverAddr = "0.0.0.0:10051"

func main() {
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGameManagerServer(grpcServer, new(handler.GameManager))
	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
