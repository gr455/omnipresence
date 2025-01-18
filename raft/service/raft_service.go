package service

import (
	"context"
	"fmt"
	"github.com/gr455/omnipresence/raft"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"github.com/gr455/omnipresence/raft/storage"
	"github.com/gr455/omnipresence/raft/types"
	"google.golang.org/grpc"
	"log"
	"os"
)

type RaftServer struct {
	raft types.RaftInterface

	pb.UnimplementedRaftServer
}

func NewRaftServer() *RaftServer {
	// Test code
	peerId, exists := os.LookupEnv("RAFT_PEER_ID")
	if !exists {
		log.Fatalf("No peer ID for the server")
		return nil
	}

	peers := []string{"peer1", "peer2", "peer3"}

	s := storage.NewRaftStorage(fmt.Sprintf("./raft/.log/%s/log.txt", peerId), fmt.Sprintf("./raft/.log/%s/persistent.txt", peerId))

	opts := []grpc.DialOption{
		grpc.WithInsecure(), // For demonstration purposes, use appropriate security
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

	r, err := raft.NewRaftConsensusObject(peerId, s, 1, peers, p2pcMap)
	if err != nil {
		log.Fatalf("Cannot create raft object: %s", err)
		return nil
	}

	// END of Test code

	return &RaftServer{raft: r}
}

func (s *RaftServer) RequestVote(ctx context.Context, request *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	log.Printf("[RPC Served] RequestVote: %v\n", request)
	go s.raft.RecvVoteRequest(request.CandidateId, request.LastLogTerm, request.LastLogIndex, request.Term, request.LastCommitIndex)
	return &pb.RequestVoteResponse{}, nil
}
func (s *RaftServer) Vote(ctx context.Context, request *pb.VoteRequest) (*pb.VoteResponse, error) {
	log.Printf("[RPC Served] Vote: %v\n", request)
	go s.raft.RecvVote(request.PeerId, request.VoteGranted, request.Term)
	return &pb.VoteResponse{}, nil
}
func (s *RaftServer) AppendEntries(ctx context.Context, request *pb.AppendRequest) (*pb.AppendResponse, error) {
	log.Printf("[RPC Served] AppendEntries: %v\n", request)
	go s.raft.Append(request.Entries, request.Term, request.PrevLogIndex, request.PrevLogTerm, request.LeaderCommit, request.LeaderId)
	return &pb.AppendResponse{}, nil
}
func (s *RaftServer) AckAppend(ctx context.Context, request *pb.AckAppendRequest) (*pb.AckAppendResponse, error) {
	log.Printf("[RPC Served] AckAppend: %v\n", request)
	go s.raft.RecvAppendAck(request.PeerId, request.Term, request.MatchIndex, request.Success)
	return &pb.AckAppendResponse{}, nil
}
func (s *RaftServer) AppendLeaderLog(ctx context.Context, request *pb.AppendLeaderLogRequest) (*pb.AppendLeaderLogResponse, error) {
	log.Printf("[RPC Served] AppendLeaderLog: %v\n", request)
	isLeader, ok := s.raft.AppendLeaderLogForCurrentTerm(request.Msg)
	return &pb.AppendLeaderLogResponse{IsLeader: isLeader, Ok: ok}, nil
}
func (s *RaftServer) CheckLeadership(ctx context.Context, request *pb.CheckLeadershipRequest) (*pb.CheckLeadershipResponse, error) {
	return &pb.CheckLeadershipResponse{IsLeader: s.raft.CheckLeadership()}, nil
}
