package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/astaxie/beego"
	"github.com/yaozijian/MiningOpt/distribution"
	"github.com/yaozijian/relayr"

	log "github.com/cihub/seelog"
)

func Init(webcfg *WebConfig) error {
	switch webcfg.StartType {
	case WebType_Manager:
		return runTaskManager(webcfg)
	case WebType_Worker:
		return runTaskWorker(webcfg)
	default:
		return fmt.Errorf("Wrong StartType")
	}
}

func runTaskManager(webcfg *WebConfig) error {

	task_manager = &manager{}

	task_manager.readin_task_list()

	sort.SliceStable(task_manager.task_list, func(x, y int) bool {
		a := task_manager.task_list[x]
		b := task_manager.task_list[y]
		return a.create.Unix() < b.create.Unix()
	})

	//-----

	rpcxcfg := distribution.RpcxServerConfig{
		ServiceAddr: fmt.Sprintf("%v:%v", webcfg.MyIpAddr, webcfg.RpcxPort),
		EtcdServers: webcfg.EtcdServers,
		URLPrefix:   fmt.Sprintf("http://%v:%v", webcfg.MyIpAddr, webcfg.HttpPort),
		Async:       true,
	}

	task_manager.urlprefix = rpcxcfg.URLPrefix

	task_manager.task_center = distribution.StartManager(rpcxcfg)

	if task_manager.task_center == nil {
		return fmt.Errorf("Error to start manager")
	}

	//-----

	go task_manager.waitNotify()

	task_manager.client_notify = relayr.NewExchange()
	task_manager.client_notify.RegisterRelay(TaskStatusNotify{})
	beego.Handler("/relayr/", task_manager.client_notify, true)

	return nil
}

func GetTaskManager() TaskManager {
	return task_manager
}

func (m *manager) readin_task_list() {

	filepath.Walk(Task_data_dir, func(p string, info os.FileInfo, e error) error {

		if e != nil {
			return e
		}

		if info.IsDir() && p != Task_data_dir {
			m.readin_task(p)
		}

		return nil
	})
}

func (m *manager) readin_task(dir string) error {

	idx := strings.LastIndex(dir, string(os.PathSeparator))
	id := dir[idx+1:]

	f := dir + "/" + Task_status_file

	cont, e := ioutil.ReadFile(f)

	if e != nil {
		log.Errorf("failed to read task status file %v: %v", f, e)
		return e
	}

	var status task_status

	if e = json.Unmarshal(cont, &status); e != nil {
		log.Errorf("failed to parse task's status file %v: %v", f, e)
		return e
	}

	item := &task_item{
		id:     id,
		create: status.Create,
		status: status.Status,
	}

	m.task_list = append(m.task_list, item)

	return nil
}
