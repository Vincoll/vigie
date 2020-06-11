package tsdb

import (
	"github.com/vincoll/vigie/pkg/teststruct"
	"sync"
)

// Global Var for now, to avoid time-consuming modifications
// in case of a change of code rearchitecture.
// TSDB state will be injected later
var TsdbManager Tsdbs

type Tsdbs struct {
	mu            sync.RWMutex
	Enabled       bool
	TsdbEndpoints []TsdbEndpoint
}

func (ts *Tsdbs) AddTsdb(endpoint TsdbEndpoint) {

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.TsdbEndpoints = append(ts.TsdbEndpoints, endpoint)
	ts.Enabled = true

	return
}

func (ts *Tsdbs) WriteOnTsdbs(task teststruct.Task) {

	var wg sync.WaitGroup
	wg.Add(len(ts.TsdbEndpoints))
	for _, tsdbEndpoint := range ts.TsdbEndpoints {

		go func(te TsdbEndpoint) {
			_ = te.WritePoint(task)
			wg.Done()
		}(tsdbEndpoint)

	}
	wg.Wait()

	return
}

func (ts *Tsdbs) UpdateTestStateToDB(task teststruct.Task) {
	var wg sync.WaitGroup
	wg.Add(len(ts.TsdbEndpoints))
	for _, tdbedpt := range ts.TsdbEndpoints {

		go func(te TsdbEndpoint) {
			_ = te.UpdateTestState(task)
			wg.Done()
		}(tdbedpt)

	}
	wg.Wait()
}
