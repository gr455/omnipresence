package main

import (
	"context"
	"fmt"
	dspb "github.com/gr455/omnipresence/datastore/kv/service/genproto"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(), // For demonstration purposes, use appropriate security
	}
	conn1, _ := grpc.Dial("localhost:50051", opts...)
	conn2, _ := grpc.Dial("localhost:50052", opts...)
	conn3, _ := grpc.Dial("localhost:50053", opts...)
	rc1 := pb.NewRaftClient(conn1)
	rc2 := pb.NewRaftClient(conn2)
	rc3 := pb.NewRaftClient(conn3)
	rclients := []pb.RaftClient{rc1, rc2, rc3}

	dc1 := dspb.NewKeyValueClient(conn1)
	dc2 := dspb.NewKeyValueClient(conn2)
	dc3 := dspb.NewKeyValueClient(conn3)
	dclients := []dspb.KeyValueClient{dc1, dc2, dc3}

	for {
		var option int
		fmt.Printf("(1) Get, (2) Put\n")
		fmt.Scan(&option)

		if option == 1 {
			var key string
			var cid int
			fmt.Printf("Key: ")
			fmt.Scan(&key)
			fmt.Printf("Client: ")
			fmt.Scan(&cid)

			fmt.Println(read(key, dclients[cid-1], cid))
		} else {
			var key string
			var val string
			fmt.Printf("Key: ")
			fmt.Scan(&key)
			fmt.Printf("Val: ")
			fmt.Scan(&val)

			sendWrite(key, val, rclients)
		}
	}

}

func sendWrite(key, val string, clients []pb.RaftClient) {
	request := &pb.AppendLeaderLogRequest{Msg: fmt.Sprintf("%v %v", key, val)}
	leader := -2

	for idx, client := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res, err := client.AppendLeaderLog(ctx, request)
		if err != nil {
			log.Printf("Err: Client could not send write request to %v: %v\n", idx, err)
			continue
		}

		if res.IsLeader {
			if leader != -2 {
				panic("multiple leaders detected")
			}
			leader = idx + 1
		}
	}

	log.Printf("Leader: %v has written to log\n", leader)

}

func read(key string, client dspb.KeyValueClient, clientId int) string {
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
