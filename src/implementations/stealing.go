package implementations

import (
	"math/rand"
	"src/concurrent"
	"sync"
)

// StealingScheduler Stealing Scheduler is the context for the Work Stealing Scheduler
type StealingScheduler struct {
	localQueues  []concurrent.DEQueue
	localWorkers []*stealingWorker
	capacity     int
	lastPushed   int
	cond         sync.Cond
	lock         *sync.Mutex
	wg           *sync.WaitGroup
	moreTasks    bool
}

// NewStealingScheduler Creates a new Work Stealing context
func NewStealingScheduler(capacity int) *StealingScheduler {
	//create local queues
	localQueues := make([]concurrent.DEQueue, capacity)
	for i := 0; i < len(localQueues); i++ {
		localQueues[i] = concurrent.NewQueue()
	}

	//stealingScheduler context to give to all the workers
	localWorkers := make([]*stealingWorker, capacity)
	stealingScheduler := StealingScheduler{localQueues: localQueues, localWorkers: localWorkers, capacity: capacity, lastPushed: -1}
	stealingScheduler.lock = &sync.Mutex{}
	stealingScheduler.cond = sync.Cond{L: stealingScheduler.lock}
	stealingScheduler.wg = &sync.WaitGroup{}
	stealingScheduler.moreTasks = true

	//create localWorkers and given them stealingScheduler context
	for i := 0; i < len(localWorkers); i++ {
		stealingScheduler.localWorkers[i] = NewStealingWorker(i, &stealingScheduler)
	}

	return &stealingScheduler
}

//PushTask inserts a new task into a queue of one of the threads in the pool
//You need to insure each thread is assigned an equal amount of work
func (scheduler *StealingScheduler) PushTask(task concurrent.Runnable, taskArg interface{}) {
	scheduler.lastPushed += 1
	idx := scheduler.lastPushed % scheduler.capacity
	scheduler.localQueues[idx].PushBottom(task, taskArg)
	scheduler.cond.Broadcast()
}

// Run starts the run() method for each thread in the pool
func (scheduler *StealingScheduler) Run() {
	scheduler.wg.Add(scheduler.capacity)
	for _, stealingWorker := range scheduler.localWorkers {
		go stealingWorker.run()
	}
	scheduler.cond.Broadcast()
}

// Done is a way for the application to tell the scheduler no more tasks will
// need to be handled by the scheduler. This method should notify the workers in some way
func (scheduler *StealingScheduler) Done() {
	scheduler.moreTasks = false
	scheduler.cond.Broadcast()
}

// Wait() is called when a goroutine wants to wait until all of the stealing threads in the pool have executed. You may use a sync.Waitgroup in this implementation but only one.
func (scheduler *StealingScheduler) Wait() {
	scheduler.wg.Wait()

}

type stealingWorker struct {
	workerId  int
	scheduler *StealingScheduler
}

// Creates a new stealer worker (i.e., a thread in the pool of workers)
func NewStealingWorker(workerId int, scheduler *StealingScheduler) *stealingWorker {
	return &stealingWorker{workerId: workerId, scheduler: scheduler}
}

// Run is the implementations function being executed by a stealing worker.
func (worker *stealingWorker) run() {
	for worker.scheduler.moreTasks {
		worker.takeFromSelf()
		worker.steal()
		worker.takeFromSelf()

		worker.scheduler.lock.Lock()
		for worker.queueIsEmpty() {
			if worker.scheduler.moreTasks {
				worker.scheduler.cond.Wait()
			} else {
				worker.scheduler.lock.Unlock()
				worker.takeFromSelf()
				worker.steal()
				worker.scheduler.wg.Done()
				return
			}
		}
		worker.scheduler.lock.Unlock()
	}
	worker.takeFromSelf()
	worker.steal()
	worker.scheduler.wg.Done()
	return
}

//takeFromSelf worker takes from own queue until nothing else in own queue
func (worker *stealingWorker) takeFromSelf() {
	myId := worker.workerId
	task := worker.scheduler.localQueues[myId].PopBottom()
	for task != nil {
		ExecuteRunnable(task.Func, task.Arg)
		task = worker.scheduler.localQueues[myId].PopBottom()
	}
}

//steal worker steals from queues until a steal comes up without a task
func (worker *stealingWorker) steal() {
	randID := worker.randomId()
	stolentask := worker.scheduler.localQueues[randID].PopTop()
	for stolentask != nil {
		ExecuteRunnable(stolentask.Func, stolentask.Arg)
		randID := worker.randomId()
		stolentask = worker.scheduler.localQueues[randID].PopTop()
	}
}

//returns true if the worker's queue is empty
func (worker *stealingWorker) queueIsEmpty() bool {
	return worker.scheduler.localQueues[worker.workerId].IsEmpty()
}

//returns a random worker ID that is not itself
func (worker *stealingWorker) randomId() int {
	randID := rand.Intn(worker.scheduler.capacity)
	for randID == worker.workerId {
		randID = rand.Intn(worker.scheduler.capacity)
	}
	return randID
}

func ExecuteRunnable(runnable concurrent.Runnable, arg interface{}) {
	runnable(arg)
}
