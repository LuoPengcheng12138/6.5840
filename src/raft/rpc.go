package raft

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term int
	CalledcandidateId int
	LastLogIndex int
	LastLogTerm int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term int
	VoteGranted bool
}

type AppendEntriesArgs struct{
	Term int
	LeaderId int
	PrevLogIndex int
	PrevLogTerm int
	Entries []LogEntry
	LeaderCommit int
	
}

type AppendEntriesReply struct{
	Term int
	Success bool
	ConflictIndex int
	ConflictTerm int
}

func (rf *Raft) genRequestVoteArgs() *RequestVoteArgs {
	args:=&RequestVoteArgs{
		Term :rf.currentTerm,
		CalledcandidateId :rf.me,
		LastLogIndex :rf.getLastLog().Index,
		LastLogTerm :rf.getLastLog().Term,
	}
	return args

}

//prevLogIndex：跟随者应该在其日志中查找的最后一条已知条目的索引
func (rf *Raft) genAppendEntriesArgs(prevLogIndex int) *AppendEntriesArgs{
	firstLogIndex := rf.getFirstLog().Index //本地第一个log的index 
	//eg:prevLogIndex=7 (peer) 那么拷贝rf中 log[prevLogIndex] 后面的entry就行 但是要减去 firstLogIndex=4 (me)基数才是logs中的位置
	entries:=make([]LogEntry,len(rf.log[prevLogIndex-firstLogIndex+1:])) //+1表明拷贝log[prevLogIndex]后面的entry
	copy(entries,rf.log[prevLogIndex-firstLogIndex+1:])
	args:=&AppendEntriesArgs{
		Term :rf.currentTerm,
		LeaderId :rf.me,
		PrevLogIndex : prevLogIndex,
		PrevLogTerm :rf.log[prevLogIndex-firstLogIndex].Term,
		Entries :entries,  //todooo
		LeaderCommit :rf.commitIndex,
	}
	return args
}
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs,reply *AppendEntriesReply) bool{
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}