package main

import (
	"os"

	"github.com/SpeedVan/go-common/app/web"
	"github.com/SpeedVan/go-common/config/env"
	"github.com/SpeedVan/go-common/log"
	"github.com/SpeedVan/runtime-proxy/worker"
)

func main() {

	if cfg, err := env.LoadAllWithoutPrefix("PROXY_"); err == nil {
		logger := log.NewCommon(log.Debug)
		manager := worker.NewManager(cfg)
		manager.Run()

		app := web.New(cfg, logger)

		// app.HandleController(&controller.ProxyController{
		// 	ProxyCall: callService,
		// })

		app.RegisterOnShutdown(manager.Close)

		if err := app.Run(log.Debug); err != nil {
			os.Exit(1)
		}
	}
}
