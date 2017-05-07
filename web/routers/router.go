package routers

import (
	"github.com/astaxie/beego"
	"github.com/yaozijian/MiningOpt/web/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
}
