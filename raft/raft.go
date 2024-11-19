/**
 * This is an implementation of the Raft consensus algorithm.
 *
 * This implementation follows the design as outlined in the paper:
 * "In Search of an Understandable Consensus Algorithm" (https://raft.github.io/raft.pdf).
 */

package raft

import (
	"fmt"
	"github.com/gr455/omnipresence/raft/raftnet"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	raftstorage "github.com/gr455/omnipresence/raft/storage"
	"github.com/gr455/omnipresence/raft/utils"
	"math"
	"math/rand"
	"sync"
)

type RaftPeerState int

const (
	RAFT_PEER_STATE_UNKNOWN = iota
	RAFT_PEER_STATE_LEADER
	RAFT_PEER_STATE_CANDIDATE
	RAFT_PEER_STATE_FOLLOWER
)

type RaftPeerRuntimeConstants struct {
	ElectionTimeoutMillis uint16
	HeartbeatTimeMillis   uint16
	// Id of this peer
	PeerIdentifer string
	// Count of peers in the raft cluster
	PeerCount uint16

	// List of all peers in the cluster
	AllPeers []string
}

type LeaderMeta struct {
	NextIndex  map[string]int64
	MatchIndex map[string]int64
}

type CandidateMeta struct {
	VoteGranted map[string]bool
}

type RaftMutex struct {
	// TODO: Have granular mutexes
	GlobalDataMutex *sync.RWMutex
}

type RaftConsensusObject struct {
	// State of the peer - Leader, follower, candidate
	PeerState RaftPeerState
	// Log term of the peer
	Term int64
	// Index of the last log commit of this peer
	LastCommitIndex int64
	// Index of the last applied commit of this peer
	LastAppliedIndex int64
	// For this peer's Term, who did they vote in the election
	TermVotePeerId string
	// Number of votes this peer got in the current term as a candidate
	CurrentTermVotes uint16
	// Timer ticks to zero, peer starts election with self as a candidate
	ElectionTimer *utils.Timer
	// Timer ticks to zero as a leader, peer sends out a heartbeat
	HeartbeatTimer *utils.Timer
	Log            []pb.LogEntry
	LogStartIdx    int64

	Storage *raftstorage.RaftStorage
	Network *raftnet.RaftNetwork

	RaftMutex
	LeaderMeta
	CandidateMeta
	RaftPeerRuntimeConstants
}

func NewRaftConsensusObject(id string, storage *raftstorage.RaftStorage, pcount uint16, peers []string) (*RaftConsensusObject, error) {
	r := &RaftConsensusObject{}
	return r.Initialize(id, storage, pcount, peers)
}

// Initialize does not lock anything. Make sure that no callbacks start running before init completes.
func (raft *RaftConsensusObject) Initialize(id string, storage *raftstorage.RaftStorage, pcount uint16, peers []string) (*RaftConsensusObject, error) {
	raft.ElectionTimeoutMillis = uint16(rand.Intn(15_000)) + 5_000
	raft.HeartbeatTimeMillis = 2000
	raft.PeerIdentifer = id
	raft.PeerCount = pcount
	raft.Storage = storage
	raft.AllPeers = peers

	raft.ElectionTimer = utils.NewTimer(raft.ElectionTimeoutMillis, raft.RequestElection)
	// raft.HeartbeatTimer = timer.NewTimer(time.HeartbeatTimeMillis, raft.SendAppends)

	// temp code for file loads
	log, startIdx, err := raft.Storage.ReadLog()
	if err != nil {
		fmt.Printf("Fatal: Error reading log - %v\n", err)
		return nil, err
	}

	raft.LogStartIdx = startIdx

	for _, logEntry := range log {
		raft.Log = append(raft.Log, pb.LogEntry(logEntry))
	}

	fmt.Println(raft.Log, raft.LogStartIdx)
	fmt.Printf("Initialized %v\n", id)

	return raft, nil
}

// Send out broadcast requests to all the peers asking them to vote this peer.
func (raft *RaftConsensusObject) RequestElection() {
	fmt.Printf("INFO: RequestElection called on %v for term %v\n", raft.PeerIdentifer, raft.Term+1)
	raft.setCurrentTerm(raft.Term + 1)
	raft.setPeerState(RAFT_PEER_STATE_CANDIDATE)
	raft.ElectionTimer.Restart() // Maybe re-requets election after another electiontimeout

	raft.CurrentTermVotes = 1
	raft.VoteGranted = make(map[string]bool)

	lastLogIdx := raft.LogStartIdx + int64(len(raft.Log)) - 1
	lastLogTerm := int64(-1)

	if lastLogIdx != -1 {
		lastLogTerm = raft.Log[lastLogIdx].Term
	}
	_ = lastLogTerm

	raft.Network.Broadcast_RequestForVotes(raft.PeerIdentifer, raft.Term, lastLogIdx, lastLogTerm, raft.LastCommitIndex)
}

