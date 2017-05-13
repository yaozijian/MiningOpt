package web

import (
	"fmt"
	"net"
	"strings"

	"github.com/astaxie/beego"
	log "github.com/cihub/seelog"
	"github.com/yaozijian/MiningOpt/web/models"
	"github.com/yaozijian/MiningOpt/web/routers"
)

const (
	console_log = `<seelog>
	<outputs formatid="detail">
		<console/>
	</outputs>
	<formats>
		<format id="detail" format="[%Date(2006-01-02 15:04:05.000)][%Level][%File:%Line] %Msg%n" />
	</formats>
</seelog>
`
)

func initlog() {
	logger, _ := log.LoggerFromConfigAsString(console_log)
	log.ReplaceLogger(logger)
}

func Run(cfg models.WebConfig) {

	defer log.Flush()

	initlog()

	cfg.MyIpAddr = getMyIpAddr()

	if len(cfg.MyIpAddr) == 0 {
		return
	}

	routers.Init(&cfg)

	models.Init(&cfg)

	beego.SetStaticPath("/"+models.Task_data_dir, models.Task_data_dir)

	beego.Run(fmt.Sprintf(":%v", cfg.HttpPort))
}

func getMyIpAddr() string {

	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {

		item := addr.String()

		if idx := strings.Index(item, "/"); idx >= 0 {
			item = item[:idx]
		}

		ip := net.ParseIP(item)

		if ip != nil && ip.To4() != nil &&
			!ip.IsLoopback() && !ip.IsMulticast() {
			item = ip.To4().String()
			log.Infof("My IP address is %v", item)
			return item
		}
	}

	log.Error("Failed to get my IP address")

	return ""
}
