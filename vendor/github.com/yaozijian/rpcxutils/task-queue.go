package rpcxutils

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/smallnest/rpcx/core"
)

type (
	TaskQueue struct {
		*clientselector.EtcdV3ClientSelector

		server_running map[*core.Client]*taskContext // Client -> Running Task

		task_total_cnt   int
		task_wait_cnt    int
		task_running_cnt int
		task_timeout_cnt int
		task_done_cnt    int

		task_running []*taskContext // tasks running by a server
		task_done    []*taskContext // tasks completed

		select_list      []reflect.SelectCase
		select_interrupt chan interface{} // interrupt select loop for some reason

		timeout timeout_manager

		exitreq chan struct{}
		sync.Once
	}

	TaskParam struct {
		Ctx     context.Context // call context
		Method  string          // call which method
		Args    interface{}     // params for the method
		Reply   interface{}     // used for receiving result
		Timeout time.Duration
	}

	TaskStatus struct {
		Status   TaskState // status of the task,is one of Task_xxx below
		Error    error     // error desc,nil for no error
		CallDone bool      // task done by server(success or fail)
		CtxDone  bool      // task done by client(such as timeout or canceled)
	}

	callContext struct {
		client         *core.Client // client used for this task
		call           *core.Call   // call context
		proxy          *rpcx.Client // proxy for this task
		select_idx     int
		heap_idx       int
		wait_call_done reflect.SelectCase // select case for task done by server
		wait_ctx_done  reflect.SelectCase // select case for task done by client
	}

	taskContext struct {
		notify  chan TaskStatus
		timeout time.Time
		param   TaskParam
		status  TaskStatus
		callctx callContext
	}

	TaskState int
)

const (
	Task_New TaskState = iota
	Task_Waiting
	Task_Running
	Task_Done_OK
	Task_Done_Failed
	Task_Timeout_Waiting
	Task_Timeout_Running
	Task_Aborted
)

var (
	Error_no_server_available  = fmt.Errorf("No server available")
	Error_task_waiting_timeout = fmt.Errorf("No server available: Task timeout")
	Error_task_running_timeout = fmt.Errorf("Server too slow: Task timeout")
	Error_task_abort           = fmt.Errorf("Task queue exit,task aborted")
)

var (
	state2desc = map[TaskState]string{
		Task_New:             "Submited",
		Task_Waiting:         "Wating",
		Task_Running:         "Running",
		Task_Done_OK:         "Success",
		Task_Done_Failed:     "Failed",
		Task_Timeout_Waiting: "Waiting Timeout",
		Task_Timeout_Running: "Running Timeout",
		Task_Aborted:         "Aborted",
	}
)

func TaskState2Desc(s TaskState) string {
	return state2desc[s]
}

func NewTaskQueue(s *clientselector.EtcdV3ClientSelector) *TaskQueue {

	ctrl := &TaskQueue{
		EtcdV3ClientSelector: s,
		server_running:       make(map[*core.Client]*taskContext),
		select_interrupt:     make(chan interface{}, 100),
	}

	item := reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctrl.select_interrupt),
	}

	ctrl.select_list = append(ctrl.select_list, item)

	item = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctrl.timeout.NotifyChnl()),
	}

	ctrl.select_list = append(ctrl.select_list, item)

	go ctrl.mainLoop()
	go ctrl.watch()

	return ctrl
}

func (s *TaskQueue) AddTask(ptr *TaskParam) *taskContext {

	obj := &taskContext{
		param:  *ptr,
		notify: make(chan TaskStatus, 10),
	}

	if ptr.Timeout > 0 {
		obj.timeout = time.Now().Add(ptr.Timeout)
	}

	obj.status.Status = Task_New

	s.select_interrupt <- obj

	return obj
}

func (s *TaskQueue) Stop() {
	s.Once.Do(func() {
		s.exitreq = make(chan struct{})
		close(s.select_interrupt)
		<-s.exitreq
		s.exitreq = nil
	})
}

func (s *taskContext) Notify() <-chan TaskStatus {
	return s.notify
}

func (s *taskContext) notifyStateChanged() {

	s.notify <- s.status

	switch s.status.Status {
	case Task_Done_OK, Task_Done_Failed, Task_Timeout_Waiting, Task_Timeout_Running, Task_Aborted:
		close(s.notify)
	default:
	}
}

//-----------------------------------------------------------------------

func (s *TaskQueue) mainLoop() {
	for {
		s.select_list[1].Chan = reflect.ValueOf(s.timeout.NotifyChnl())

		idx, recv, ok := reflect.Select(s.select_list)

		if idx == 0 {
			if ok {
				switch val := recv.Interface().(type) {
				case *taskContext:
					s.addNewTask(val)
				case *core.Client:
					delete(s.server_running, val)
				case struct{}: // watch 观察到了服务器事件,需要尝试下能否运行新的任务
					s.runNexTask()
				}
			} else {
				s.timeout.Clear()
				close(s.exitreq)
				return
			}
		} else if idx == 1 {
			s.processTimeout()
		} else if ok {
			s.onTaskCompleted(idx)
			s.runNexTask()
		}
	}
}

