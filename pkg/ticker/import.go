package ticker

import (
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
)

// Registers all TestSuites/TC/TStep as pointer into TickersPools
// Those TickersPools will run the tests and the results will be
// wrote into the Vigie Instance.
// That means v.tp[n].task.ts[1] = v.testsuite[x]
// The Goal is to limitate redondant concurent tickers centralizing them in the webapi instance.
// Each testStep with the same duration is register to a tickerpool
func (tpm *TickerPoolManager) ImportTS(nTS map[uint64]*teststruct.TestSuite) *TickerPoolManager {

	// On each TestSuites Collected
	// ImportTS TestCaseCount and Tickers ()

	for _, ts := range nTS {
		// Create Tickers based on TestSuites frequency
		ts2 := ts
		for _, tc := range ts2.TestCases {
			tc2 := tc
			for _, tstp := range tc2.TestSteps {
				tstp2 := tstp
				// Create/Add a new TickerPool (TP)
				freq := tstp2.ProbeWrap.Frequency

				// Create TP if needed
				if !tpm.IsTickerPool(freq) {
					// if does not exists => create new tickerpool
					err := tpm.AddTickerPool(freq)
					if err != nil {
						utils.Log.Errorf("can not create a Tickerpool: %s", err.Error())
					}
				}
				// Add Task in tickerpool

				ntask := teststruct.Task{
					TestSuite: ts2,
					TestCase:  tc2,
					TestStep:  tstp2,
				}

				tpm.AddTask(ntask)
			}
		}
	}
	return tpm
}
