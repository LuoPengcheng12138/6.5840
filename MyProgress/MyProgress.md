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
 func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply)只对心跳进行了处理,后续添加
4 锁写的很凌乱