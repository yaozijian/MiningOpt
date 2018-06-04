package distribution

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/smallnest/rpcx"
	"github.com/smallnest/rpcx/plugin"
)

type (
	RpcxServerConfig struct {
		ServiceAddr string
		EtcdServers []string
		URLPrefix   string
		Async       bool
	}

	newServerFunc func(RpcxServerConfig) interface{}
)

const (
	rpcx_base_path      = "/miningopt"
	rpcx_worker_name    = "worker"
	rpcx_manager_name   = "manager"
	rpcx_manager_notify = "manager.Notify"
	rpcx_worker_method  = "worker.MiningOpt"
)

const (
	task_data_dir    = "tasks"
	task_data_file   = "data.gz"
	task_param_file  = "param.json"
	task_result_file = "result.gz"
)

func StartWorker(cfg RpcxServerConfig) error {
	if startRpcxServer(rpcx_worker_name, cfg, NewWorker) == nil {
		return fmt.Errorf("Error to start worker")
	} else {
		return nil
	}
}

func StartManager(cfg RpcxServerConfig) Manager {
	srvobj := startRpcxServer(rpcx_manager_name, cfg, NewManager)
	if srvobj != nil {
		return srvobj.(Manager)
	} else {
		return nil
	}
}

func startRpcxServer(name string, cfg RpcxServerConfig, newobj newServerFunc) interface{} {

	etcdplugin := &plugin.EtcdV3RegisterPlugin{
		ServiceAddress:      fmt.Sprintf("tcp4@%v", cfg.ServiceAddr),
		EtcdServers:         cfg.EtcdServers,
		BasePath:            rpcx_base_path,
		DialTimeout:         10 * time.Second,
		UpdateIntervalInSec: 10,
	}

	if err := etcdplugin.Start(); err != nil {
		log.Errorf("Error to start etcd register plugin: %v", err)
		return nil
	}

	server := rpcx.NewServer()
	server.PluginContainer.Add(etcdplugin)

	srvobj := newobj(cfg)

	server.RegisterName(name, srvobj)

	log.Infof("Begin %v at %v", name, cfg.ServiceAddr)

	if cfg.Async {
		go func() {
			err := server.Serve("tcp4", cfg.ServiceAddr)
			log.Infof("End %v at %v,err=%v", name, cfg.ServiceAddr, err)
		}()
	} else {
		err := server.Serve("tcp4", cfg.ServiceAddr)
		log.Infof("End %v at %v,err=%v", name, cfg.ServiceAddr, err)
	}

	return srvobj
}
