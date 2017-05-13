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
	"github.com/yaozijian/MiningOpt/web/models"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run worker to do tasks assigned by manager",
	Long:  "Run worker to do tasks assigned by manager",
	Run: func(cmd *cobra.Command, args []string) {
		runWorker(cmd, args)
	},
}

func init() {

	RootCmd.AddCommand(workerCmd)

	flagset := workerCmd.PersistentFlags()
	flagset.StringSliceP("etcd-servers", "e", nil, "Etcd servers used for coordinate with manager")
	flagset.Uint16P("http-port", "w", 8081, "Http listen port")
	flagset.Uint16P("rpcx-port", "r", 9528, "Rpcx listen port")
}

func runWorker(cmd *cobra.Command, args []string) {
	runWeb(cmd, args, models.WebType_Worker)
}
