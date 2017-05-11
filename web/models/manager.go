package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/yaozijian/MiningOpt/distribution"
	"github.com/yaozijian/MiningOpt/optimization"
	"github.com/yaozijian/rpcxutils"

	log "github.com/cihub/seelog"
)

type (
	Task interface {
		Id() string
		Create() time.Time
		Server() string
		Status() string
		DataURL() string
		ParamURL() string
		ResultURL() string
	}

	TaskManager interface {
		TaskList() []Task
		ServerList() []Server
		NewTask(id string) error
	}

	task_item struct {
		id     string
		create time.Time
		status string
		desc   string
		outurl string
		server Server
		sync.Mutex
	}

	task_status struct {
		Create time.Time
		Status string
	}

	manager struct {
		task_list   []*task_item
		server_list []*server_item
		task_center distribution.Manager
		urlprefix   string
		sync.Mutex
	}
)

const (
	Task_data_dir    = "tasks"
	Task_id_len      = 8
	Task_data_file   = "data.gz"
	Task_param_file  = "param.json"
	Task_result_file = "result.gz"
	Task_status_file = "status.json"
)

const (
	Task_New         = "Submited"
	Task_Waiting     = "Wating"
	Task_Running     = "Running"
	Task_Done_OK     = "Done OK"
	Task_Done_Failed = "Done Failed"
	Task_Timeout     = "Timeout"
	Task_Aborted     = "Canceled"
)

var (
	task_manager *manager
)

func (m *manager) TaskList() []Task {

	m.Lock()
	defer m.Unlock()

	if len(m.task_list) == 0 {
		return nil
	}

	retlist := make([]Task, len(m.task_list))

	for idx, item := range m.task_list {
		retlist[idx] = item
	}

	return retlist
}

func (m *manager) ServerList() []Server {

	m.Lock()
	defer m.Unlock()

	if len(m.server_list) == 0 {
		return nil
	}

	retlist := make([]Server, len(m.server_list))

	for idx, item := range m.server_list {
		retlist[idx] = item
	}

	return retlist
}

func (m *manager) NewTask(id string) error {

	file := fmt.Sprintf("%v/%v/%v", Task_data_dir, id, Task_status_file)

	status := &task_status{
		Create: time.Now(),
		Status: Task_New,
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	if e := enc.Encode(status); e != nil {
		log.Errorf("Failed to encode task status: %v", e)
		return e
	} else if e := ioutil.WriteFile(file, buf.Bytes(), 0666); e != nil {
		log.Errorf("Failed to write task status file %v: %v", file, e)
		return e
	} else {

		task := &task_item{
			id:     id,
			create: status.Create,
			status: status.Status,
		}

		m.Lock()
		m.task_list = append(m.task_list, task)
		m.Unlock()

		m.run_task(task)

		return nil
	}
}

func (m *manager) run_task(task *task_item) {

	args := optimization.MiningOptParams{
		TaskId:    task.id,
		InputFile: m.urlprefix + task.DataURL(),
		ParamFile: m.urlprefix + task.ParamURL(),
	}

	m.task_center.MiningOpt(context.Background(), args, &task.outurl)
}

func (m *manager) waitNotify() {
	for {
		status := <-m.task_center.NotifyChnl()

		log.Infof(
			"Notify: id=%v status=%v desc=%v",
			status.TaskId,
			rpcxutils.TaskState2Desc(status.TaskStatus.Status),
			status.Desc,
		)

		m.Lock()
		for _, task := range m.task_list {
			if status.TaskId == task.id {
				task.status = rpcxutils.TaskState2Desc(status.TaskStatus.Status)
				task.desc = status.Desc
				break
			}
		}
		m.Unlock()
	}
}
