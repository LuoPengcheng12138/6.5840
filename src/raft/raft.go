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
	//"math/rand"
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
	nextIndex []int // index of the next log entry to send to that server
	matchIndex []int // index of highest log entry known to be replicated on server

	//other argument
	state int //Leader:0 Candidate:1 Follower:2
	applyCh chan ApplyMsg
	electionTime *time.Timer
	heartbeatTime *time.Timer

	applyCond      chan int    // condition variable for apply goroutine
	replicatorCond []chan int  // condition variable for replicator goroutine
	// applyCond      *sync.Cond    // condition variable for apply goroutine
	// replicatorCond []*sync.Cond  // condition variable for replicator goroutine

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
		rf.electionTime.Reset(RandomElectionTimeout())
		rf.heartbeatTime.Stop() // stop heartbeat
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
	rf.mu.Lock()
	defer rf.mu.Unlock()
	Debug(dInfo,"{Node %v} receives RequestVoteArgs %v and raft :currentTerm: %v votedfor %v", rf.me, args,rf.currentTerm,rf.votedFor)
	// Your code here (3A, 3B).
	if args.Term < rf.currentTerm || (args.Term == rf.currentTerm && rf.votedFor != -1 && rf.votedFor != args.CalledcandidateId) {
		reply.Term, reply.VoteGranted = rf.currentTerm, false
		return
	}

	if args.Term > rf.currentTerm {
		rf.ChangeState(Follower)
		rf.currentTerm, rf.votedFor = args.Term, -1
		rf.persist()
	}

	// if candidate's log is not up-to-date, reject the vote(§5.4)
	if !rf.IsLogUpdate(args.LastLogIndex, args.LastLogTerm) {
		reply.Term, reply.VoteGranted = rf.currentTerm, false
		return
	}

	rf.votedFor = args.CalledcandidateId
	rf.persist()
	rf.electionTime.Reset(RandomElectionTimeout())
	reply.Term, reply.VoteGranted = rf.currentTerm, true
	
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply){ 
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer Debug(dInfo,"{Node %v}'s state is {state %v, term %v}} after processing AppendEntries,  AppendEntriesArgs %v and AppendEntriesReply %v ", rf.me, rf.state, rf.currentTerm, args, reply)
	
	// Reply false if term < currentTerm(§5.1)
	if args.Term < rf.currentTerm {
		reply.Term, reply.Success = rf.currentTerm, false
		return
	}
	// this is heartbeat 
	if args.Term > rf.currentTerm {
		rf.currentTerm, rf.votedFor = args.Term, -1
		rf.persist()
	}
	// AppendEntries 都是心跳 都要changeState
	rf.ChangeState(Follower) 
	rf.electionTime.Reset(RandomElectionTimeout())

	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm(§5.3)
	// Todo

	// if an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it(§5.3)
	if !rf.isLogMatched(args.PrevLogIndex,args.PrevLogTerm){ //return false ：index 不匹配 ｜｜term不匹配
		reply.Term,reply.Success=rf.currentTerm,false // **
		if args.PrevLogIndex>rf.getLastLog().Index{ //index 不匹配 
			reply.ConflictIndex=rf.getLastLog().Index+1
			reply.ConflictTerm=-1
		}else{		//term不匹配
			index:=args.PrevLogIndex
			fristlogindex:=rf.getFirstLog().Index
			for index>=fristlogindex && rf.log[index-fristlogindex].Term==args.PrevLogTerm{
				index--
			}
			reply.ConflictIndex=index+1
			reply.ConflictTerm=args.PrevLogTerm
		}
		return // **
	}
	//
	firstLogIndex:=rf.getFirstLog().Index
	rf.log=append(rf.log[:args.PrevLogIndex+1-firstLogIndex],args.Entries...)
	Debug(dWarn,"################# folllower log: %v",rf.log)

	// If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry) (paper)
	newCommitIndex := Min(args.LeaderCommit, rf.getLastLog().Index)
	if newCommitIndex > rf.commitIndex {
		Debug(dInfo,"{Node %v} advances commitIndex from %v to %v with leaderCommit %v in term %v", rf.me, rf.commitIndex, newCommitIndex, args.LeaderCommit, rf.currentTerm)
		rf.commitIndex = newCommitIndex
		//rf.applyCond.Signal()
	}
	reply.Term, reply.Success = rf.currentTerm, true
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
//开始对一个新的命令进行一致性协议（即将该命令添加到 Raft 的日志中）
//新日志索引：如果命令被提交，命令将在日志中的位置
//当前任期：Raft 实例的当前任期
//是否是领导者：当前服务器是否认为自己是领导者
// submit a command 
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := false

	//Debug(dLeader,"{Node %v} Start function ",rf.me)

	// Your code here (3B).
	if rf.state!=Leader {
		return -1,-1,false
	}
	//将command添加到日志
	newLogIndex := rf.getLastLog().Index + 1
	rf.log=append(rf.log, LogEntry{
		Term :rf.currentTerm,
		Index :rf.getLastLog().Index+1,
		Command :command,
	})
	
	Debug(dWarn,"{Node %v}将{command %v}添加到日志 index:%v",rf.me,command,rf.getLastLog().Index)
	Debug(dWarn,"################# leader log: %v",rf.log)
	//** 更新匹配和下一索引 **
	rf.nextIndex[rf.me]=newLogIndex+1
	rf.matchIndex[rf.me]=newLogIndex

	Debug(dWarn,"{Node %v} starts agreement on a new log entry with command %v in term %v", rf.me, command, rf.currentTerm)
	
	//广播日志 TODO 自己不做 通过chan 通知 replicator 的 goruntion 去做，会循环查询 replicatorCond 状态
	for peer:=range rf.peers {
		if peer==rf.me {
			continue
		}
		Debug(dWarn,"{Node %v} replicatorCond[%v]<-1", rf.me, peer)
		rf.replicatorCond[peer]<-1 //signal

	}
	Debug(dWarn,"{Node %v} 广播日志 index:%v",rf.me,rf.getLastLog().Index)

	index=newLogIndex
	term=rf.currentTerm
	isLeader=true
	Debug(dWarn,"index %v,term %v,isLeader %v",index,term,isLeader)
	return index,term,isLeader
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
						rf.currentTerm+=1
						rf.electionTime.Reset(RandomElectionTimeout())
						rf.StartElection()
				}
				rf.mu.Unlock()

		}

	}
}

