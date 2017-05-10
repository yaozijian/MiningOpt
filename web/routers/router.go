package routers

import (
	"github.com/astaxie/beego"
	"github.com/yaozijian/MiningOpt/web/controllers"
	"github.com/yaozijian/MiningOpt/web/models"
)

func Init(cfg *models.WebConfig) {
	if cfg.StartType == models.WebType_Manager {
		beego.Router("/", &controllers.MainController{})
	}
}
