package raft

import ("log"
		"time"
		"math/rand"	
		"fmt")


// Debugging
// const Debug = false

// func DPrintf(format string, a ...interface{}) {
// 	if Debug {
// 		log.Printf(format, a...)
// 	}
// }
const debug = 1
func Debug(topic logTopic, format string, a ...interface{}) {
	if debug >= 1 {
		time := time.Since(debugStart).Microseconds()
		time /= 100
		prefix := fmt.Sprintf("%06d %v ", time, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}

type LogEntry struct{
	Term int
	Index int
	Command interface{}
}

func RandomElectionTimeout() time.Duration {
	return time.Duration(1000 + (rand.Int63() % 2000)) * time.Millisecond //1000-2000ms
}

func StableHeartbeatTimeout() time.Duration {
	return time.Duration(125) * time.Millisecond //125ms
}

func (rf *Raft) getLastLog() LogEntry {
	return rf.log[len(rf.log)-1]
}

func (rf *Raft) getFirstLog() LogEntry {
	return rf.log[0]
}

func (rf *Raft) genRequestVoteArgs() *RequestVoteArgs{
	args:=&RequestVoteArgs{
		Term: rf.currentTerm,
		CalledcandidateId: rf.me,
		LastLogIndex: rf.getLastLog().Index,
		LastLogTerm: rf.getLastLog().Term,
	}
	return args
}

func (rf *Raft) genAppendEntriesArgs() *AppendEntriesArgs{
	firstLogIndex := rf.getFirstLog().Index //第一个log的index
	prevLogIndex :=rf.nextIndex[rf.me] - 1 //最后一个log的index
	args:=&AppendEntriesArgs{
		Term :rf.currentTerm,
		LeaderId :rf.me,
		PrevLogIndex : prevLogIndex,
		PrevLogTerm :rf.log[prevLogIndex-firstLogIndex].Term,
		Entries :nil,  //todooo
		LeaderCommit :rf.commitIndex,
	}
	return args
}

