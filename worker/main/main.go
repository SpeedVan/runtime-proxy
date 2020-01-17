package main

import (
	"github.com/SpeedVan/go-common/config/mock"
	"github.com/SpeedVan/runtime-proxy/worker"
)

func main() {
	cfg := mock.New(map[string]string{})
	manager := worker.New(cfg)

	manager.StartWorker("manager-test", "/Users/admin/projects/go/src/github.com/SpeedVan/python-runtime/test_func/func_1.func")
}
