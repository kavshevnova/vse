package main

// Задача: Raft Consensus — leader election, log replication, snapshots, membership.

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type NodeState string

const (
	StateFollower  NodeState = "follower"
	StateCandidate NodeState = "candidate"
	StateLeader    NodeState = "leader"
)

type LogEntry struct {
	Index     int64
	Term      int64
	Command   interface{}
	Timestamp time.Time
}

type Snapshot struct {
	LastIncludedIndex int64
	LastIncludedTerm  int64
	Data              []byte
	Timestamp         time.Time
}

type VoteRequest struct {
	Term         int64
	CandidateID  string
	LastLogIndex int64
	LastLogTerm  int64
}

type VoteResponse struct {
	Term        int64
	VoteGranted bool
}

type AppendEntriesRequest struct {
	Term         int64
	LeaderID     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []LogEntry
	LeaderCommit int64
}

type AppendEntriesResponse struct {
	Term          int64
	Success       bool
	ConflictTerm  int64
	ConflictIndex int64
}

type SnapshotRequest struct {
	Term              int64
	LeaderID          string
	LastIncludedIndex int64
	LastIncludedTerm  int64
	Offset            int64
	Data              []byte
	Done              bool
}

type SnapshotResponse struct{ Term int64 }

type MemberStatus string

const (
	MemberStatusAlive   MemberStatus = "alive"
	MemberStatusSuspect MemberStatus = "suspect"
	MemberStatusDead    MemberStatus = "dead"
)

type MemberRole string

const (
	MemberRoleVoter    MemberRole = "voter"
	MemberRoleNonVoter MemberRole = "non_voter"
	MemberRoleLearner  MemberRole = "learner"
)

type Member struct {
	ID       string
	Address  string
	Status   MemberStatus
	Role     MemberRole
	JoinedAt time.Time
}

type ConsensusConfig struct {
	NodeID              string
	ElectionTimeoutMin  time.Duration
	ElectionTimeoutMax  time.Duration
	HeartbeatInterval   time.Duration
	SnapshotInterval    int64
	MaxEntriesPerAppend int
	MaxSnapshotSize     int64
}

// --- InMemoryLogStore ---

type InMemoryLogStore struct {
	mu      sync.RWMutex
	entries []LogEntry
}

func (s *InMemoryLogStore) Append(entries []LogEntry) error {
	s.mu.Lock()
	s.entries = append(s.entries, entries...)
	s.mu.Unlock()
	return nil
}

func (s *InMemoryLogStore) Get(index int64) (*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 1 || int(index) > len(s.entries) {
		return nil, fmt.Errorf("index %d out of range", index)
	}
	e := s.entries[index-1]
	return &e, nil
}

func (s *InMemoryLogStore) GetRange(from, to int64) ([]LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if from < 1 {
		from = 1
	}
	if int(to) > len(s.entries) {
		to = int64(len(s.entries))
	}
	return append([]LogEntry{}, s.entries[from-1:to]...), nil
}

func (s *InMemoryLogStore) GetLastIndex() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.entries))
}

func (s *InMemoryLogStore) GetLastTerm() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.entries) == 0 {
		return 0
	}
	return s.entries[len(s.entries)-1].Term
}

func (s *InMemoryLogStore) DeleteRange(from int64) error {
	s.mu.Lock()
	if int(from) <= len(s.entries) {
		s.entries = s.entries[:from-1]
	}
	s.mu.Unlock()
	return nil
}

func (s *InMemoryLogStore) Compact(index int64) error {
	s.mu.Lock()
	if int(index) <= len(s.entries) {
		s.entries = s.entries[index:]
	}
	s.mu.Unlock()
	return nil
}

// --- InMemorySnapshotStore ---

type InMemorySnapshotStore struct {
	mu        sync.RWMutex
	snapshots []Snapshot
}

func (s *InMemorySnapshotStore) Save(snap Snapshot) error {
	s.mu.Lock()
	s.snapshots = append(s.snapshots, snap)
	s.mu.Unlock()
	return nil
}

func (s *InMemorySnapshotStore) Load() (*Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.snapshots) == 0 {
		return nil, nil
	}
	snap := s.snapshots[len(s.snapshots)-1]
	return &snap, nil
}

func (s *InMemorySnapshotStore) List() ([]Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Snapshot{}, s.snapshots...), nil
}

func (s *InMemorySnapshotStore) Delete(lastIncludedIndex int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var remaining []Snapshot
	for _, snap := range s.snapshots {
		if snap.LastIncludedIndex != lastIncludedIndex {
			remaining = append(remaining, snap)
		}
	}
	s.snapshots = remaining
	return nil
}

// --- MembershipManager ---

