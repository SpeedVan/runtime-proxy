package worker

import (
	"fmt"
	"net/http"

	"github.com/SpeedVan/runtime-proxy/worker/stage/command"
	"github.com/SpeedVan/runtime-proxy/worker/stage/consume"
)

// FunctionWorker todo
type FunctionWorker struct {
	CommandStage   *command.StageImpl
	ConsumeStage   *consume.StageImpl
	passChan       chan string
	StageCloseFunc []func() error
	Port           string
}

// NewWorkder todo
func NewWorkder(forkEntrypoint, funcName, port, entrypoint, freebackCode string, httpclient *http.Client) *FunctionWorker {
	// funcName entrypoint freebackCode -> command
	// entrypoint e.g. "/Users/admin/projects/go/src/github.com/SpeedVan/python-runtime/test_func/func_1.func"
	cmd := fmt.Sprintf(forkEntrypoint+" {\"bind\":\"0.0.0.0:%v\",\"ENTRYPOINT\":\"%v\",\"FREEBACK_CODE\":\"%v\"}", port, entrypoint, freebackCode)
	// cmd = cfg.GetString("APP_CMD", "/usr/local/bin/python3 /app/app.py")
	streamName := funcName

	commandStage := command.New(funcName, cmd)

	consumeStage, _ := consume.New(streamName, port, httpclient)

	return &FunctionWorker{
		CommandStage:   commandStage,
		ConsumeStage:   consumeStage,
		passChan:       make(chan string),
		StageCloseFunc: []func() error{},
	}
}

// Pass todo
func (s *FunctionWorker) Pass() {
	s.passChan <- ""
}

// Start todo
func (s *FunctionWorker) Start() {
	err := s.CommandStage.Do()
	if err != nil {
		fmt.Println(err.Error())
	}
	s.StageCloseFunc = append(s.StageCloseFunc, s.CommandStage.Close)
	<-s.passChan
	err = s.ConsumeStage.Do()
	s.StageCloseFunc = append(s.StageCloseFunc, s.ConsumeStage.Close)
}

// Close todo
func (s *FunctionWorker) Close() error {
	close(s.passChan)
	for _, item := range s.StageCloseFunc {
		err := item()
		if err != nil {
			fmt.Println("stage close error:" + err.Error())
		}
	}
	return nil
}
