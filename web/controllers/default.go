package controllers

import (
	"encoding/hex"
	"fmt"
	"os"

	"crypto/rand"

	"github.com/astaxie/beego"
	"github.com/yaozijian/MiningOpt/web/models"
)

type MainController struct {
	beego.Controller
	manager models.TaskManager
}

func (c *MainController) Prepare() {
	c.manager = models.GetTaskManager()
}

func (c *MainController) Get() {
	c.Data["tasklist"] = c.manager.TaskList()
	c.Data["serverlist"] = c.manager.ServerList()
	c.TplName = "index.tpl"
}

func (c *MainController) Post() {

	id, e := c.saveFile("", "data-file", models.Task_data_file)

	if e == nil {
		if _, e = c.saveFile(id, "param-file", models.Task_param_file); e == nil {
			e = c.manager.NewTask(id)
		}
	}

	if e != nil {
		if len(id) > 0 {
			os.Remove(models.Task_data_dir + "/" + id)
		}
		c.Data["error"] = e.Error()
		c.Data["msg"] = ""
	} else {
		c.Data["msg"] = "New Task OK"
		c.Data["error"] = ""
	}

	c.TplName = "newtask.tpl"
}

func (c *MainController) saveFile(id, field, name string) (string, error) {

	f, _, e := c.GetFile(field)

	defer func() {
		if e != nil {
			c.Data["error"] = fmt.Sprint(e)
		}
	}()

	if e != nil {
		return "", e
	}
	defer f.Close()

	if len(id) == 0 {
		id = c.newTaskId()
		if e = os.MkdirAll(models.Task_data_dir+"/"+id, 0777); e != nil {
			return "", e
		}
	}

	if e = c.SaveToFile(field, fmt.Sprintf("%v/%v/%v", models.Task_data_dir, id, name)); e != nil {
		return "", e
	} else {
		return id, nil
	}
}

func (c *MainController) newTaskId() string {
	buf := make([]byte, models.Task_id_len)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}
