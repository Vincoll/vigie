package tsdb

import (
	"github.com/vincoll/vigie/pkg/teststruct"
	"sync"
)

// Global Var for now, to avoid time-consuming modifications
// in case of a change of code rearchitecture.
// TSDB state will be injected later
var TsdbMgr Manager

type Manager struct {
	mu            sync.RWMutex
	Enabled       bool
	TsdbEndpoints []TsdbEndpoint    // TsdbEndpoint are all interfaces
	Tags          map[string]string // Vigie host tags
	chWrite       chan teststruct.Task
}

func (tsmgr *Manager) AddTsdb(endpoint TsdbEndpoint) {

	tsmgr.mu.Lock()
	defer tsmgr.mu.Unlock()

	tsmgr.TsdbEndpoints = append(tsmgr.TsdbEndpoints, endpoint)
	tsmgr.Enabled = true

	return
}

// WriteOnTsdbs write on every configured TSDB
func (tsmgr *Manager) WriteOnTsdbs(task teststruct.Task, vigieResult *teststruct.VigieResult) {

	var wg sync.WaitGroup
	wg.Add(len(tsmgr.TsdbEndpoints))
	for _, tsdbEndpoint := range tsmgr.TsdbEndpoints {

		go func(te TsdbEndpoint) {
			_ = te.WritePoint(task, vigieResult, tsmgr.Tags)
			wg.Done()
		}(tsdbEndpoint)

	}
	wg.Wait()

	return
}

func (tsmgr *Manager) UpdateTestStateToDB(task teststruct.Task) {
	var wg sync.WaitGroup
	wg.Add(len(tsmgr.TsdbEndpoints))
	for _, tdbedpt := range tsmgr.TsdbEndpoints {

		go func(te TsdbEndpoint) {
			_ = te.UpdateTestState(task)
			wg.Done()
		}(tdbedpt)

	}
	wg.Wait()
}
