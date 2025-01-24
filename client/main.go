package main

import (
	"context"
	"encoding/json"
	"fmt"
	dspb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
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
	opts := []grpc.DialOption{
		grpc.WithInsecure(), // For demonstration purposes, use appropriate security
	}

	peerMap, err := ParseConfig("./config.json")
	if err != nil {
		log.Fatalf("Fatal: Could not parse config json")
		return
	}

	p2prcMap := make(map[string]pb.RaftClient)
	p2pdcMap := make(map[string]dspb.KeyValueClient)

	for peerId, endpoint := range peerMap {
		conn, err := grpc.Dial(endpoint, opts...)
		if err != nil {
			log.Printf("Could not dial %v at %v. Skipping...", peerId, endpoint)
			continue
		}

		raftClient := pb.NewRaftClient(conn)
		dataClient := dspb.NewKeyValueClient(conn)
		p2prcMap[peerId] = raftClient
		p2pdcMap[peerId] = dataClient
		log.Printf("Registered peer: %v\n", peerId)
	}

	for {
		var option int
		fmt.Printf("(1) Get, (2) Put\n(input): ")
		fmt.Scan(&option)

		if option == 1 {
			var key string
			var cid string
			fmt.Printf("(input) Key: ")
			fmt.Scan(&key)
			fmt.Printf("(input) Client: ")
			fmt.Scan(&cid)

			if _, ok := p2pdcMap[cid]; ok {
				fmt.Println(read(key, p2pdcMap[cid], cid))
			} else {
				log.Printf("No such client\n")
			}
		} else {
			var key string
			var val string
			fmt.Printf("(input) Key: ")
			fmt.Scan(&key)
			fmt.Printf("(input) Val: ")
			fmt.Scan(&val)

			sendWrite(key, val, p2prcMap)
		}
	}

}

func sendWrite(key, val string, clients map[string]pb.RaftClient) {
	request := &pb.AppendLeaderLogRequest{Msg: fmt.Sprintf("%v %v", key, val)}
	leader := ""

	for peerId, client := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, err := client.AppendLeaderLog(ctx, request)
		if err != nil {
			log.Printf("Err: Client could not send write request to %v: %v\n", peerId, err)
			continue
		}

		if res.IsLeader {
			if leader != "" {
				panic("multiple leaders detected")
			}
			leader = peerId
		}
	}

	log.Printf("Leader: %v has written to log\n", leader)

}

func read(key string, client dspb.KeyValueClient, clientId string) string {
	request := &dspb.GetRequest{Key: key}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := client.Get(ctx, request)
	if err != nil {
		log.Printf("Err: Client could not send get request to %v: %v\n", clientId, err)
		return ""
	}

	return res.Value
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