type MembershipManager struct {
	mu      sync.RWMutex
	members map[string]Member
}

func NewMembershipManager() *MembershipManager {
	return &MembershipManager{members: make(map[string]Member)}
}

func (m *MembershipManager) GetMembers() []Member {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Member, 0, len(m.members))
	for _, mem := range m.members {
		result = append(result, mem)
	}
	return result
}

func (m *MembershipManager) AddMember(member Member) error {
	m.mu.Lock()
	m.members[member.ID] = member
	m.mu.Unlock()
	return nil
}

func (m *MembershipManager) RemoveMember(nodeID string) error {
	m.mu.Lock()
	delete(m.members, nodeID)
	m.mu.Unlock()
	return nil
}

func (m *MembershipManager) GetMember(nodeID string) (*Member, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mem, ok := m.members[nodeID]
	if !ok {
		return nil, fmt.Errorf("member %q not found", nodeID)
	}
	return &mem, nil
}

func (m *MembershipManager) GetQuorum() int {
	m.mu.RLock()
	voters := 0
	for _, mem := range m.members {
		if mem.Role == MemberRoleVoter && mem.Status == MemberStatusAlive {
			voters++
		}
	}
	m.mu.RUnlock()
	return voters/2 + 1
}

// --- RaftNode (simplified, single-node or in-memory cluster) ---

type RaftNode struct {
	mu          sync.RWMutex
	config      ConsensusConfig
	state       NodeState
	currentTerm int64
	votedFor    string
	leader      string
	log         *InMemoryLogStore
	snapshots   *InMemorySnapshotStore
	membership  *MembershipManager
	commitIndex int64
	lastApplied int64
	stateMachine StateMachine
	peers       map[string]*RaftNode // simplified: direct refs instead of RPC
	cancel      context.CancelFunc
	electionTimer *time.Timer
}

type StateMachine interface {
	Apply(entry LogEntry) error
	GetState() interface{}
	CreateSnapshot() ([]byte, error)
	RestoreSnapshot(snapshot []byte) error
}

// KVStateMachine: простой key-value store как state machine
type KVStateMachine struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewKVStateMachine() *KVStateMachine { return &KVStateMachine{data: make(map[string]string)} }

func (kv *KVStateMachine) Apply(entry LogEntry) error {
	if cmd, ok := entry.Command.(map[string]string); ok {
		kv.mu.Lock()
		for k, v := range cmd {
			kv.data[k] = v
		}
		kv.mu.Unlock()
	}
	return nil
}

func (kv *KVStateMachine) GetState() interface{} {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	cp := make(map[string]string, len(kv.data))
	for k, v := range kv.data {
		cp[k] = v
	}
	return cp
}

func (kv *KVStateMachine) CreateSnapshot() ([]byte, error) {
	state := kv.GetState().(map[string]string)
	var buf []byte
	for k, v := range state {
		buf = append(buf, []byte(k+"="+v+"\n")...)
	}
	return buf, nil
}

func (kv *KVStateMachine) RestoreSnapshot(data []byte) error {
	kv.mu.Lock()
	kv.data = make(map[string]string)
	kv.mu.Unlock()
	return nil
}

func NewRaftNode(config ConsensusConfig, sm StateMachine) *RaftNode {
	n := &RaftNode{
		config:       config,
		state:        StateFollower,
		log:          &InMemoryLogStore{},
		snapshots:    &InMemorySnapshotStore{},
		membership:   NewMembershipManager(),
		stateMachine: sm,
		peers:        make(map[string]*RaftNode),
	}
	n.membership.AddMember(Member{
		ID: config.NodeID, Status: MemberStatusAlive, Role: MemberRoleVoter, JoinedAt: time.Now(),
	})
	return n
}

func (n *RaftNode) randomElectionTimeout() time.Duration {
	min := n.config.ElectionTimeoutMin
	max := n.config.ElectionTimeoutMax
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

func (n *RaftNode) Start(ctx context.Context) error {
	ctx, n.cancel = context.WithCancel(ctx)
	go n.runElectionTimer(ctx)
	return nil
}

func (n *RaftNode) Stop(_ context.Context) error {
	if n.cancel != nil {
		n.cancel()
	}
	return nil
}

func (n *RaftNode) runElectionTimer(ctx context.Context) {
	timeout := n.randomElectionTimeout()
	n.electionTimer = time.NewTimer(timeout)
	defer n.electionTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-n.electionTimer.C:
			n.mu.RLock()
			state := n.state
			n.mu.RUnlock()
			if state != StateLeader {
				n.startElection(ctx)
			}
			n.electionTimer.Reset(n.randomElectionTimeout())
		}
	}
}

