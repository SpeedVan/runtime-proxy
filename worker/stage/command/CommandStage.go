package command

import (
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/SpeedVan/runtime-proxy/worker/stage"
)

// StageImpl todo
type StageImpl struct {
	stage.Stage
	NextStage   stage.Stage
	ForkCommand string
	Process     *os.Process
}

// New todo
func New(name, command string) *StageImpl {
	lock := &sync.Mutex{}
	lock.Lock()
	return &StageImpl{
		ForkCommand: command,
	}
}

// Do todo
func (s *StageImpl) Do() error {
	command := s.ForkCommand
	cmd := strings.Split(command, " ")

	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, //其他变量如果不清楚可以不设定
	}

	p, err := os.StartProcess(cmd[0], cmd, attr)

	if err != nil {
		return err
	}

	s.Process = p

	return nil
}

// Close todo
func (s *StageImpl) Close() error {
	if s.Process != nil {
		err := s.Process.Signal(syscall.SIGKILL)
		time.Sleep(1 * time.Second)
		return err
	}
	return nil
}

// Next todo
func (s *StageImpl) Next() stage.Stage {
	return s.NextStage
}
