package web

import (
	"github.com/astaxie/beego"
	"github.com/yaozijian/MiningOpt/web/models"
	_ "github.com/yaozijian/MiningOpt/web/routers"
)

func Run() {

	models.Init()

	beego.SetStaticPath("/tasks", "tasks")

	beego.Run()
}
