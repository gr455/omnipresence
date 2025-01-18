package types

import (
	pb "github.com/gr455/omnipresence/raft/service/genproto"
)

type RaftInterface interface {
	RequestElection()
	RecvVoteRequest(candidateId string, candidatePrevLogTerm, candidatePrevLogIndex, candidateTerm, candidateLastCommitIndex int64)
	RecvVote(peerId string, granted bool, term int64)
	Append(msgs []*pb.LogEntry, leaderTerm, prevLogIdx, prevLogTerm, leaderCommit int64, leaderId string)
	RecvAppendAck(peerId string, peerTerm, matchIndex int64, success bool)
	AppendLeaderLogForCurrentTerm(msg string) (bool, bool)
	CheckLeadership() bool
}
