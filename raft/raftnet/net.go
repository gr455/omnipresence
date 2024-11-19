// raftnet package handles all network communication between Raft peers
package raftnet

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"log"
	"sync"
	"time"
)

// RaftNetwork defines the interface for communicating with other Raft peers.
type RaftNetwork struct {
	PeerToPeerClient map[string]pb.RaftClient
	mu               sync.Mutex
}

// NewRaftNetwork creates a new RaftNetwork.
func NewRaftNetwork() *RaftNetwork {
	return &RaftNetwork{
		PeerToPeerClient: make(map[string]pb.RaftClient),
	}
}

// RegisterPeer registers a new Raft peer in the network. Replaces peer client if it exists.
func (rn *RaftNetwork) RegisterPeer(id string, client pb.RaftClient) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.PeerToPeerClient[id] = client
}

// GetPeer returns the address of a peer by its ID.
func (rn *RaftNetwork) GetPeerClient(id string) (pb.RaftClient, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	peer, exists := rn.PeerToPeerClient[id]
	return peer, exists
}

// Broadcast_RequestForVotes broadcasts a vote request to all peers in the network.
func (rn *RaftNetwork) Broadcast_RequestForVotes(candidateId string, term, lastLogIndex, lastLogTerm, lastCommit int64) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	for _, peer := range rn.PeerToPeerClient {
		go rn.sendVoteRequest(peer, candidateId, term, lastLogIndex, lastLogTerm)
	}
	return nil
}

// sendVoteRequest sends a vote request to a single peer.
func (rn *RaftNetwork) sendVoteRequest(peerClient pb.RaftClient, candidateId string, term int64, lastLogIndex int64, lastLogTerm int64) {
	// Simulate network delay and request sending.
	time.Sleep(time.Millisecond * 100)
	// Here you can call a real RPC, like gRPC, to send the request.
}

// ToPeer_Vote sends a vote decision (granted or not) to a specific peer.
func (rn *RaftNetwork) ToPeer_Vote(candidateId, voterId string, voteGranted bool, term int64) error {
	candidateClient, exists := rn.GetPeerClient(candidateId)
	if !exists {
		return errors.New("Peer is unknown\n")
	}

	voteRequest := &pb.VoteRequest{
		PeerId:      voterId,
		VoteGranted: voteGranted,
		Term:        term,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := candidateClient.Vote(ctx, voteRequest)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not send vote request: %v\n", err))
	}

	log.Printf("Vote granted = %v for candidate %s\n", voteGranted, candidateId)
	return nil
}

// ToPeer_Append sends an append RPC to a specific peer. Releases wg
func (rn *RaftNetwork) ToPeer_Append(peerId string, msgs []*pb.LogEntry, term int64, prevLogIndex int64, prevLogTerm int64, leaderCommit int64, leaderId string, wg *sync.WaitGroup) error {
	peerClient, exists := rn.GetPeerClient(peerId)
	if !exists {
		return errors.New("Peer is unknown\n")
	}

	appendRequest := &pb.AppendRequest{
		LeaderId:     leaderId,
		Term:         term,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		LeaderCommit: leaderCommit,
		Entries:      msgs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := peerClient.AppendEntries(ctx, appendRequest)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not send append request: %v\n", err))
	}

	log.Printf("Sent append to: %s", peerId)

	wg.Done()
	return nil
}

// ToLeader_AppendAck sends an append acknowledgment to the leader from a peer.
func (rn *RaftNetwork) ToLeader_AppendAck(leaderId, peerId string, term int64, success bool, matchIndex int64) error {
	peerClient, exists := rn.GetPeerClient(leaderId)
	if !exists {
		return errors.New("Peer is unknown\n")
	}

	ackAppendRequest := &pb.AckAppendRequest{
		PeerId:     peerId,
		Success:    success,
		Term:       term,
		MatchIndex: matchIndex,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := peerClient.AckAppend(ctx, ackAppendRequest)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not send append ack: %v\n", err))
	}

	log.Printf("Sent append ack to: %s", peerId)

	return nil
}
