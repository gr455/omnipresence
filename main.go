package main

import (
	"fmt"
	dservice "github.com/gr455/omnipresence/datastore/kv/service"
	dspb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	"github.com/gr455/omnipresence/mq"
	rservice "github.com/gr455/omnipresence/raft/service"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"github.com/gr455/omnipresence/raft/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"log"
	"net"
	"os"
	"time"
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

	peerId, exists := os.LookupEnv("RAFT_PEER_ID")
	if !exists {
		log.Fatalf("Fatal: No peer ID for the server")
		return
	}

	peers := []string{"peer1", "peer2", "peer3"}

	s := storage.NewRaftStorage(fmt.Sprintf("./raft/.log/%s/log.txt", peerId), fmt.Sprintf("./raft/.log/%s/persistent.txt", peerId))

	opts := []grpc.DialOption{
		grpc.WithInsecure(), // For demonstration purposes, use appropriate security
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay: 200 * time.Millisecond,
				MaxDelay:  200 * time.Millisecond,
			},
		}),
	}
	conn1, err := grpc.Dial("localhost:50051", opts...)
	conn2, err := grpc.Dial("localhost:50052", opts...)
	conn3, err := grpc.Dial("localhost:50053", opts...)

	c1 := pb.NewRaftClient(conn1)
	c2 := pb.NewRaftClient(conn2)
	c3 := pb.NewRaftClient(conn3)
	p2pcMap := map[string]pb.RaftClient{
		"peer1": c1,
		"peer2": c2,
		"peer3": c3,
	}

	server := grpc.NewServer()
	raftService := rservice.NewRaftServer(peerId, p2pcMap, peers, s, messageQueue)
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
