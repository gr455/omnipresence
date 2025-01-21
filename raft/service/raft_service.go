package service

import (
	"context"
	"github.com/gr455/omnipresence/mq"
	"github.com/gr455/omnipresence/raft"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"github.com/gr455/omnipresence/raft/storage"
	"github.com/gr455/omnipresence/raft/types"
	"log"
)

type RaftServer struct {
	raft types.RaftInterface

	pb.UnimplementedRaftServer
}

func NewRaftServer(peerId string, p2pcMap map[string]pb.RaftClient, peers []string, s *storage.RaftStorage, mq *mq.MessageQueue) *RaftServer {

	r, err := raft.NewRaftConsensusObject(peerId, s, uint16(len(peers)), peers, p2pcMap, mq)
	if err != nil {
		log.Fatalf("Cannot create raft object: %s", err)
		return nil
	}

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
