package models

import (
	"fmt"

	"github.com/yaozijian/MiningOpt/distribution"
)

func runTaskWorker(webcfg *WebConfig) error {

	rpcxcfg := distribution.RpcxServerConfig{
		ServiceAddr: fmt.Sprintf("%v:%v", webcfg.MyIpAddr, webcfg.RpcxPort),
		EtcdServers: webcfg.EtcdServers,
		URLPrefix:   fmt.Sprintf("http://%v:%v", webcfg.MyIpAddr, webcfg.HttpPort),
		Async:       true,
	}

	return distribution.StartWorker(rpcxcfg)
}
