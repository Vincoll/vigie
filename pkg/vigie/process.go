package vigie

import (
	"github.com/vincoll/vigie/pkg/ticker"
	"time"

	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// Start Vigie
// Tous les arguments sont chargés
// importingFileToVigie Brut des Fichiers TestSuite_OLD
// Preparation des TestSuites
// Création des Tickers dans des goroutines
func (v *Vigie) Start() error {

	// Importing Files in Vigie
	utils.Log.Debug("Importing Files into Vigie")
	errImport := v.importingFileToVigie()
	if errImport != nil {
		log.Fatal(errImport)
	} else {
		utils.Log.Info("All files have been loaded with success")
	}

	// registerTasksToTickerPool Read and registerTasksToTickerPool TestSuite
	utils.Log.Debug("Prepare Tests to be startEachTickerpool")
	errPrepare := v.registerTasksToTickerPool()
	if errPrepare != nil {
		log.Fatal(errPrepare)
	}

	// At this point everything is loaded in Vigie Instance
	v.Status = 1

	// startEachTickerpool Start the TestSuites
	utils.Log.Debug("Process Tests")
	errProcess := v.startEachTickerpool()
	if errProcess != nil {
		log.Fatal(errProcess)
	}

	return nil
}

// importingFileToVigie files (testsuite, vars) into Vigie
func (v *Vigie) importingFileToVigie() error {

	// Init
	v.TestSuites = map[int64]*teststruct.TestSuite{}
	v.tickerpools = map[time.Duration]*ticker.TickerPool{}

	start := time.Now()
	// Read and Initialize TestSuites Files
	// Unmarshall append here
	if err := v.loadTestFiles(); err != nil {
		return err
	}

	elapsed := time.Since(start)
	utils.Log.WithFields(log.Fields{
		"package": "vigie",
	}).Debugf("Importing files in: %s", elapsed)

	return nil
}

// Registers all TestSuites/TC/TStep as pointer into TickersPools
// Those TickersPools will run the tests and the results will be
// wrote into the Vigie Instance.
// That means v.tp[n].task.ts[1] = v.testsuite[x]
// The Goal is to limitate redondant tickers centralizing them in the vigie instance.
// Each testStep with the same duration is register to a tickerpool
func (v *Vigie) registerTasksToTickerPool() error {

	// On each TestSuites Collected
	// registerTasksToTickerPool TestCaseCount and Tickers ()

	for _, ts := range v.TestSuites {
		// Create Tickers based on TestSuites frequency
		ts2 := ts
		for _, tc := range ts2.TestCases {
			tc2 := tc
			for _, tstp := range tc2.TestSteps {
				tstp2 := tstp
				// Create/Add a new TickerPool (TP)
				if !v.getTickerPool(tstp2.ProbeWrap.Frequency) {
					// if !exists => create new tickerpool
					_ = v.createTickerPool(tstp2.ProbeWrap.Frequency)
				}
				// Add Task in tickerpool
				v.tickerpools[tstp2.ProbeWrap.Frequency].AddTask(ts2, tc2, tstp2)

			}
		}
	}

	return nil
}

// startEachTickerpool déclenche tout les Tickers afin de débuter les tests.
func (v *Vigie) startEachTickerpool() error {

	utils.Log.WithFields(log.Fields{}).Info("Start Monitoring.")

	// Go for TickerHandler
	for _, tp := range v.tickerpools {
		tp.Start()
	}

	return nil
}
