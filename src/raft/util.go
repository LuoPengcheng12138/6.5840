package raft

import ("log"
		"time"
		"math/rand"	
		"fmt"
		"sort")


// Debugging
// const Debug = false

// func DPrintf(format string, a ...interface{}) {
// 	if Debug {
// 		log.Printf(format, a...)
// 	}
// }
const debug = 0
func Debug(topic logTopic, format string, a ...interface{}) {
	if debug >= 1{
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
func (rf *Raft) isLogMatched(index, term int) bool {
	return index <= rf.getLastLog().Index && term == rf.log[index-rf.getFirstLog().Index].Term
}

func (rf *Raft) IsLogUpdate(logterm,logindex int) bool{
	lastlog:=rf.log[len(rf.log)-1]
	if logterm>lastlog.Term ||(logterm==lastlog.Term&&logindex>=lastlog.Index){
		return true
	}
	return false
}
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// shrinkEntriesArray discards the underlying array used by the entries slice
// if most of it isn't being used. This avoids holding references to a bunch of
// potentially large entries that aren't needed anymore. Simply clearing the
// entries wouldn't be safe because clients might still be using them.
func shrinkEntries(entries []LogEntry) []LogEntry {
	const lenMultiple = 2
	if cap(entries) > len(entries)*lenMultiple {
		newEntries := make([]LogEntry, len(entries))
		copy(newEntries, entries)
		return newEntries
	}
	return entries
}
func (rf *Raft) advanceCommitIndexForLeader() { //根据已复制到大多数服务器的日志条目来更新领导者的提交索引
	n := len(rf.matchIndex)
	sortMatchIndex := make([]int, n)
	copy(sortMatchIndex, rf.matchIndex)
	sort.Ints(sortMatchIndex)
	// get the index of the log entry with the highest index that is known to be replicated on a majority of servers
	newCommitIndex := sortMatchIndex[n-(n/2+1)]
	if newCommitIndex > rf.commitIndex {
		if rf.isLogMatched(newCommitIndex, rf.currentTerm) {
			Debug(dCommit,"{Node %v} advances commitIndex from %v to %v in term %v", rf.me, rf.commitIndex, newCommitIndex, rf.currentTerm)
			rf.commitIndex = newCommitIndex
			rf.applyCond<-1
		}
	}
}

//TIMR,VOTE,LEAD,TERM,LOG1,LOG2,CMIT,PERS,SNAP,DROP,CLNT,TEST,INFO,WARN,ERRO,TRCE