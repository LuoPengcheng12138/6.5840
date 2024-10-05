package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"6.5840/labgob"
	"6.5840/labrpc"
)


// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	//Persistent
	currentTerm int
	votedFor int
	log []LogEntry

	//volatile state on all servers
	commitIndex int
	lastApplied int

	//volatile state on leaders
	nextIndex []int
	matchIndex []int

	//other argument
	state int //Leader:0 Candidate:1 Follower:2
	applyCh chan ApplyMsg
	electionTime *time.Timer
	heartbeatTime *time.Timer


}

const Leader int =0
const Candidate int =1
const Follower int =2

func(rf *Raft) ChangeState(state int){
	if rf.state == state {
		return
	}
	Debug(dLeader,"{Node %v} changes state from %v to %v", rf.me, rf.state, state)
	rf.state = state
	switch state {
	case Follower:
		rf.electionTime.Reset(RandomElectionTimeout())
		rf.heartbeatTime.Stop() // stop heartbeat
	case Candidate:
	case Leader:
		rf.electionTime.Stop() // stop election
		rf.heartbeatTime.Reset(StableHeartbeatTimeout())
	}

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int
	var isleader bool
	// Your code here (3A).
	term=rf.currentTerm
	isleader= (rf.state==0)
	return term, isleader
}

func (rf *Raft) encodeState() []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	return w.Bytes()
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}


// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}


// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

}



// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	Debug(dInfo,"{Node %v} receives RequestVoteArgs %v and raft :currentTerm: %v votedfor %v", rf.me, args,rf.currentTerm,rf.votedFor)
	// Your code here (3A, 3B).
	if args.Term<rf.currentTerm {
		reply.Term=rf.currentTerm
		reply.VoteGranted=false
		return
	}
	if  rf.votedFor!=-1&&rf.votedFor!=args.CalledcandidateId {
		reply.Term=rf.currentTerm
		reply.VoteGranted=false
		return
	}
	if !rf.IsLogUpdate(args.LastLogTerm,args.LastLogIndex){
		reply.Term=rf.currentTerm
		reply.VoteGranted=false
		return
	}
	if args.Term>rf.currentTerm {
		reply.Term=args.Term
		reply.VoteGranted=true
		rf.currentTerm=args.Term
		rf.votedFor=args.CalledcandidateId
		rf.ChangeState(Follower)
		rf.electionTime.Reset(RandomElectionTimeout())
		return
	}else{
		reply.Term=args.Term
		reply.VoteGranted=true
		rf.votedFor=args.CalledcandidateId
		rf.electionTime.Reset(RandomElectionTimeout())
	}

}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply){ 
	if args.Entries==nil { //心跳
		Debug(dInfo,"{Node %v} state:%v receives Heartbeat from {Node %v} ", rf.me, rf.state,args.LeaderId)
		if args.Term > rf.currentTerm {
			rf.currentTerm, rf.votedFor = args.Term, -1
			rf.persist()
		}
		rf.votedFor=-1
		rf.ChangeState(Follower) 
		rf.currentTerm=args.Term
		rf.heartbeatTime.Reset(StableHeartbeatTimeout())
		rf.electionTime.Reset(RandomElectionTimeout())
		reply.Success=false
	}
} 

func (rf *Raft) IsLogUpdate(logterm,logindex int) bool{
	lastlog:=rf.log[len(rf.log)-1]
	if logterm>lastlog.Term ||(logterm==lastlog.Term&&logindex>=lastlog.Index){
		return true
	}
	return false
}



// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := false

	// Your code here (3B).


	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here (3A)
		// Check if a leader election should be started.
		select{
			case <-rf.heartbeatTime.C: //接收到心跳信号
				rf.mu.Lock()
				switch rf.state {
					case Leader: //leader 发送heartbeat & 重设定时器
						rf.heartbeatTime.Reset(StableHeartbeatTimeout())
						rf.BoardCastHeartbeat()
					case Follower: //重设定时器 在rf.AppendEntries中完成	
					case Candidate:	
				}
				rf.mu.Unlock()
			case <-rf.electionTime.C:
				rf.mu.Lock()
				switch rf.state {
					case Leader: //leader 不需要 electionTime
					case Follower: 
						rf.ChangeState(Candidate)
						rf.currentTerm+=1
						rf.electionTime.Reset(RandomElectionTimeout())
						rf.StartElection()

					case Candidate:	
						rf.electionTime.Reset(RandomElectionTimeout())
						rf.StartElection()
				}
				rf.mu.Unlock()

		}
		
		// // pause for a random amount of time between 50 and 350
		// // milliseconds.
		// ms := 50 + (rand.Int63() % 300)
		// time.Sleep(time.Duration(ms) * time.Millisecond)
		
	}
}

