package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/magodo/azure-rest-api-bridge/ctrl"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver"
)

func main() {
	addr := flag.String("addr", "localhost", "Mock server address")
	port := flag.Int("port", 8888, "Mock server port")
	configFile := flag.String("config", "", "Execution config file")
	logLevel := flag.String("log-level", "INFO", "Log level")
	specdir := flag.String("specdir", "", "Swagger specification directory")
	index := flag.String("index", "", "Swagger index file")
	continueOnErr := flag.Bool("k", false, "Whether to continue on error")
	execFrom := flag.String("from", "", "Run execution from the specified one (inclusively), in form of `name.type`")
	execTo := flag.String("to", "", "Run execution until the specified one (exclusively), in form of `name.type`")
	timeout := flag.Int("timeout", 60, "The mock server read/write timeout in second")

	flag.Parse()

	logOpt := &hclog.LoggerOptions{
		Name:  "azure-rest-api-bridge",
		Level: hclog.LevelFromString(*logLevel),
		Color: hclog.AutoColor,
	}
	logger := hclog.New(logOpt)
	log.SetLogger(logger)

	ctrl, err := ctrl.NewCtrl(ctrl.Option{
		ConfigFile:    *configFile,
		ContinueOnErr: *continueOnErr,
		ServerOption: mockserver.Option{
			Addr:    *addr,
			Port:    *port,
			Index:   *index,
			SpecDir: *specdir,
			Timeout: time.Duration(*timeout) * time.Second,
		},
		ExecFrom: *execFrom,
		ExecTo:   *execTo,
	})
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err := ctrl.Run(context.TODO()); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
