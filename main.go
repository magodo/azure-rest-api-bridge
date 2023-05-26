package main

import (
	"context"
	"flag"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/magodo/azure-rest-api-bridge/ctrl"
	"github.com/magodo/azure-rest-api-bridge/log"
)

func main() {
	addr := flag.String("addr", "localhost", "Mock server address")
	port := flag.Int("port", 8888, "Mock server port")
	specFile := flag.String("spec", "", "Execution spec file")
	verbose := flag.Bool("verbose", false, "Show debug log")

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
		MockSrvAddr: *addr,
		MockSrvPort: *port,
		ExecFile:    *specFile,
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
