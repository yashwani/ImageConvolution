package concurrent

import "sync"

//newQueue returns a pointer to a double ended queue
func NewQueue() DEQueue {
	top := Node{task: &Task{Func: nil, Arg: 0}}
	bot := Node{task: &Task{Func: nil, Arg: 1}}
	top.next = &bot
	bot.prev = &top
	size := 0
	q := Queue{top: &top, bot: &bot, size: &size, lock: &sync.Mutex{}}
	return &q
}

//PushBottom is called by the main goroutine to add tasks to a worker queue
func (q *Queue) PushBottom(task Runnable, taskArg interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	newTask := Task{Func: task, Arg: taskArg}
	newNode := &Node{task: &newTask}
	beforeBot := q.bot.prev
	beforeBot.next = newNode
	newNode.prev = beforeBot
	newNode.next = q.bot
	q.bot.prev = newNode
	*q.size += 1

}

//IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return *q.size == 0

}

//PopTop pops a task from the top of a queue. Used by a worker when stealing from another worker.
func (q *Queue) PopTop() *Task {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.IsEmpty() {
		return nil
	}
	returnNode := q.top.next
	afterTop := returnNode.next
	q.top.next = afterTop
	afterTop.prev = q.top

	*q.size -= 1

	return returnNode.task
}

//PopBottom pops a task from the bottom of a queue. Used by workers to get tasks from their local queues
func (q *Queue) PopBottom() *Task {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.IsEmpty() {
		return nil
	}

	returnNode := q.bot.prev
	beforeBot := returnNode.prev
	q.bot.prev = beforeBot
	beforeBot.next = q.bot

	*q.size -= 1

	return returnNode.task
}
