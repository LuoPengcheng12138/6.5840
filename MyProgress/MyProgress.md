遇到的问题：

KV-Server：
1 内存测试不通过，需要重发PutAppendArgs，让服务器删除map元素

Lab3A：
1 rf.votedfor在接收到心跳时没有初始化，导致重新选举时rf.votedFor!=args.CalledcandidateId ，reply false
    1.5 补充 看一下 *RequestVote* 函数 是怎么改votedfor的
2 func(rf *Raft) ChangeState(state int)没有重置定时器
3 当一个leader的心跳rpc调用全都返回false，认为leader掉线，changeState(Follower)
    3.5 补充：3可以不写这个逻辑，因为节点重新连接后会收到Term比它大的心跳，会自动转为Follower
4 BoardCastHeartbeat 需要 go func(peer int) {}(peer) 来并行发心跳
5 func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply)只对心跳进行了处理,后续添加
6 genAppendEntriesArgs中需要传入 prevLogIndex：跟随者应该在其日志中查找的最后一条已知条目的索引 
  是prevLogIndex := rf.nextIndex[peer] - 1 而不是 是prevLogIndex := rf.nextIndex[rf.me] - 1

Lab3B:
1 Start提交命令时 更新自己的 rf.nextIndex[rf.me]=newLogIndex+1 && rf.matchIndex[rf.me]=newLogIndex
2 func(rf *Raft) replicator(peer int) 不能加锁 有很多peer 会发生死锁
3 for i := range rf.peers {
		rf.replicatorCond[i] = make(chan int) //无缓冲通道
	}  // 使用chan作为信号量 要设置无缓冲通道
4 AppendEntries 添加逻辑	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm(§5.3) 
5 需要 applier 函数 固定到状态机
6 advanceCommitIndexForLeader 函数根据已复制到大多数服务器的日志条目来更新领导者的提交索引，在BoardCastLogs 日志添加成功时调用
