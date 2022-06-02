package concurrent

import "sync"

type Runnable func(arg interface{})

type Task struct {
	Func Runnable
	Arg  interface{}
}

type DEQueue interface {
	PushBottom(task Runnable, taskArg interface{})
	IsEmpty() bool
	PopTop() *Task
	PopBottom() *Task
}

type Node struct {
	task *Task
	prev *Node
	next *Node
}

type Queue struct {
	top  *Node
	bot  *Node
	size *int
	lock *sync.Mutex
}
