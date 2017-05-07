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
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run web server and/or worker server",
	Long:  "Run web server and/or worker server",
	Run: func(cmd *cobra.Command, args []string) {
		run_server(cmd, args)
	},
}

func init() {

	RootCmd.AddCommand(serverCmd)

	flagset := serverCmd.PersistentFlags()
	flagset.BoolP("ui", "u", false, "Run web server")
	flagset.BoolP("worker", "w", false, "Run worker server")
}

func run_server(cmd *cobra.Command, args []string) {

	viper.BindPFlags(cmd.Flags())

	ui := viper.GetBool("ui")
	worker := viper.GetBool("worker")

	if !ui && !worker {
		cmd.Usage()
		return
	}

	if ui {
		web.Run()
	}
}
