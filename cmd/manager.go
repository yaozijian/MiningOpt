// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yaozijian/MiningOpt/web"
	"github.com/yaozijian/MiningOpt/web/models"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Run manager with a web ui to manage tasks",
	Long:  "Run manager with a web ui to manage tasks",
	Run: func(cmd *cobra.Command, args []string) {
		runManager(cmd, args)
	},
}

func init() {

	RootCmd.AddCommand(managerCmd)

	flagset := managerCmd.PersistentFlags()
	flagset.StringSliceP("etcd-servers", "e", nil, "Etcd servers used for coordinate with workers")
	flagset.Uint16P("http-port", "w", 8080, "Http listen port")
	flagset.Uint16P("rpcx-port", "r", 9527, "Rpcx listen port")
}

func runManager(cmd *cobra.Command, args []string) {
	runWeb(cmd, args, models.WebType_Manager)
}

func runWeb(cmd *cobra.Command, args []string, typ models.WebType) {

	viper.BindPFlags(cmd.Flags())

	etcd_servers := viper.GetStringSlice("etcd-servers")
	http_port := viper.GetInt("http-port")
	rpcx_port := viper.GetInt("rpcx-port")

	if len(etcd_servers) == 0 || rpcx_port < 0 || rpcx_port > 65535 || http_port < 0 || http_port > 65535 {
		cmd.Usage()
	} else {
		cfg := models.WebConfig{
			EtcdServers: etcd_servers,
			HttpPort:    uint16(http_port),
			RpcxPort:    uint16(rpcx_port),
			StartType:   typ,
		}
		web.Run(cfg)
	}
}