func(rf *Raft) replicator(peer int){
	// rf.mu.Lock() 	//不能加锁 有很多peer 会发生死锁
	// defer rf.mu.Unlock()  
	Debug(dInfo,"{Node %v} starts replicator  to {Node %v}", rf.me,peer)
	//Debug(dLeader,"{Node %v} received replicatorCond[%v] value:%v", rf.me, peer,<-rf.replicatorCond[peer])
	for !rf.killed(){
		select {
			case <-rf.replicatorCond[peer]:
				Debug(dLeader,"{Node %v} received replicatorCond[%v]<-1", rf.me, peer)
				if rf.state == Leader && rf.matchIndex[peer] < rf.getLastLog().Index {
					rf.BoardCastLogs(peer)
				}
			default:
				time.Sleep(50 * time.Millisecond) // 可调节的休眠时间
			}
	}
}

func (rf *Raft) BoardCastLogs(peer int) { 
	rf.mu.Lock()
	defer rf.mu.Unlock() 
	Debug(dLeader,"{Node %v} starts BoardCast-Logs", rf.me)

	prevLogIndex := rf.nextIndex[peer] - 1
	args:=rf.genAppendEntriesArgs(prevLogIndex)
	reply:=new(AppendEntriesReply)
	if rf.sendAppendEntries(peer,args,reply){
		if !reply.Success{ //false 说明log不匹配 或者 follower term > leader
			if reply.Term>rf.currentTerm{
				rf.ChangeState(Follower)
				rf.currentTerm, rf.votedFor = reply.Term, -1
			}else { //log not match
				rf.nextIndex[peer]=reply.ConflictIndex
				if reply.ConflictIndex!=-1{
					index:=args.PrevLogIndex
					firstLogIndex:=rf.getFirstLog().Index
					for index>=firstLogIndex{
						if rf.log[index-firstLogIndex].Term==reply.ConflictTerm{
							rf.nextIndex[peer]=index
							break
						}
						index--
					}
				}
			}
		}else{ //没有冲突 日志添加成功
			rf.matchIndex[peer]=args.PrevLogIndex+len(args.Entries)
			rf.nextIndex[peer]=rf.matchIndex[peer]+1
			// **
		}
		Debug(dLog,"BoardCast-Logs :{Node %v} sends AppendEntriesArgs %v to {Node %v} and receives AppendEntriesReply %v", rf.me, args, peer, reply)
	}
}

func (rf *Raft) BoardCastHeartbeat() {  //leader can do
	//Debug(dLeader,"{Node %v} starts BoardCast-Heartbeat", rf.me)
	// rf.mu.Lock()
	// defer rf.mu.Unlock() //会发生死锁 BoardCastHeartbeat->StartElection->ticker(有锁)
	for peer:=range rf.peers {
		
		if peer==rf.me {
			continue
		}
		go func(peer int) {
			prevLogIndex := rf.nextIndex[peer] - 1
			args := rf.genAppendEntriesArgs(prevLogIndex)
			reply := new(AppendEntriesReply)
			if rf.sendAppendEntries(peer, args, reply) {
				if args.Term == rf.currentTerm && rf.state == Leader {			
					if reply.Term > rf.currentTerm {
						// indicate current server is not the leader
						rf.ChangeState(Follower)
						rf.currentTerm, rf.votedFor = reply.Term, -1
						rf.persist()
					} else if reply.Term == rf.currentTerm {
						// decrease nextIndex and retry
						// TODO: optimize the nextIndex finding, maybe use binary search		
					}				
				}
				Debug(dLeader,"BoardCast-Heartbeat :{Node %v} sends AppendEntriesArgs %v to {Node %v} and receives AppendEntriesReply %v", rf.me, args, peer, reply)
			}
		}(peer)
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
			Debug(dVote,"{Node %v} starts election to {Node %v}", rf.me, peer)
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
	// Your initialization code here (3A, 3B, 3C).
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
	rf.state=Follower
	rf.electionTime=time.NewTimer(RandomElectionTimeout())
	rf.heartbeatTime=time.NewTimer(StableHeartbeatTimeout())
	rf.applyCh= applyCh
	rf.applyCond=make(chan int)
	rf.replicatorCond=make([]chan int, len(peers))
	for i := range rf.peers {
		rf.replicatorCond[i] = make(chan int) //无缓冲通道
	}

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	for peer :=range peers {
		rf.matchIndex[peer]=0
		rf.nextIndex[peer]=rf.getLastLog().Index+1
		if peer!=rf.me {
			//Debug(dInfo,"{Node %v} starts replicator go runtion %v", rf.me,peer)
			go rf.replicator(peer)
		}

	}


	// start ticker goroutine to start elections
	go rf.ticker()


	return rf
}
