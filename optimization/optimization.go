package optimization

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/cihub/seelog"
)

type (
	MiningOptParams struct {
		TaskId     string
		Notify     chan string
		InputFile  string
		OutputFile string
		ParamFile  string
	}
)

func DoMiningOptimization(opt MiningOptParams) error {

	log.Infof("Being parsing parameters from %v", opt.ParamFile)
	notifyStatus(opt.Notify, "Parsing parameters file")

	var params Parameters

	if e := readJsonFile(opt.ParamFile, &params); e != nil {
		return e
	}

	log.Infof("Begin reading input from %v", opt.InputFile)
	notifyStatus(opt.Notify, "Reading input data")

	if e := params.Input.initializeFromGzip(opt.InputFile); e != nil {
		return e
	}

	selection, status := params.optimizating(opt.Notify)

	if status != 0 {
		e := fmt.Errorf("Failed do optimization")
		log.Error(e)
		return e
	}

	var writer io.Writer
	var write_head bool
	var doclose func() error

	if len(opt.OutputFile) == 0 {
		writer = os.Stdout
		write_head = true
	} else {

		file, e := os.Create(opt.OutputFile)
		if e != nil {
			e = fmt.Errorf("Failed to create output file %v: %v", opt.OutputFile, e)
			log.Error(e)
			return e
		}
		defer file.Close()
		writer = file

		if strings.HasSuffix(opt.OutputFile, ".gz") {
			zipwriter := gzip.NewWriter(writer)
			doclose = zipwriter.Close
			writer = zipwriter
		} else {
			write_head = false
		}
	}

	if write_head {
		fmt.Fprintln(writer, "ultpit output")
		fmt.Fprintln(writer, "1")
		fmt.Fprintln(writer, "Pit")
	}

	for _, row := range selection {
		for _, v := range row {
			if v {
				fmt.Fprintln(writer, "1")
			} else {
				fmt.Fprintln(writer, "0")
			}
		}
	}

	if doclose != nil {
		doclose()
	}

	return nil
}

func notifyStatus(ch chan<- string, status string) {
	if ch != nil {
		select {
		case ch <- status:
		default:
		}
	}
}