// Recieve a request for vote, decide whether to vote
func (raft *RaftConsensusObject) RecvVoteRequest(candidateId string, candidatePrevLogTerm, candidatePrevLogIndex, candidateTerm, candidateLastCommitIndex int64) {
	vote := raft.DecideVote(candidateId, candidatePrevLogTerm, candidatePrevLogIndex, candidateTerm, candidateLastCommitIndex)

	raft.Network.ToPeer_Vote(candidateId, raft.PeerIdentifer, vote, raft.Term)
}

// Not an RPC action
func (raft *RaftConsensusObject) WinElection() {
	fmt.Printf("INFO: WinElection called by %v for term %v\n", raft.PeerIdentifer, raft.Term)
	raft.ElectionTimer.Stop()
	raft.setPeerState(RAFT_PEER_STATE_LEADER)

	raft.HeartbeatTimer.Restart()
}

// Recv a vote, increment currentTermVotes, and win election if majority
func (raft *RaftConsensusObject) RecvVote(peerId string, granted bool, term int64) {
	if raft.PeerState != RAFT_PEER_STATE_CANDIDATE || term != raft.Term {
		return
	}

	raft.GlobalDataMutex.Lock()

	wasGranted, ok := raft.VoteGranted[peerId]
	if (!ok || !wasGranted) && granted {
		raft.CurrentTermVotes++
	}

	raft.VoteGranted[peerId] = granted

	raft.GlobalDataMutex.Unlock()

	if raft.CurrentTermVotes > uint16(math.Ceil(float64(raft.PeerCount)/2.0)) {
		raft.WinElection()
	}
}

// ANY Rpc () that you get, you must check if the term of that sender is greater than self, then immediately update your term
// If the sending peer is candidate, you must update your term and vote yes to the candidate
// If you are a candidate, you must immedately demote to follower if this happens.

// Decide if current peer should vote true or false to a vote request by another peer. Only returns vote decision.
func (raft *RaftConsensusObject) DecideVote(candidateId string, candidatePrevLogTerm, candidatePrevLogIndex, candidateTerm, candidateLastCommitIndex int64) bool {
	// check if you have already voted someone else this term.
	if raft.Term > candidateTerm || (raft.TermVotePeerId != "" && raft.TermVotePeerId != candidateId) {
		return false
	}

	// vote false if either voter has more logs, or voter's last log term is higher than candidate's.
	raft.GlobalDataMutex.RLock()
	if raft.LogStartIdx+int64(len(raft.Log)) > candidatePrevLogIndex+1 ||
		(raft.LogStartIdx+int64(len(raft.Log)) != 0 && raft.Log[len(raft.Log)-1].Term > candidatePrevLogTerm) {
		return false
	}
	raft.GlobalDataMutex.RUnlock()
	// Vote false if peer's commit index is higher than candidate's.
	if raft.LastCommitIndex > candidateLastCommitIndex {
		return false
	}

	return true
}

// Check append conditions and append msgs[] starting at prevLogIdx + 1. Also ack the leader once appended
// Ack true if could append, else ack false.
// msgs should be empty for heartbeats or commits.

// Note that prevLogIdx is the logIdx after which messages are being appended. Peer might have a higher value, that is fine
// This higher value cannot have been committed though.
func (raft *RaftConsensusObject) Append(msgs []pb.LogEntry, leaderTerm, prevLogIdx, prevLogTerm, leaderCommit int64, leaderId string) {
	if raft.PeerState != RAFT_PEER_STATE_FOLLOWER {
		fmt.Printf("INFO: Demoted %v from %v to follower", raft.PeerIdentifer, raft.PeerState)
		raft.PeerState = RAFT_PEER_STATE_FOLLOWER
	}

	raft.GlobalDataMutex.Lock()

	fmt.Printf("INFO: Append called by %v, for term %v\n", raft.PeerIdentifer, raft.Term)
	success := true

	_ = raft.checkAndUpdateTerm(leaderTerm)

	// TODO: even when snapshotting, keep one extra log index so that this check can be done
	// the check being: term of the log at prevLogIdx should be the same on peer and leader.
	if leaderTerm < raft.Term ||
		(raft.LogStartIdx+int64(len(raft.Log)-1) < prevLogIdx) ||
		(prevLogIdx != -1 && raft.Log[prevLogIdx-raft.LogStartIdx].Term != prevLogTerm) {

		success = false
	}

	matchIndex := int64(0)
	if success {
		for i, newLog := range msgs {
			if prevLogIdx+1+int64(i) < raft.LogStartIdx+int64(len(raft.Log)) {
				raft.Log[prevLogIdx+1+int64(i)-raft.LogStartIdx].Term = newLog.Term
				raft.Log[prevLogIdx+1+int64(i)-raft.LogStartIdx].Entry = newLog.Entry
			} else {
				raft.Log = append(raft.Log, pb.LogEntry{Term: newLog.Term, Entry: newLog.Entry})
			}
		}

		// Discard rest of the log after overwriting
		raft.Log = raft.Log[:prevLogIdx+int64(len(msgs))+1]
		matchIndex = prevLogIdx + int64(len(msgs))

		// CHECK: only commit (even earlier logs) if append was successful.
		// Yes this is okay.
		// If at log index, term is equal. All logs before that index are equal.
		// Therefore safe to commit till leaderCommit.
		if raft.LastCommitIndex < leaderCommit {
			raft.Commit(leaderCommit, leaderTerm)
		}
	}

	raft.GlobalDataMutex.Unlock()
	raft.Network.ToLeader_AppendAck(leaderId, raft.PeerIdentifer, raft.Term, success, matchIndex)
}

