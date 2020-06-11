package vigie

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
	"strconv"
	"time"
)

// loadAndRun, but keep current state
// 1. Load TestFiles
// 2. Prepare Tickers
// 3. Swap Testfiles
// 4. Start or Keep Going
func (v *Vigie) loadAndRun() error {

	// Read and Initialize TestSuites Files
	// Unmarshall append here
	// Importing Files in Vigie
	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"role":    "load testsuite and generate scheduling",
	}).Info("(Re)Load Testsuite and generate scheduling")

	newTSs, err := v.ImportManager.LoadTestSuites()
	if err != nil {
		return err
	} else {
		utils.Log.WithFields(log.Fields{
			"package": "vigie",
			"role":    "load Testsuite and generate scheduling",
		}).Debug("All the TestFiles have been unmarshalled with success")
	}

	// LOCK Vigie
	v.mu.Lock()
	defer v.mu.Unlock()

	start := time.Now()
	//
	// Import a new set of Testsuites
	//
	importedTS, anyChanges := v.ImportAllTestSuites(newTSs)
	if anyChanges == false {
		// No changes, no need to swap, keep the Vigie state as it is
		utils.Log.WithFields(log.Fields{
			"package": "vigie",
			"desc":    "load Testsuite and generate scheduling",
			"type":    "info",
		}).Infof("No changes detected")

		return nil
	}
	//
	// Create new TickerPools
	//
	TPools := v.createTickerPools(importedTS)
	//utils.Log.Debug("New Tickerpool generated")

	elapsed := time.Since(start)
	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"desc":    "load Testsuite and generate scheduling",
		"type":    "perf_measurement",
		"value":   elapsed.Seconds(),
	}).Debugf("TOTAL Load Testsuite and generate scheduling duration: %s", elapsed)

	// Now that TS and TickerPools are set, we need to
	// swap the old and running Vigie state by the new state.

	v.swapStateAndRun(importedTS, TPools)

	return nil
}

// loadAndPushConsul
func (v *Vigie) loadAndPushConsul() error {

	// Read and Initialize TestSuites Files
	// Unmarshall append here
	start := time.Now()
	// Importing Files in Vigie
	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"role":    "load testsuite and generate scheduling",
	}).Info("(Re)Load Testsuite and generate scheduling")

	newTSs, err := v.ImportManager.LoadTestSuites()
	if err != nil {
		return err
	} else {
		utils.Log.WithFields(log.Fields{
			"package": "vigie",
			"role":    "load Testsuite and generate scheduling",
		}).Debug("All the TestFiles have been unmarshalled with success")
	}

	err = v.pushTestsToConsul(newTSs)
	if err != nil {
		return fmt.Errorf("fail to push Testsuites into Consul")
	}

	elapsed := time.Since(start)
	utils.Log.WithFields(log.Fields{
		"package": "vigie",
		"desc":    "load Testsuite and generate scheduling",
		"type":    "perf_measurement",
		"value":   elapsed.Seconds(),
	}).Debugf("TOTAL Load Testsuite and generate scheduling duration: %s", elapsed)

	return nil
}

// pushTestsToConsul
func (v *Vigie) pushTestsToConsul(TSs map[uint64]*teststruct.TestSuite) error {

	// Get a handle to the KV API
	csl := v.ConsulClient.Consul.KV()

	for _, ts := range TSs {

		tsid := strconv.FormatUint(ts.ID, 10)
		tsPath := fmt.Sprintf("vigie/%s/value", tsid)

		kvTS := &consul.KVPair{Key: tsPath, Value: ts.ToConsul()}

		_, err := csl.Put(kvTS, nil)
		if err != nil {
			return err
		}

		for _, tc := range ts.TestCases {

			tcid := strconv.FormatUint(tc.ID, 10)
			tcPath := fmt.Sprintf("vigie/%s/%s/value", tsid, tcid)

			kvTS := &consul.KVPair{Key: tcPath, Value: ts.ToConsul()}

			_, err := csl.Put(kvTS, nil)
			if err != nil {
				return err
			}

			for _, tstep := range tc.TestSteps {

				tstepid := strconv.FormatUint(tstep.ID, 10)
				tstepPath := fmt.Sprintf("vigie/%s/%s/%s/value", tsid, tcid, tstepid)

				kvTS := &consul.KVPair{Key: tstepPath, Value: ts.ToConsul()}

				_, err := csl.Put(kvTS, nil)
				if err != nil {
					return err
				}

			}

		}

	}

	return nil
}

// ImportAllTestCases will add new TCs, remove oldTC that are absent from the new TCs, keep common TCs
func (v *Vigie) ImportAllTestSuites(newTSs map[uint64]*teststruct.TestSuite) (TSs map[uint64]*teststruct.TestSuite, anyChanges bool) {

	// Compilation to avoid multiples loops
	newStateTS := make(map[string]uint64, 0)
	for _, nts := range newTSs {
		newStateTS[nts.Name] = nts.ID
	}

	// UPDATE
	for _, oTS := range v.TestSuites {
		// If old TC name is in newTCs
		if nTSid, alreadyExists := newStateTS[oTS.Name]; alreadyExists {
			// Name exists, but have they the same ID(hash) ?
			if oTS.ID != nTSid {
				// Name is identical to an existing TC, but ID is different.
				// That means that the old import have changed. And so TSteps.
				// The full old TestCase state is keep, but we need to go deeper to update this TestCase.
				anyChanges = true
				newTSs[nTSid].ImportAllTestCases(oTS.TestCases)
			} else {
				// Name and ID are identical = No Changes > replace newTS by "OldTS" to keep the state
				newTSs[nTSid] = oTS
			}
		}
	}

	// Handle the first time case
	if (len(v.TestSuites) == 0) && (len(newTSs) != 0) {
		anyChanges = true
	}

	return newTSs, anyChanges
}

// ---

// addTestSuite simply  add a TestSuite into Vigie, no checks concurency safe
func (v *Vigie) addTestSuite(newTS *teststruct.TestSuite) {

	v.mu.Lock()
	v.TestSuites[newTS.ID] = newTS
	v.mu.Unlock()

}

// removeTestSuite simply removes a TestSuite from Vigie, concurency safe
func (v *Vigie) removeTestSuite(ID uint64) {

	v.mu.Lock()
	delete(v.TestSuites, ID)
	v.mu.Unlock()

}
