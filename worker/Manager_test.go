package worker

import "testing"

import "github.com/SpeedVan/go-common/config/mock"

import "time"

func Test(t *testing.T) {

	cfg := mock.New(map[string]string{})
	manager := NewManager(cfg)
	defer manager.Close()
	manager.Run()
	manager.StartWorker("manager-test", "/Users/admin/projects/go/src/github.com/SpeedVan/python-runtime/test_func/func_1.func")

	time.Sleep(20 * time.Second)
}
