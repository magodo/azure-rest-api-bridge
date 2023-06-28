package main

import (
	"context"
	"flag"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/magodo/azure-rest-api-bridge/ctrl"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
)

func main() {
	addr := flag.String("addr", "localhost", "Mock server address")
	port := flag.Int("port", 8888, "Mock server port")
	configFile := flag.String("config", "", "Execution config file")
	logLevel := flag.String("log-level", "INFO", "Log level")
	specdir := flag.String("specdir", "", "Swagger specification directory")
	index := flag.String("index", "", "Swagger index file")
	synUseEnum := flag.Bool("syn-use-enum", false, "Whether to use enum values defined in Swagger when synthesize responses")

	flag.Parse()

	logOpt := &hclog.LoggerOptions{
		Name:  "azure-rest-api-bridge",
		Level: hclog.LevelFromString(*logLevel),
		Color: hclog.AutoColor,
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
			SynthOpt: &swagger.SynthesizerOption{
				UseEnumValues: *synUseEnum,
			},
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
