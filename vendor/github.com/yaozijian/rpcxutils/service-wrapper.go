package rpcxutils

import (
	"context"
	"time"
)

type (
	ServiceWrapper struct {
		queue *TaskQueue
	}
)

const (
	nested_call_flag = "Nested_Call_Flag"
)

func (w *ServiceWrapper) NeedNestedCall(ctx context.Context) bool {
	return (w.queue != nil) && (&MapContextHelper{ctx}).Get(nested_call_flag) != nested_call_flag
}

func (w *ServiceWrapper) SetTaskQueue(queue *TaskQueue) {
	w.queue = queue
}

func (w *ServiceWrapper) NestedCall(method string, ctx context.Context, args, reply interface{}) (bool, error) {
	if w.NeedNestedCall(ctx) {
		h := &MapContextHelper{ctx}
		h.Header2().Set(nested_call_flag, nested_call_flag)
		return true, w.Call(h.NewContext(), method, args, reply)
	} else {
		return false, nil
	}
}

func (w *ServiceWrapper) Call(ctx context.Context, method string, args, reply interface{}) error {

	task := w.call(ctx, method, args, reply)

	for state := range task.Notify() {
		switch state.Status {
		case Task_Done_OK, Task_Done_Failed, Task_Timeout_Waiting, Task_Timeout_Running, Task_Aborted:
			return state.Error
		}
	}

	return nil
}

func (w *ServiceWrapper) Go(ctx context.Context, method string, args, reply interface{}) *taskContext {
	return w.call(ctx, method, args, reply)
}

func (w *ServiceWrapper) call(ctx context.Context, method string, args, reply interface{}) *taskContext {

	param := &TaskParam{
		Ctx:    ctx,
		Method: method,
		Args:   args,
		Reply:  reply,
	}

	h := &ContextHelper{ctx}

	if timeout := h.Get("Timeout"); len(timeout) > 0 {
		if dura, e := time.ParseDuration(timeout); e == nil {
			param.Timeout = dura
		}
	}

	task := w.queue.AddTask(param)

	return task
}
