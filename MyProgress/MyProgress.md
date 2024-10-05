遇到的问题：

KV-Server：
1 内存测试不通过，需要重发PutAppendArgs，让服务器删除map元素

Lab3A：
1 rf.votedfor在接收到心跳时没有初始化，导致重新选举时rf.votedFor!=args.CalledcandidateId ，reply false
2 func(rf *Raft) ChangeState(state int)没有重置定时器
3 当一个leader的心跳rpc调用全都返回false，认为leader掉线，changeState(Follower)
3 func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply)只对心跳进行了处理，结构很糟糕，需要重构
4 锁写的很凌乱