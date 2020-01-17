package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	uuid "github.com/satori/go.uuid"

	"github.com/SpeedVan/go-common/client/httpclient"
	"github.com/SpeedVan/go-common/config"
	"github.com/SpeedVan/go-common/net/netcat"
)

// Manager todo
type Manager struct {
	ForkEntrypoint string
	IncPort        int
	FuncConfig     map[string]string
	sm             *sync.Map
	Netcat         *netcat.Netcat
	HTTPClient     *http.Client
}

// NewManager todo
func NewManager(cfg config.Config) *Manager {

	funcConfig := make(map[string]string)
	json.Unmarshal([]byte(cfg.Get("FUNC_CONFIG")), &funcConfig)

	sm := &sync.Map{}
	nc := &netcat.Netcat{
		AddressAndPort: "0.0.0.0:2018",
		WhenAccept: func(code string) {
			fmt.Println("get msg:" + code)
			if v, ok := sm.Load(code); ok {
				if w, ok := v.(*FunctionWorker); ok {
					w.Pass()
				}
			}
		},
	}
	httpclient, _ := httpclient.New(cfg)
	return &Manager{
		ForkEntrypoint: cfg.Get("FORK_ENTRYPOINT"),
		IncPort:        cfg.GetInt("IncPort", 5000),
		FuncConfig:     funcConfig,
		sm:             sm,
		Netcat:         nc,
		HTTPClient:     httpclient,
	}
}

// Run todo
func (s *Manager) Run() {
	go s.Netcat.Run()
	for k, v := range s.FuncConfig {
		s.StartWorker(k, v)
	}
}

// StartWorker todo
func (s *Manager) StartWorker(name, entrypoint string) (int, error) {
	uid, _ := uuid.NewV4()
	freebackCode := uid.String()
	port := strconv.Itoa(s.IncPort)
	s.IncPort++
	worker := NewWorkder(s.ForkEntrypoint, name, port, entrypoint, freebackCode, s.HTTPClient)
	// worker.CommandStage.Lock()
	s.sm.Store(freebackCode, worker)
	go worker.Start()
	return 0, nil
}

// Close todo
func (s *Manager) Close() {
	s.Netcat.Close()
	s.sm.Range(func(k, w interface{}) bool {
		worker := w.(*FunctionWorker)
		err := worker.Close()
		if err != nil {
			fmt.Println("worker close error:" + err.Error())
		}
		return true
	})
}
