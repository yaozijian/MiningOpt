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
	"fmt"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yaozijian/MiningOpt/optimization"
)

const (
	log_cfg_tmpl = `<seelog minlevel="info">
		<outputs formatid="detail">
			{{OutputDest}}
		</outputs>
		<formats>
			<format id="detail" format="[%File:%Line][%Date(2006-01-02 15:04:05.000)] %Msg%n" />
		</formats>
	</seelog>`
	log_out_dest  = "{{OutputDest}}"
	log_file_tmpl = `<rollingfile filename="%s" type="size" maxsize="10247680" maxrolls="10"/>`
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a mining optimization task",
	Long:  "run a mining optimization task",
	Run: func(cmd *cobra.Command, args []string) {
		doMiningOperation(cmd, args)
	},
}

func init() {

	RootCmd.AddCommand(runCmd)

	flagset := runCmd.PersistentFlags()
	flagset.StringP("input", "i", "", "The input file")
	flagset.StringP("output", "o", "", "The output file")
	flagset.StringP("log", "l", "", "Log information to a file")
}

func doMiningOperation(cmd *cobra.Command, args []string) {

	viper.BindPFlags(cmd.Flags())

	logfile := viper.GetString("log")
	infile := viper.GetString("input")
	outfile := viper.GetString("output")

	if len(infile) == 0 || len(outfile) == 0 || len(args) != 1 {
		cmd.Usage()
		return
	}

	//-------

	outputDest := "<console/>"

	if len(logfile) > 0 {
		outputDest = fmt.Sprintf(log_file_tmpl, logfile)
	}

	log_cfg := strings.Replace(log_cfg_tmpl, log_out_dest, outputDest, -1)

	logger, _ := log.LoggerFromConfigAsString(log_cfg)

	if logger != nil {
		log.ReplaceLogger(logger)
	}

	//-------

	param := optimization.MiningOptParams{
		InputFile:  infile,
		OutputFile: outfile,
		ParamFile:  args[0],
	}

	log.Info("ultpit begin")

	optimization.DoMiningOptimization(param)

	log.Info("ultpit finished")
	log.Flush()
}
