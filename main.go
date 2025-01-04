package main

import (
	"fmt"
	"github.com/gr455/omnipresence/raft/service"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

func main() {
	port, exists := os.LookupEnv("RAFT_PEER_PORT")
	if !exists {
		log.Printf("No port, defaulting...")
		port = ":0"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalf("Cannot create grpc listener: %s", err)
	}

	server := grpc.NewServer()
	service := service.NewRaftServer()

	pb.RegisterRaftServer(server, service)

	log.Printf("Raft server started at port: %v\n", listener.Addr().(*net.TCPAddr).Port)
	err = server.Serve(listener)
	if err != nil {
		log.Fatalf("Cannot start raft server: %s", err)
	}

}
