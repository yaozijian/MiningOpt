package distribution

import (
	"context"
	"fmt"
	"time"

	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/yaozijian/MiningOpt/optimization"
	"github.com/yaozijian/rpcxutils"
)

type (
	Manager interface {
		MiningOpt(ctx context.Context, args optimization.MiningOptParams, returl *string)
		NotifyChnl() <-chan TaskStatus
	}

	manager struct {
		cfg    RpcxServerConfig
		taskq  *rpcxutils.TaskQueue
		notify chan TaskStatus
		rpcxutils.ServiceWrapper
	}
)

func NewManager(cfg RpcxServerConfig) interface{} {

	s := clientselector.NewEtcdV3ClientSelector(
		cfg.EtcdServers,
		fmt.Sprintf("%v/%v", rpcx_base_path, rpcx_worker_name),
		time.Minute,
		rpcx.RoundRobin,
		15*time.Second,
	)

	w := &manager{
		cfg:    cfg,
		taskq:  rpcxutils.NewTaskQueue(s),
		notify: make(chan TaskStatus, 1000),
	}

	w.ServiceWrapper.SetTaskQueue(w.taskq)

	return w
}

func (m *manager) MiningOpt(ctx context.Context, args optimization.MiningOptParams, returl *string) {

	task := m.ServiceWrapper.Go(ctx, rpcx_worker_method, args, returl)

	go func() {

		status := &TaskStatus{
			TaskId: args.TaskId,
		}

		for notify := range task.Notify() {

			status.TaskStatus = notify

			if notify.Error != nil {
				status.Desc = notify.Error.Error()
			} else {
				status.Desc = ""
			}

			m.notify <- *status
		}
	}()
}

func (m *manager) NotifyChnl() <-chan TaskStatus {
	return m.notify
}

func (m *manager) Notify(status TaskStatus, retval *int) error {
	select {
	case m.notify <- status:
	default:
	}
	return nil
}
