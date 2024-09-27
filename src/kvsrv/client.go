package kvsrv

import "6.5840/labrpc"
import "crypto/rand"
import "math/big"
import "time"

//注意：	kv.mu.Lock()
//		   defer kv.mu.Unlock()只是对一个clerk中的读写进行锁定
//但在并发时，是多个client进行读写，此时仍然会出现race问题，需要对map进行并发管理

//出现memory占用过高问题 可以使client收到reply后，再次和server发送args表示已收到，此时server可以删除那个requestId

type Clerk struct {
	server *labrpc.ClientEnd
	// You will have to modify this struct.
	clerkId int64
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(server *labrpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.server = server
	// You'll have to add code here.
	ck.clerkId = nrand()
	return ck
}

// fetch the current value for a key.
// returns "" if the key does not exist.
// keeps trying forever in the face of all other errors.
//
// you can send an RPC with code like this:
// ok := ck.server.Call("KVServer.Get", &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) Get(key string) string {

	// You will have to modify this function.
	args:=GetArgs{}
	reply:=GetReply{}
	args.Key=key
	args.RequestId=nrand()
	args.ClerkId=ck.clerkId

    for {  //没有收到就重试
        ok := ck.server.Call("KVServer.Get", &args, &reply)
        if ok {
            return reply.Value
        }
        // 可以添加一个睡眠以避免紧密循环
        time.Sleep(100 * time.Millisecond) // 休眠 100 毫秒
    }

}

// shared by Put and Append.
//
// you can send an RPC with code like this:
// ok := ck.server.Call("KVServer."+op, &args, &reply)
//
// the types of args and reply (including whether they are pointers)
// must match the declared types of the RPC handler function's
// arguments. and reply must be passed as a pointer.
func (ck *Clerk) PutAppend(key string, value string, op string) string {
	// You will have to modify this function.
	args:=PutAppendArgs{}
	reply:=PutAppendReply{}
	args.Key=key
	args.Value=value
	args.RequestId=nrand()
	args.ClerkId=ck.clerkId

	switch op {
    case "Put":
		for{
			ok := ck.server.Call("KVServer.Put", &args, &reply)
			if ok{
				replynone:=PutAppendReply{}
				ck.server.Call("KVServer.UpdatePutAppend", &args,&replynone)
				return reply.Value
			}
			time.Sleep(100* time.Millisecond)
		}
    case "Append":
		for{
			ok := ck.server.Call("KVServer.Append", &args, &reply)
			if ok{
				replynone:=PutAppendReply{}
				ck.server.Call("KVServer.UpdatePutAppend", &args,&replynone)
				return reply.Value
			}
			time.Sleep(100 * time.Millisecond)
		}
    default:
        DPrintf("worng")
	}
	
	return ""
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "Put")
}

// Append value to key's value and return that value
func (ck *Clerk) Append(key string, value string) string {
	return ck.PutAppend(key, value, "Append")
}


