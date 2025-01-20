package main

import (
	"fmt"
	dservice "github.com/gr455/omnipresence/datastore/kv/service"
	dspb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	"github.com/gr455/omnipresence/mq"
	rservice "github.com/gr455/omnipresence/raft/service"
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
	raftService := rservice.NewRaftServer(messageQueue)
	if raftService == nil {
		log.Fatalf("Fatal: Could not create Raft server")
		return
	}

	dsService := dservice.NewKeyValueServer(messageQueue)
	if dsService == nil {
		log.Fatalf("Fatal: Could not create datastore server")
		return
	}

	pb.RegisterRaftServer(server, raftService)
	dspb.RegisterKeyValueServer(server, dsService)

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
