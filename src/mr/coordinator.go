package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)


type Coordinator struct {
	// Your definitions here.

	nMap int	//total map number
	nReduce int

	filenames []string

	mapfinished int 
	reducefinished int
	mapTaskLog []int 	// 0:not allocated 1:waiting 2:finished
	reduceTaskLog []int	

	mutex sync.Mutex

}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) ReceivedFinishedMap(args *WorkerArgs,reply *WokerReply) error{
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.mapfinished++
	c.mapTaskLog[args.MapTaskNumber]=2
	return nil
}

func (c *Coordinator) ReceivedFinishedReduce(args *WorkerArgs,reply *WokerReply) error{
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.reducefinished++
	c.reduceTaskLog[args.ReduceTaskNumber]=2
	return nil
}

func (c *Coordinator) ReplyAllocateTask (args *WorkerArgs,reply *WokerReply) error{
	c.mutex.Lock()
	defer c.mutex.Unlock()
	//allocate map
	if(c.mapfinished<c.nMap){
		allocate:=-1
		for i, value := range c.mapTaskLog {
			if(value==0){
				allocate=i
				reply.NMap=c.nMap
				reply.NReduce=c.nReduce
				reply.MapTaskNumber=allocate
				reply.Filename=c.filenames[i]
				reply.Tasktype=0//map task
				c.mapTaskLog[allocate]=1			
				go func(){
					time.Sleep(time.Second*10)
					if c.mapTaskLog[allocate]==1{
						c.mapTaskLog[allocate]=0
					}
				}()
				break
			}
			
		}
		if(allocate==-1){ //all allcated but waiting
			reply.Tasktype=2 //waiting
			reply.NMap=c.nMap
			reply.NReduce=c.nReduce			
			// go func(){
			// 	time.Sleep(time.Second*10)
			// 	for i, value := range c.mapTaskLog{
			// 		if(value==1){
			// 			c.mapTaskLog[i]=0
			// 			break
			// 		}
			// 	}
			// }()
		}
	}else if c.mapfinished==c.nMap && c.reducefinished<c.nReduce{
		allocate:=-1
		for i, value := range c.reduceTaskLog {
			if(value==0){
				allocate=i
				reply.NMap=c.nMap
				reply.NReduce=c.nReduce
				reply.ReduceTaskNumber=allocate
				reply.Tasktype=1//reduce task
				c.reduceTaskLog[allocate]=1
				go func(){
					time.Sleep(time.Second*10)
					if c.reduceTaskLog[allocate]==1{
						c.reduceTaskLog[allocate]=0
					}
				}()
				break
			}
		}
		if(allocate==-1){ //all allcated but waiting
			reply.Tasktype=2 //waiting
			reply.NMap=c.nMap
			reply.NReduce=c.nReduce
		}	
	}else{
		reply.Tasktype=3
	}

	return nil

}
//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.
	if c.mapfinished==c.nMap && c.reducefinished==c.nReduce{
		ret=true
	}

	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.filenames=files
	c.nMap=len(files)
	c.nReduce=nReduce
	c.mapTaskLog=make([]int, c.nMap)
	c.reduceTaskLog=make([]int, c.nReduce)
	c.mapfinished=0
	c.reducefinished=0

	c.server()
	return &c
}
