package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/cihub/seelog"
)

func (t *task_item) Id() string {
	return t.id
}

func (t *task_item) Create() time.Time {
	return t.create
}

func (t *task_item) Server() string {
	t.Lock()
	defer t.Unlock()
	return t.server
}

func (t *task_item) Status() string {
	t.Lock()
	defer t.Unlock()
	if len(t.desc) == 0 {
		return t.status
	} else {
		return fmt.Sprintf("%v: %v", t.status, t.desc)
	}
}

func (t *task_item) DataURL() (data string) {
	data = fmt.Sprintf("/%v/%v/%v", Task_data_dir, t.id, Task_data_file)
	return
}

func (t *task_item) ParamURL() (param string) {
	param = fmt.Sprintf("/%v/%v/%v", Task_data_dir, t.id, Task_param_file)
	return
}

func (t *task_item) ResultURL() string {

	t.Lock()
	defer t.Unlock()

	if t.status == task_status_OK {
		return fmt.Sprintf("/%v/%v/%v", Task_data_dir, t.id, Task_result_file)
	} else {
		return ""
	}
}

func (t *task_item) Encode2Json() string {
	x := map[string]string{
		"Id":        t.id,
		"Status":    t.Status(),
		"Server":    t.Server(),
		"ResultURL": t.ResultURL(),
	}
	c, _ := json.Marshal(x)
	return string(c)
}

func (t *task_item) downloadResultFile() error {

	body, err := http.Get(t.outurl)

	if err != nil {
		err = fmt.Errorf("Failed to download result file from %v for task %v: %v", t.outurl, t.id, err)
	} else if body == nil || body.Body == nil {
		err = fmt.Errorf("Failed to download result file from %v for task %v: empty reply body", t.outurl, t.id)
	} else {
		defer body.Body.Close()

		filename := fmt.Sprintf("%v/%v/%v", Task_data_dir, t.id, Task_result_file)

		if file, err := os.Create(filename); err != nil {
			err = fmt.Errorf("Error to create result file %v for task %v: %v", filename, t.id, err)
		} else if _, err = io.Copy(file, body.Body); err != nil {
			file.Close()
			os.Remove(filename)
			err = fmt.Errorf("Error to write result file %v for task %v: %v", filename, t.id, err)
		} else {
			file.Close()
			log.Infof("Download result file for task %v from %v: OK", t.id, t.outurl)
			err = nil
		}
	}

	t.Lock()
	defer t.Unlock()

	t.desc = ""

	if err == nil {
		t.status = task_status_OK
	} else {
		t.status = task_status_Download_Failed
		log.Error(err)
	}

	return err
}
