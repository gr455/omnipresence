package main

import (
	"encoding/json"
	"fmt"
	dservice "github.com/gr455/omnipresence/datastore/kv/service"
	dspb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	"github.com/gr455/omnipresence/mq"
	rservice "github.com/gr455/omnipresence/raft/service"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"github.com/gr455/omnipresence/raft/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

type Config struct {
	Peers []PeerInfo `json:"peers"`
}

type PeerInfo struct {
	Id       string `json:"id"`
	Endpoint string `json:"endpoint"`
}

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

	s := storage.NewRaftStorage(fmt.Sprintf("./raft/.log/%s/log.txt", peerId), fmt.Sprintf("./raft/.log/%s/persistent.txt", peerId))

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay: 200 * time.Millisecond,
				MaxDelay:  200 * time.Millisecond,
			},
		}),
	}

	peerMap, err := ParseConfig("./config.json")
	if err != nil {
		log.Fatalf("Fatal: Could not parse config json")
		return
	}
	p2pcMap := make(map[string]pb.RaftClient)
	peers := make([]string, 0)

	for peerId, endpoint := range peerMap {
		conn, err := grpc.Dial(endpoint, opts...)
		if err != nil {
			log.Printf("Could not dial %v at %v. Skipping...", peerId, endpoint)
			continue
		}

		raftClient := pb.NewRaftClient(conn)
		p2pcMap[peerId] = raftClient
		peers = append(peers, peerId)
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

func ParseConfig(configPath string) (map[string]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	peerMap := make(map[string]string)

	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	for _, peer := range config.Peers {
		peerMap[peer.Id] = peer.Endpoint
	}

	return peerMap, nil
}
