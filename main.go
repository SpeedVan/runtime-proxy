package main

import (
	"encoding/json"
	"os"
	"strconv"
	"syscall"

	"github.com/SpeedVan/go-common/app/web"
	"github.com/SpeedVan/go-common/client/httpclient"
	"github.com/SpeedVan/go-common/config/env"
	"github.com/SpeedVan/go-common/log"
	lc "github.com/SpeedVan/go-common/log/common"
	"github.com/SpeedVan/runtime-proxy/consumer"
	"github.com/SpeedVan/runtime-proxy/controller"
	"github.com/SpeedVan/runtime-proxy/service/localhttpcall"
)

func main() {

	funcConfig := make(map[string]map[string]string)
	json.Unmarshal([]byte(os.Getenv("FUNC_CONFIG")), &funcConfig)
	streams := []string{}
	for k := range funcConfig {
		streams = append(streams, k)
	}

	if cfg, err := env.LoadAllWithoutPrefix("PROXY_"); err == nil {
		logger := lc.NewCommon(log.Debug)

		command := cfg.Get("FORK_COMMAND")
		cmdArg := cfg.Get("FORK_COMMAND_ARG")
		logger.Debug("Run Command:" + command + " " + cmdArg)
		attr := &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, //其他变量如果不清楚可以不设定
		}
		p, err := os.StartProcess(command, []string{command, cmdArg}, attr)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		defer func() {
			p.Signal(syscall.SIGTERM) //kill process
		}()

		hc, _ := httpclient.New(cfg)

		callService := localhttpcall.New(cfg.Get("PORT"), hc)

		// EVENTSTORE_ENDPOINT e.g. "tcp://admin:changeit@10.10.139.35:1113"
		callConsumer, err := consumer.New("node_1", cfg.Get("EVENTSTORE_ENDPOINT"), callService, streams)

		callConsumer.Subscribe()

		// feedback, err := exec.SimpleExec(command, cmdArg)

		logger.Debug("Run Command Pid:" + strconv.Itoa(p.Pid))

		app := web.New(cfg, logger)

		app.HandleController(&controller.ProxyController{
			ProxyCall: callService,
		})

		app.RegisterOnShutdown(func() {
			callConsumer.Close() //kill process
		})

		if err := app.Run(log.Debug); err != nil {
			os.Exit(1)
		}
	}
}
