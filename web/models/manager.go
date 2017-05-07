package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	log "github.com/cihub/seelog"
)

type (
	Task interface {
		Id() string
		Create() time.Time
		Server() string
		Status() string
		Progress() float64
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
		id       string
		create   time.Time
		status   string
		progress float64
		server   Server
		sync.Mutex
	}

	task_status struct {
		Create time.Time
		Status string
	}

	manager struct {
		task_list   []*task_item
		server_list []*server_item
		sync.Mutex
	}
)

const (
	Task_data_dir    = "tasks"
	Task_id_len      = 8
	Task_data_file   = "data.gz"
	Task_param_file  = "param.json"
	Task_result_file = "result.gz"
)

const (
	Task_New         = "Submiited"
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

	file := fmt.Sprintf("%v/%v/status.json", Task_data_dir, id)

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

		item := &task_item{
			id:     id,
			create: status.Create,
			status: status.Status,
		}

		m.Lock()
		m.task_list = append(m.task_list, item)
		m.Unlock()

		return nil
	}
}