func (n *RaftNode) startElection(ctx context.Context) {
	n.mu.Lock()
	n.currentTerm++
	n.state = StateCandidate
	n.votedFor = n.config.NodeID
	term := n.currentTerm
	lastIdx := n.log.GetLastIndex()
	lastTerm := n.log.GetLastTerm()
	n.mu.Unlock()

	votes := 1 // vote for self
	quorum := n.membership.GetQuorum()
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, peer := range n.peers {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := peer.RequestVote(ctx, VoteRequest{
				Term:         term,
				CandidateID:  n.config.NodeID,
				LastLogIndex: lastIdx,
				LastLogTerm:  lastTerm,
			})
			if err != nil || !resp.VoteGranted {
				return
			}
			mu.Lock()
			votes++
			mu.Unlock()
		}()
	}
	wg.Wait()

	n.mu.Lock()
	if votes >= quorum && n.currentTerm == term {
		n.state = StateLeader
		n.leader = n.config.NodeID
		n.mu.Unlock()
		go n.sendHeartbeats(ctx)
		return
	}
	n.state = StateFollower
	n.mu.Unlock()
}

func (n *RaftNode) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(n.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.mu.RLock()
			if n.state != StateLeader {
				n.mu.RUnlock()
				return
			}
			term := n.currentTerm
			prevIdx := n.log.GetLastIndex()
			prevTerm := n.log.GetLastTerm()
			commitIdx := n.commitIndex
			n.mu.RUnlock()
			for _, peer := range n.peers {
				go peer.AppendEntries(ctx, AppendEntriesRequest{
					Term: term, LeaderID: n.config.NodeID,
					PrevLogIndex: prevIdx, PrevLogTerm: prevTerm,
					LeaderCommit: commitIdx,
				})
			}
		}
	}
}

func (n *RaftNode) RequestVote(_ context.Context, req VoteRequest) (*VoteResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if req.Term < n.currentTerm {
		return &VoteResponse{Term: n.currentTerm, VoteGranted: false}, nil
	}
	if req.Term > n.currentTerm {
		n.currentTerm = req.Term
		n.state = StateFollower
		n.votedFor = ""
	}
	lastTerm := n.log.GetLastTerm()
	lastIdx := n.log.GetLastIndex()
	logOK := req.LastLogTerm > lastTerm || (req.LastLogTerm == lastTerm && req.LastLogIndex >= lastIdx)
	if (n.votedFor == "" || n.votedFor == req.CandidateID) && logOK {
		n.votedFor = req.CandidateID
		n.resetElectionTimer()
		return &VoteResponse{Term: n.currentTerm, VoteGranted: true}, nil
	}
	return &VoteResponse{Term: n.currentTerm, VoteGranted: false}, nil
}

func (n *RaftNode) AppendEntries(_ context.Context, req AppendEntriesRequest) (*AppendEntriesResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if req.Term < n.currentTerm {
		return &AppendEntriesResponse{Term: n.currentTerm, Success: false}, nil
	}
	n.currentTerm = req.Term
	n.state = StateFollower
	n.leader = req.LeaderID
	n.resetElectionTimer()

	if len(req.Entries) > 0 {
		n.log.Append(req.Entries)
	}
	if req.LeaderCommit > n.commitIndex {
		n.commitIndex = min64(req.LeaderCommit, n.log.GetLastIndex())
		// Apply committed entries
		for n.lastApplied < n.commitIndex {
			n.lastApplied++
			if e, err := n.log.Get(n.lastApplied); err == nil {
				n.stateMachine.Apply(*e)
			}
		}
	}
	return &AppendEntriesResponse{Term: n.currentTerm, Success: true}, nil
}

func (n *RaftNode) InstallSnapshot(_ context.Context, req SnapshotRequest) (*SnapshotResponse, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if req.Term < n.currentTerm {
		return &SnapshotResponse{Term: n.currentTerm}, nil
	}
	snap := Snapshot{
		LastIncludedIndex: req.LastIncludedIndex,
		LastIncludedTerm:  req.LastIncludedTerm,
		Data:              req.Data,
		Timestamp:         time.Now(),
	}
	n.snapshots.Save(snap)
	n.stateMachine.RestoreSnapshot(req.Data)
	return &SnapshotResponse{Term: n.currentTerm}, nil
}

