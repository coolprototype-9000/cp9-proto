package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

type TaskType int

const (
	Map    TaskType = iota
	Reduce TaskType = iota
	Done   TaskType = iota
)

type AskArgs struct{}
type TellReply struct{}

type AskReply struct {
	Task        TaskType
	TaskId      int
	ReduceTasks int // undefined if Task == Reduce
	Filename    string
}

type TellArgs struct {
	Task     TaskType // probably not needed but whatever
	TaskId   int
	Filename string
}