func (rf *Raft) BoardCastHeartbeat(){  //leader can do
	// Debug(dInfo,"{Node %v} starts BoardCastHeartbeat", rf.me)
	// for peer:=range rf.peers {
	// 	if peer==rf.me {
	// 		continue
	// 	}
	// 	firstLogIndex := rf.getFirstLog().Index //第一个log的index
	// 	prevLogIndex:=0
	// 	if rf.nextIndex[peer] - 1>=0 {
	// 		prevLogIndex =rf.nextIndex[peer] - 1 //最后一个log的index
	// 	}else {
	// 		prevLogIndex=0
	// 	}
	// 	args:=&AppendEntriesArgs{
	// 		Term :rf.currentTerm,
	// 		LeaderId :rf.me,
	// 		PrevLogIndex : prevLogIndex,
	// 		PrevLogTerm :rf.log[prevLogIndex-firstLogIndex].Term,
	// 		Entries :nil,
	// 		LeaderCommit :rf.commitIndex,
	// 	}
	// 	reply:=&AppendEntriesReply{}
	// 	Debug(dInfo,"{Node %v} starts BoardCastHeartbeat to {Node %v}", rf.me,peer)
	// 	rf.sendAppendEntries(peer,args,reply)

	// }
	args := rf.genAppendEntriesArgs(prevLogIndex)
	rf.mu.RUnlock()
	reply := new(AppendEntriesReply)
	if rf.sendAppendEntries(peer, args, reply) {
		rf.mu.Lock()
		if args.Term == rf.currentTerm && rf.state == Leader {
			if !reply.Success {
				if reply.Term > rf.currentTerm {
					// indicate current server is not the leader
					rf.ChangeState(Follower)
					rf.currentTerm, rf.votedFor = reply.Term, -1
					rf.persist()
				} else if reply.Term == rf.currentTerm {
					// decrease nextIndex and retry
					rf.nextIndex[peer] = reply.ConflictIndex
					// TODO: optimize the nextIndex finding, maybe use binary search
					if reply.ConflictTerm != -1 {
						firstLogIndex := rf.getFirstLog().Index
						for index := args.PrevLogIndex - 1; index >= firstLogIndex; index-- {
							if rf.logs[index-firstLogIndex].Term == reply.ConflictTerm {
								rf.nextIndex[peer] = index
								break
							}
						}
					}
				}
			} else {
				rf.matchIndex[peer] = args.PrevLogIndex + len(args.Entries)
				rf.nextIndex[peer] = rf.matchIndex[peer] + 1
				// advance commitIndex if possible
				rf.advanceCommitIndexForLeader()
			}
		}
		rf.mu.Unlock()
		DPrintf("{Node %v} sends AppendEntriesArgs %v to {Node %v} and receives AppendEntriesReply %v", rf.me, args, peer, reply)
	}

	
}



func (rf *Raft) StartElection() {
	rf.votedFor = rf.me //给自己投票
	//rf.persist()
	args := rf.genRequestVoteArgs()
	grantedVotes := 1 //统计票数
	Debug(dVote,"{Node %v} starts election with RequestVoteArgs %v", rf.me, args)
	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}
		go func(peer int) {
			reply := new(RequestVoteReply)
			if rf.sendRequestVote(peer, args, reply) {
				rf.mu.Lock()
				defer rf.mu.Unlock()
				Debug(dInfo,"{Node %v} receives RequestVoteReply %v from {Node %v} after sending RequestVoteArgs %v", rf.me, reply, peer, args)
				if args.Term == rf.currentTerm && rf.state == Candidate {
					if reply.VoteGranted {
						grantedVotes += 1
						// check over half of the votes
						if grantedVotes > len(rf.peers)/2 {
							Debug(dInfo,"{Node %v} receives over half of the votes", rf.me)
							rf.ChangeState(Leader)
							rf.BoardCastHeartbeat()
						}
					} else if reply.Term > rf.currentTerm {
						rf.ChangeState(Follower)
						rf.currentTerm, rf.votedFor = reply.Term, -1
						//rf.persist()
					}
				}
			}
		}(peer)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.mu=sync.Mutex{}
	rf.dead=0
	rf.currentTerm=0
	rf.votedFor=-1
	rf.log=make([]LogEntry, 1)
	rf.commitIndex=0
	rf.lastApplied=0
	rf.nextIndex=make([]int, len(peers))
	rf.nextIndex=make([]int, len(peers))
	rf.matchIndex=make([]int, len(peers))
	for i := range rf.matchIndex {
		rf.matchIndex[i] = -1
	}
	rf.state=Follower
	rf.electionTime=time.NewTimer(RandomElectionTimeout())
	rf.heartbeatTime=time.NewTimer(StableHeartbeatTimeout())
	rf.applyCh= applyCh



	// Your initialization code here (3A, 3B, 3C).

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()


	return rf
}
