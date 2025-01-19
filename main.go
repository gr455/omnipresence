package main

import (
	"fmt"
	"github.com/gr455/omnipresence/mq"
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
		log.Printf("No port given, one will be picked...")
		port = "0"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalf("Cannot create grpc listener: %s", err)
		return
	}

	messageQueue := mq.Initialize( /*bufferSize=*/ 100)

	server := grpc.NewServer()
	raftService := service.NewRaftServer(messageQueue)

	pb.RegisterRaftServer(server, raftService)

	log.Printf("Raft server started at port: %v\n", listener.Addr().(*net.TCPAddr).Port)

	go func() {
		err = server.Serve(listener)
		if err != nil {
			log.Fatalf("Cannot start raft server: %s", err)
			return
		}
	}()

	// Sleep forever
	select {}
}