// Deviation from paper: Also send matchIndex, to know replication status.
func (raft *RaftConsensusObject) RecvAppendAck(peerId string, peerTerm, matchIndex int64, success bool) {
	if raft.PeerState != RAFT_PEER_STATE_LEADER {
		return
	}

	// If self term as a leader is smaller than a peer's, step down.
	if !raft.checkAndUpdateTerm(peerTerm) {
		raft.setPeerState(RAFT_PEER_STATE_FOLLOWER)
		raft.HeartbeatTimer.Stop()
		return
	}

	raft.GlobalDataMutex.Lock()

	fmt.Printf("INFO: Append ack recvd by %v, ack from %v with success: %d\n", raft.PeerIdentifer, peerId, success)

	raft.MatchIndex[peerId] = matchIndex
	if !success {
		raft.NextIndex[peerId]--
		return
	}

	raft.NextIndex[peerId] = matchIndex + 1

	raft.GlobalDataMutex.Unlock()
	raft.GlobalDataMutex.RLock()
	// Commit till the latest maximum match
	maxMatch := raft.getMaximumMatch()
	raft.GlobalDataMutex.RUnlock()
	raft.Commit(maxMatch, raft.Term)
}

// Send append RPCs to all peers
func (raft *RaftConsensusObject) SendAppends() {
	if raft.PeerState != RAFT_PEER_STATE_LEADER {
		return
	}

	var wg sync.WaitGroup

	for _, peerId := range raft.AllPeers {
		var msgs []*pb.LogEntry
		nextIndex := int64(0)
		mapNextIdx, ok := raft.NextIndex[peerId]

		if ok {
			nextIndex = mapNextIdx
		}

		for i := int64(nextIndex - raft.LogStartIdx); i < int64(len(raft.Log)); i++ {
			msgs = append(msgs, &raft.Log[i])
		}

		wg.Add(1)
		// TODO: Should release waitgroup
		go raft.Network.ToPeer_Append(peerId, msgs, raft.Term, nextIndex-1, raft.Log[nextIndex-1].Term, raft.LastCommitIndex, raft.PeerIdentifer, &wg)
	}

	wg.Wait()
	fmt.Printf("INFO: Done sending appends\n")
}

// NOT TOP LEVEL, DO NOT USE GLOBAL MUTEX
// Write the current state of the log to disk. Not an RPC action, called from Append.
func (raft *RaftConsensusObject) Commit(tillIdx, term int64) {
	for raft.LastCommitIndex < tillIdx {
		// Don't attempt to commit past your own log.
		if raft.LastCommitIndex > int64(len(raft.Log))+raft.LogStartIdx-1 {
			raft.LastCommitIndex = int64(len(raft.Log)) + raft.LogStartIdx - 1
			break
		}

		// TODO: potentially push to a queue
		err := raft.Storage.WriteToLog(raft.Log[raft.LastCommitIndex+1-raft.LogStartIdx].Entry, term, raft.LastCommitIndex+1)
		if err != nil {
			fmt.Printf("Err: Could not commit to log - %v\n", err)
			return
		}

		raft.LastCommitIndex++
	}

}

// Returns the maximum index of log which atleast the majority peers have replicated
func (raft *RaftConsensusObject) getMaximumMatch() int64 {
	for i := int64(len(raft.Log)) + raft.LogStartIdx - 1; i >= raft.LogStartIdx; i-- {
		var thisMatchCount int64 = 0
		for _, idx := range raft.MatchIndex {
			if idx >= i {
				thisMatchCount++
			}
		}

		if thisMatchCount >= int64(math.Ceil(float64(raft.PeerCount)/2.0)) {
			return i
		}
	}

	return raft.LogStartIdx - 1
}

// Check if incoming request has a greater term. If so, update current term.
// Also update curr term voted for peerid.
// Returns whether the term was up to date to begin with.
func (raft *RaftConsensusObject) checkAndUpdateTerm(peerTerm int64) bool {
	if peerTerm > raft.Term {
		raft.setCurrentTerm(peerTerm)
		raft.setCurrentTermVotedFor("")
		return false
	}

	return true
}

func (raft *RaftConsensusObject) setCurrentTerm(term int64) {
	raft.Term = term
	go raft.Storage.WritePersistent(raft.Term, raft.TermVotePeerId)
}

func (raft *RaftConsensusObject) setCurrentTermVotedFor(peerId string) {
	raft.TermVotePeerId = peerId
	go raft.Storage.WritePersistent(raft.Term, raft.TermVotePeerId)
}

func (raft *RaftConsensusObject) setPeerState(state RaftPeerState) {
	raft.PeerState = state
}
