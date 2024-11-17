// raftnet package handles all network communication between Raft peers
package raftnet

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// RaftNetwork defines the interface for communicating with other Raft peers.
type RaftNetwork interface {
	Broadcast_RequestForVotes(candidateId string, term int64, lastLogIndex int64, lastLogTerm int64)
	ToPeer_Vote(voteGranted bool)
	Broadcast_NewLeader(leaderId string)
	ToPeer_Append(peerId string, msgs []LogEntry, term int64, prevLogIndex int64, prevLogTerm int64, leaderCommit int64, leaderId string, wg *sync.WaitGroup)
	ToLeader_AppendAck(leaderId string, term int64, success bool, matchIndex int64)
}

// LogEntry represents a log entry in the Raft log.
type LogEntry struct {
	Entry string
	Term  int64
}

// RaftPeer defines the basic structure for a peer in the Raft network.
type RaftPeer struct {
	Id   string
	Addr string
}

// RaftNetworkImpl is the implementation of the RaftNetwork interface.
type RaftNetworkImpl struct {
	peers map[string]*RaftPeer
	mu    sync.Mutex
}

// NewRaftNetwork creates a new RaftNetworkImpl.
func NewRaftNetwork() *RaftNetworkImpl {
	return &RaftNetworkImpl{
		peers: make(map[string]*RaftPeer),
	}
}

// RegisterPeer registers a new Raft peer in the network.
func (rn *RaftNetworkImpl) RegisterPeer(id, addr string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	rn.peers[id] = &RaftPeer{
		Id:   id,
		Addr: addr,
	}
}

// GetPeer returns the address of a peer by its ID.
func (rn *RaftNetworkImpl) GetPeer(id string) (*RaftPeer, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	peer, exists := rn.peers[id]
	return peer, exists
}

// Broadcast_RequestForVotes broadcasts a vote request to all peers in the network.
func (rn *RaftNetworkImpl) Broadcast_RequestForVotes(candidateId string, term int64, lastLogIndex int64, lastLogTerm int64) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	for _, peer := range rn.peers {
		go rn.sendVoteRequest(peer, candidateId, term, lastLogIndex, lastLogTerm)
	}
}

// sendVoteRequest sends a vote request to a single peer.
func (rn *RaftNetworkImpl) sendVoteRequest(peer *RaftPeer, candidateId string, term int64, lastLogIndex int64, lastLogTerm int64) {
	// Simulate network delay and request sending.
	time.Sleep(time.Millisecond * 100)
	fmt.Printf("Sending vote request from %s to %s: Term %d, Last Log: [%d, %d]\n", candidateId, peer.Id, term, lastLogIndex, lastLogTerm)
	// Here you can call a real RPC, like gRPC, to send the request.
}

// ToPeer_Vote sends a vote decision (granted or not) to a specific peer.
func (rn *RaftNetworkImpl) ToPeer_Vote(voteGranted bool) {
	// Simulate sending vote response to peer
	// This would normally be an RPC call to the candidate.
	fmt.Printf("Vote granted: %v\n", voteGranted)
}

// Broadcast_NewLeader broadcasts a new leader announcement to all peers.
func (rn *RaftNetworkImpl) Broadcast_NewLeader(leaderId string) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	for _, peer := range rn.peers {
		go rn.sendNewLeaderAnnouncement(peer, leaderId)
	}
}

// sendNewLeaderAnnouncement sends a new leader announcement to a single peer.
func (rn *RaftNetworkImpl) sendNewLeaderAnnouncement(peer *RaftPeer, leaderId string) {
	// Simulate sending new leader notification.
	time.Sleep(time.Millisecond * 100)
	fmt.Printf("Announcing new leader: %s to peer %s\n", leaderId, peer.Id)
	// Actual RPC code to notify the peer of the new leader would go here.
}

// ToPeer_Append sends an append RPC to a specific peer.
func (rn *RaftNetworkImpl) ToPeer_Append(peerId string, msgs []LogEntry, term int64, prevLogIndex int64, prevLogTerm int64, leaderCommit int64, leaderId string, wg *sync.WaitGroup) {
	defer wg.Done()
	peer, exists := rn.GetPeer(peerId)
	if !exists {
		fmt.Printf("Peer %s not found!\n", peerId)
		return
	}
	// Simulate sending append request to peer.
	time.Sleep(time.Millisecond * 100)
	fmt.Printf("Sending append request from leader %s to %s: Term %d, Prev Log: [%d, %d], Log Entries: %v\n", leaderId, peer.Id, term, prevLogIndex, prevLogTerm, msgs)
	// Actual RPC code for sending log entries would go here.
}

// ToLeader_AppendAck sends an append acknowledgment to the leader from a peer.
func (rn *RaftNetworkImpl) ToLeader_AppendAck(leaderId string, term int64, success bool, matchIndex int64) {
	// Simulate sending append acknowledgment.
	fmt.Printf("Append ACK received by leader %s from peer: Success: %v, Match Index: %d\n", leaderId, success, matchIndex)
}
