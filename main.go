package main

import (
	"context"
	"flag"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/magodo/azure-rest-api-bridge/ctrl"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver"
)

func main() {
	addr := flag.String("addr", "localhost", "Mock server address")
	port := flag.Int("port", 8888, "Mock server port")
	configFile := flag.String("config", "", "Execution config file")
	verbose := flag.Bool("verbose", false, "Be verbose")
	specdir := flag.String("specdir", "", "Swagger specification directory")
	index := flag.String("index", "", "Swagger index file")

	flag.Parse()

	logOpt := &hclog.LoggerOptions{
		Name:  "azure-rest-api-bridge",
		Level: hclog.Info,
		Color: hclog.AutoColor,
	}
	if *verbose {
		logOpt.Level = hclog.Debug
	}
	logger := hclog.New(logOpt)
	log.SetLogger(logger)

	ctrl, err := ctrl.NewCtrl(ctrl.Option{
		ConfigFile: *configFile,
		ServerOption: mockserver.Option{
			Addr:    *addr,
			Port:    *port,
			Index:   *index,
			SpecDir: *specdir,
		},
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
