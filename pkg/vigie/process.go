package vigie

import (
	"time"

	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/ticker"
	"github.com/vincoll/vigie/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// Start Vigie
// Tous les arguments sont chargés
// importingFileToVigie Brut des Fichiers TestSuite_OLD
// Preparation des TestSuites
// Création des Tickers dans des goroutines
func (v *Vigie) Start() error {

	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"desc":    "Vigie is starting",
	}).Debugf("This Vigie is starting")

	v.setStatus("Starting")

	// Listen for any new tests
	v.receiveTests()

	// Chan ImportManager => Vigie
	v.ImportManager.OutgoingTests = v.incomingTests

	// Is Vigie HA ?
	if v.ConsulClient == nil {

		v.ImportManager.Start()

	} else {

		// Vigie with Consul
		// Startup will depend on the status of this Vigie regarding others Vigie
		// registered in Consul.
		// Only a Leader is allowed to load the TestFiles and schedule tests.
		// Followers will watched for the "scheduling file" stored in Consul
		// and pull tests from Consul K:V

		v.ImportManager.ConsulClient = v.ConsulClient

		v.setStatus("Waiting for Leader Election result")

		err := v.loadAndPushConsul()

		// Gestion du super sheduler ...

	}

	return nil
}

func (v *Vigie) _Start0() error {

	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"desc":    "Vigie is starting",
	}).Debugf("This Vigie is starting")
	v.setStatus("Starting")

	// Want for any

	// Is Vigie HA ?
	if v.ConsulClient == nil {

		// Vigie without Consul
		err := v.loadAndRun()
		if err != nil {
			utils.Log.Errorf("Error while loading TestSuites: %s", err)
		} else {
			utils.Log.Infof("All files have been loaded with success")
		}

		// SET the Testfile Reloader
		if v.ImportManager.Frequency != 0 {
			go v.activateConfigReloader()
		}
	} else {

		// Vigie with Consul
		// Startup will depend on the status of this Vigie regarding others Vigie
		// registered in Consul.
		// Only a Leader is allowed to load the TestFiles and schedule tests.
		// Followers will watched for the "scheduling file" stored in Consul
		// and pull tests from Consul K:V

		v.setStatus("Waiting for Leader Election result")

		err := v.loadAndPushConsul()
		if err != nil {
			utils.Log.Errorf("Error while loading TestSuites: %s", err)
		} else {
			utils.Log.Infof("All files have been loaded with success")
		}

	}

	// At this point everything is loaded and running in Vigie Instance
	v.setStatus("Ready")

	return nil
}

func (v *Vigie) setStatus(s string) {
	v.mu.Lock()
	v.Status = s
	v.mu.Unlock()
}

// receiveTests listen for incoming tests

func (v *Vigie) receiveTests() {

	utils.Log.Infof("Vigie is waiting for test to be load")
	v.setStatus("Ready")

	go func() {

		for {
			select {

			case allTestSuites := <-v.incomingTests:

				err := v.loadAndRun2(allTestSuites)
				if err != nil {
					utils.Log.Errorf("Error while loading TestSuites: %s", err)
				} else {
					utils.Log.Infof("All files have been loaded with success")
				}

			}
		}
	}()
}

func (v *Vigie) swapStateAndRun(newTSs map[uint64]*teststruct.TestSuite, tpm *ticker.TickerPoolManager) {

	// Lock on Vigie has been made by the parent func.
	utils.Log.Debug("Swap OLD / NEW TSs and TP")

	// Stop and close Old Tickers Goroutines to avoid leak.
	v.TickerPoolManager.StopEachTickerPool()
	v.TestSuites = newTSs
	v.TickerPoolManager = tpm
	// (Re) initiate the tickers pools
	v.TickerPoolManager.StartEachTickerPool()
	return
}

// activateConfigReloader load and generates new TestSuites from the TestFiles
func (v *Vigie) activateConfigReloader() {

	utils.Log.Infof("Vigie will reload it state every %s", v.ImportManager.Frequency)

	importTicker := time.NewTicker(v.ImportManager.Frequency)

	for {
		select {
		case <-importTicker.C:
			err := v.loadAndRun()
			if err != nil {
				utils.Log.Errorf("Error while loading TestSuites: %s", err)
			}
		}
	}

}

// Registers all TestSuites/TC/TStep as pointer into TickersPools
// Those TickersPools will run the tests and the results will be
// wrote into the Vigie Instance.
// That means v.tp[n].task.ts[1] = v.testsuite[x]
// The Goal is to limitate redondant concurent tickers centralizing them in the vigie instance.
// Each testStep with the same duration is register to a tickerpool
func (v *Vigie) createTickerPools(nTS map[uint64]*teststruct.TestSuite) *ticker.TickerPoolManager {

	// On each TestSuites Collected
	// createTickerPools TestCaseCount and Tickers ()

	TPMngr := ticker.NewTickerPoolManager(v.TickerPoolManager.ChanToSched)

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
				if !TPMngr.IsTickerPool(freq) {
					// if does not exists => create new tickerpool
					err := TPMngr.AddTickerPool(freq)
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

				TPMngr.AddTask(ntask)
			}
		}
	}
	return TPMngr
}
