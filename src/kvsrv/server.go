package kvsrv

import (
	"log"
	"sync"
	"strconv"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}


type KVServer struct {
	mu sync.Mutex
	kvmap map[string]string
	seenId sync.Map //key:requestId+clerkId	value:old record
	//seenClerk map[int64]bool
	// Your definitions here.
}


func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	reply.Value=kv.kvmap[args.Key]

}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	v, ok := kv.seenId.Load(strconv.FormatInt(args.ClerkId,10)+strconv.FormatInt(args.RequestId,10))
	if ok{
		str, _:= v.(string)
		reply.Value=str
		return
	}
	kv.kvmap[args.Key]=args.Value
	kv.seenId.Store(strconv.FormatInt(args.ClerkId,10)+strconv.FormatInt(args.RequestId,10),args.Value)

}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	// Your code here.
	kv.mu.Lock()
	defer kv.mu.Unlock()
	v, ok := kv.seenId.Load(strconv.FormatInt(args.ClerkId,10)+strconv.FormatInt(args.RequestId,10))
	if ok{
		str, _:= v.(string)
		reply.Value=str
		return
	}

	old:=kv.kvmap[args.Key]
	reply.Value=old
	kv.kvmap[args.Key]=kv.kvmap[args.Key]+args.Value
	kv.seenId.Store(strconv.FormatInt(args.ClerkId,10)+strconv.FormatInt(args.RequestId,10),old)

}


func (kv *KVServer) UpdatePutAppend(args *PutAppendArgs, reply *PutAppendReply) {

	kv.seenId.Delete(strconv.FormatInt(args.ClerkId,10)+strconv.FormatInt(args.RequestId,10))

}



func StartKVServer() *KVServer {
	kv := new(KVServer)

	// You may need initialization code here.
	kv.kvmap = make(map[string]string)

	return kv
}
