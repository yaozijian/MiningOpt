package distribution

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/cihub/seelog"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/clientselector"
	"github.com/yaozijian/MiningOpt/optimization"
	"github.com/yaozijian/rpcxutils"
)

type (
	Worker struct {
		client *rpcx.Client
		cfg    RpcxServerConfig
	}

	TaskStatus struct {
		rpcxutils.TaskStatus
		TaskId string
		Desc   string
		Worker string
	}
)

func NewWorker(cfg RpcxServerConfig) interface{} {

	s := clientselector.NewEtcdV3ClientSelector(
		cfg.EtcdServers,
		fmt.Sprintf("%v/%v", rpcx_base_path, rpcx_manager_name),
		time.Minute,
		rpcx.RoundRobin,
		time.Second*3,
	)

	w := &Worker{
		client: rpcx.NewClient(s),
		cfg:    cfg,
	}

	return w
}

func (w *Worker) MiningOpt(ctx context.Context, args optimization.MiningOptParams, returl *string) error {

	var e error

	if e = w.downloadFile(args.TaskId, args.InputFile, task_data_file); e != nil {
		log.Error(e)
		return e
	} else if e = w.downloadFile(args.TaskId, args.ParamFile, task_param_file); e != nil {
		log.Error(e)
		return e
	} else {

		dir := fmt.Sprintf("%v/%v", task_data_dir, args.TaskId)

		args.OutputFile = fmt.Sprintf("%v/%v", dir, task_result_file)
		args.InputFile = fmt.Sprintf("%v/%v", dir, task_data_file)
		args.ParamFile = fmt.Sprintf("%v/%v", dir, task_param_file)

		if e = w.doOptimization(&args); e == nil {
			*returl = fmt.Sprintf("%v/%v", w.cfg.URLPrefix, args.OutputFile)
		}

		return e
	}
}

func (w *Worker) doOptimization(args *optimization.MiningOptParams) (err error) {

	args.Notify = make(chan string, 10)

	// Notify manager of task's status
	go func() {

		var hasdata bool

		status := TaskStatus{
			TaskId: args.TaskId,
			Worker: w.cfg.ServiceAddr,
		}
		status.TaskStatus.Status = rpcxutils.Task_Running

		for {
			if status.Desc, hasdata = <-args.Notify; hasdata {
				w.client.Call(context.Background(), rpcx_manager_notify, status, nil)
			} else {
				break
			}
		}
	}()

	err = optimization.DoMiningOptimization(*args)

	close(args.Notify)

	return
}

func (w *Worker) downloadFile(taskid, url, name string) error {

	body, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("Failed to download file %v from %v for task %v: %v", name, url, taskid, err)
	} else if body == nil || body.Body == nil {
		return fmt.Errorf("Failed to download file %v from %v for task %v: empty reply body", name, url, taskid)
	} else {
		defer body.Body.Close()

		dir := fmt.Sprintf("%v/%v", task_data_dir, taskid)

		os.MkdirAll(dir, 0777)

		if file, err := os.Create(dir + "/" + name); err != nil {
			os.Remove(dir)
			return fmt.Errorf("Error to create file %v for task %v: %v", name, taskid, err)
		} else if _, err = io.Copy(file, body.Body); err != nil {
			file.Close()
			os.Remove(dir)
			return fmt.Errorf("Error to write file %v for task %v: %v", name, taskid, err)
		} else {
			file.Close()
			log.Infof("Download file %v from %v for task %v: OK", name, url, taskid)
			return nil
		}
	}
}
