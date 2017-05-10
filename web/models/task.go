package models

import (
	"fmt"
	"time"
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
	if t.server != nil {
		return t.server.Address()
	} else {
		return ""
	}
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

	if t.status == Task_Done_OK {
		return fmt.Sprintf("/%v/%v/%v", Task_data_dir, t.id, Task_result_file)
	} else {
		return ""
	}
}
