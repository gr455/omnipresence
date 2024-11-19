package service

import (
	"context"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"github.com/gr455/omnipresence/raft/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// "log"
)

type RaftServer struct {
	raft *types.RaftInterface

	pb.UnimplementedRaftServer
}

func (s *RaftServer) RequestVote(context.Context, *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}
func (s *RaftServer) Vote(context.Context, *pb.VoteRequest) (*pb.VoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Vote not implemented")
}
func (s *RaftServer) AppendEntries(context.Context, *pb.AppendRequest) (*pb.AppendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (s *RaftServer) AckAppend(context.Context, *pb.AckAppendRequest) (*pb.AckAppendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AckAppend not implemented")
}