func (n *RaftNode) Propose(ctx context.Context, command interface{}) error {
	n.mu.Lock()
	if n.state != StateLeader {
		n.mu.Unlock()
		return fmt.Errorf("not leader; current leader: %s", n.leader)
	}
	term := n.currentTerm
	n.mu.Unlock()

	entry := LogEntry{
		Index:     n.log.GetLastIndex() + 1,
		Term:      term,
		Command:   command,
		Timestamp: time.Now(),
	}
	n.log.Append([]LogEntry{entry})

	// Replicate to peers and wait for quorum
	quorum := n.membership.GetQuorum()
	replicated := 1
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, peer := range n.peers {
		peer := peer
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := peer.AppendEntries(ctx, AppendEntriesRequest{
				Term: term, LeaderID: n.config.NodeID,
				PrevLogIndex: entry.Index - 1,
				PrevLogTerm:  term,
				Entries:      []LogEntry{entry},
				LeaderCommit: n.commitIndex,
			})
			if err == nil && resp.Success {
				mu.Lock()
				replicated++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if replicated >= quorum {
		n.mu.Lock()
		n.commitIndex = entry.Index
		n.lastApplied = entry.Index
		n.mu.Unlock()
		n.stateMachine.Apply(entry)
	}
	return nil
}

func (n *RaftNode) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state == StateLeader
}

func (n *RaftNode) GetState() NodeInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return NodeInfo{
		ID: n.config.NodeID, State: n.state,
		CurrentTerm: n.currentTerm, VotedFor: n.votedFor,
		Leader: n.leader, LastLogIndex: n.log.GetLastIndex(),
		CommitIndex: n.commitIndex, LastApplied: n.lastApplied,
	}
}

func (n *RaftNode) AddNode(_ context.Context, nodeID, address string) error {
	return n.membership.AddMember(Member{
		ID: nodeID, Address: address, Status: MemberStatusAlive,
		Role: MemberRoleVoter, JoinedAt: time.Now(),
	})
}

func (n *RaftNode) RemoveNode(_ context.Context, nodeID string) error {
	return n.membership.RemoveMember(nodeID)
}

func (n *RaftNode) AddPeer(peer *RaftNode) {
	n.mu.Lock()
	n.peers[peer.config.NodeID] = peer
	n.mu.Unlock()
}

func (n *RaftNode) resetElectionTimer() {
	if n.electionTimer != nil {
		n.electionTimer.Reset(n.randomElectionTimeout())
	}
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

type NodeInfo struct {
	ID           string
	State        NodeState
	CurrentTerm  int64
	VotedFor     string
	Leader       string
	LastLogIndex int64
	LastLogTerm  int64
	CommitIndex  int64
	LastApplied  int64
}

func main() {
	cfg := func(id string) ConsensusConfig {
		return ConsensusConfig{
			NodeID:              id,
			ElectionTimeoutMin:  150 * time.Millisecond,
			ElectionTimeoutMax:  300 * time.Millisecond,
			HeartbeatInterval:   50 * time.Millisecond,
			SnapshotInterval:    100,
			MaxEntriesPerAppend: 10,
		}
	}

	n1 := NewRaftNode(cfg("node-1"), NewKVStateMachine())
	n2 := NewRaftNode(cfg("node-2"), NewKVStateMachine())
	n3 := NewRaftNode(cfg("node-3"), NewKVStateMachine())

	// Wire up peers
	for _, n := range []*RaftNode{n1, n2, n3} {
		for _, peer := range []*RaftNode{n1, n2, n3} {
			if peer.config.NodeID != n.config.NodeID {
				n.AddPeer(peer)
				n.AddNode(context.Background(), peer.config.NodeID, "")
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n1.Start(ctx)
	n2.Start(ctx)
	n3.Start(ctx)

	// Wait for leader election
	time.Sleep(500 * time.Millisecond)

	// Find leader
	var leader *RaftNode
	for _, n := range []*RaftNode{n1, n2, n3} {
		if n.IsLeader() {
			leader = n
			break
		}
	}

	if leader != nil {
		fmt.Printf("Leader: %s, term: %d\n", leader.config.NodeID, leader.GetState().CurrentTerm)
		err := leader.Propose(ctx, map[string]string{"key": "hello", "value": "raft"})
		fmt.Println("Propose err:", err)
		state := leader.stateMachine.GetState()
		fmt.Println("State machine:", state)
	} else {
		fmt.Println("No leader elected yet (increase timeout and try again)")
	}

	fmt.Println("node-1 state:", n1.GetState().State)
	fmt.Println("node-2 state:", n2.GetState().State)
	fmt.Println("node-3 state:", n3.GetState().State)

	// Log store demo
	ls := &InMemoryLogStore{}
	ls.Append([]LogEntry{{Index: 1, Term: 1, Command: "set x=1"}})
	ls.Append([]LogEntry{{Index: 2, Term: 1, Command: "set x=2"}})
	fmt.Println("last log index:", ls.GetLastIndex())
	entry, _ := ls.Get(1)
	fmt.Println("log entry 1:", entry.Command)
}
