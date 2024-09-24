package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//


import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//
type WorkerArgs struct{
	MapTaskNumber int 	//finished number
	ReduceTaskNumber int
}

type WokerReply struct{
	Tasktype int 	//0:map,1:reduce,2:waiting,3:finished
	NMap int	//total map number
	NReduce int

	MapTaskNumber int 	// to be finish number
	ReduceTaskNumber int	
	Filename string 	//for map task
}

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