func (s *TaskQueue) addNewTask(obj *taskContext) {

	obj.status.Status = Task_Waiting

	obj.notifyStateChanged()

	s.timeout.AddTask(obj)

	s.task_total_cnt++
	s.task_wait_cnt++

	if s.task_wait_cnt == 1 {
		s.runNexTask()
	}
}

func (s *TaskQueue) runNexTask() error {
	if s.task_wait_cnt <= 0 {
		return nil
	} else if task := s.timeout.GetNextWatingTask(); task != nil {
		if e := s.runTask(task); e == nil {
			return nil
		} else {
			return e
		}
	} else {
		return nil
	}
}

func (s *TaskQueue) runTask(obj *taskContext) error {

	ptr := &obj.param

	obj.callctx.proxy = rpcx.NewClient(s)

	s.SetClient(obj.callctx.proxy)

	client, err := s.Select(obj.callctx.proxy.ClientCodecFunc)

	if err != nil {
		return err
	}

	s.server_running[client] = obj

	obj.callctx.client = client
	obj.callctx.call = client.Go(ptr.Ctx, ptr.Method, ptr.Args, ptr.Reply, nil)

	obj.callctx.wait_call_done = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(obj.callctx.call.Done),
	}

	obj.callctx.wait_ctx_done = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ptr.Ctx.Done()),
	}

	obj.callctx.select_idx = len(s.select_list)
	s.select_list = append(s.select_list, obj.callctx.wait_call_done)
	s.select_list = append(s.select_list, obj.callctx.wait_ctx_done)

	s.task_running = append(s.task_running, obj)

	obj.status.Status = Task_Running

	obj.notifyStateChanged()

	s.task_wait_cnt--
	s.task_running_cnt++

	return nil
}

func (s *TaskQueue) onTaskCompleted(selectidx int) {

	var taskidx int
	var calldone bool

	/*
		0  select interrupt signal
		1  timeout signal
		2,4,6,... task call done
		3,5,7,... task context done
	*/
	if calldone = (selectidx%2 == 0); calldone {
		taskidx = (selectidx - 2) / 2
	} else {
		taskidx = (selectidx - 3) / 2
	}

	ptask := s.task_running[taskidx]

	if calldone {
		ptask.status.CallDone = true
		ptask.status.Error = ptask.callctx.call.Error
	} else {
		ptask.status.CtxDone = true
		ptask.status.Error = ptask.param.Ctx.Err()
	}

	// Reset timeout
	s.timeout.RemoveTask(ptask)

	s.removeRunningTask(ptask, selectidx, taskidx)

	if calldone && ptask.status.Error == nil {
		ptask.status.Status = Task_Done_OK
	} else {
		ptask.status.Status = Task_Done_Failed
	}

	ptask.notifyStateChanged()

	s.task_running_cnt--
	s.task_done_cnt++
}

func (s *TaskQueue) removeRunningTask(task *taskContext, selectidx, taskidx int) {

	// remove from running list
	s.task_running = append(s.task_running[:taskidx], s.task_running[taskidx+1:]...)

	for _, item := range s.task_running[taskidx:] {
		item.callctx.select_idx -= 2
	}

	// remove from select_list
	s.select_list = append(s.select_list[:selectidx], s.select_list[selectidx+2:]...)

	// server used by this task is idle now
	delete(s.server_running, task.callctx.client)

	// add to done list
	s.task_done = append(s.task_done, task)
}

//-----------------------------

func (s *TaskQueue) processTimeout() {

	task, old := s.timeout.ProcessTimeout()

	if task == nil {
		return
	}

	switch old {

	case Task_Waiting:

		s.task_wait_cnt--
		s.task_timeout_cnt++
		s.task_done = append(s.task_done, task)

	case Task_Running:

		selectidx := task.callctx.select_idx
		taskidx := (task.callctx.select_idx - 2) / 2

		s.removeRunningTask(task, selectidx, taskidx)

		s.task_running_cnt--
		s.task_timeout_cnt++

		s.runNexTask()
	}
}

//------------------------------------------------------------------------

func (s *TaskQueue) HandleFailedClient(client *core.Client) {
	s.EtcdV3ClientSelector.HandleFailedClient(client)
	s.select_interrupt <- client
}

func (s *TaskQueue) watch() {

	sel := s.EtcdV3ClientSelector
	watcher := sel.KeysAPI.Watch(context.Background(), sel.BasePath, clientv3.WithPrefix())

	for range watcher {
		// 延迟一下,等EtcdV3ClientSelector先处理(可能是增加可用的服务器)
		time.Sleep(time.Second)
		// 服务器列表发生变化,通知可尝试运行下一个等待中的任务
		select {
		case s.select_interrupt <- struct{}{}:
		default:
		}
	}
}

func (s *TaskQueue) Select(clientCodecFunc rpcx.ClientCodecFunc, options ...interface{}) (client *core.Client, err error) {

	server_cnt := len(s.EtcdV3ClientSelector.Servers)

	// Try for every server at most.If one server is not running a task,then select it.
	for i := 0; i < server_cnt; i++ {
		if client, err = s.EtcdV3ClientSelector.Select(clientCodecFunc, options...); client != nil {
			if _, ok := s.server_running[client]; !ok {
				return
			}
		} else {
			return
		}
	}

	err = Error_no_server_available

	return
}
