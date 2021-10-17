package process

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
)

// ProcessTask runs the teststep then write the result into a DB
func ProcessTask(task teststruct.Task) *teststruct.VigieResult {

	testResult := runTestStep(task.TestStep)

	return &testResult

}

// runTestStep runs Probe and Check Assertions (if no timeout or failure)
func runTestStep(tStep *teststruct.TestStep) teststruct.VigieResult {

	tStep.Mutex.RLock()
	tStepName := tStep.Name
	tStep.Mutex.RUnlock()

	utils.Log.WithFields(logrus.Fields{
		"package":  "process",
		"teststep": tStepName,
	}).Trace("Start test")

	var testRes teststruct.VigieResult
	testRes.LastAttempt = time.Now()

	// Run the Probe
	probeReturns, issue := runTestStepProbe(&tStep.ProbeWrap)
	if issue != nil {
		// timeout: No Probe results => No need to assert any subtests
		testRes.Status = teststruct.Timeout
		testRes.Issue = issue.Error()
		return testRes
	}

	// Loop and check on all IPs resolved in the []ProbeReturn
	// If one of the check fails =>
	// Set the worst case as TestStep Status (Err>timeout>AssertFail>Success)
	// For now only useful for debug and log
	// vigieResults contains (TestResult, AssertionResult, Final Status)
	//

	vigieResults := make([]teststruct.TestResult, 0, len(probeReturns))

	start := time.Now()

	utils.Log.WithFields(logrus.Fields{
		"package":  "process",
		"teststep": tStep.Name,
	}).Trace("Asserting test probe result")

	for _, pr := range probeReturns {
		vr := processProbeResult(tStep, pr)
		vigieResults = append(vigieResults, vr)
	}

	utils.Log.WithFields(logrus.Fields{
		"package":  "process",
		"teststep": tStep.Name,
		"desc":     "Teststep Assertion",
		"type":     "perfmon",
		"value":    time.Since(start),
	}).Tracef("Time to complete Teststep Assertion")

	testRes.TestResults = vigieResults
	testRes.Status = getFinalResultStatus(vigieResults)

	logStatus(tStepName, testRes)

	return testRes
}

// processProbeResult
// Error / timeout ...
// Assertion
func processProbeResult(tStep *teststruct.TestStep, pr probe.ProbeReturnInterface) (tr teststruct.TestResult) {

	// Add the TestResults
	tr = teststruct.TestResult{
		ProbeReturn:     pr,
		AssertionResult: nil,
		Status:          0,
	}

	// Error detection in order to avoid the assertion step if not needed
	// TODO: Décider de garder ce switch qui skip Assert car un Timeout pê un attendu 10/2020
	switch pr.GetProbeInfo().Status {

	case probe.Failure:
		// The probe has failed to create or send the request
		// No result to assert => Exit
		tr.Status = teststruct.Failure
		return tr

	case probe.Error:
		// The probe has encountered a error (can be considered as a desired state)
		prbCode := pr.GetProbeInfo().ProbeCode
		tr.Status = teststruct.Error

		// Despite the error if a probeCode is set (managed by the probe)
		// that error can be a desired state.
		// eg: Absence of a DNS domain / record (Monitor for Typosquatting)
		if prbCode == 0.0 {
			// Unhandled error: no result to assert properly => Exit
			return tr
		}

	case probe.Timeout:
		// The probe has encountered a timeout
		// no result to assert => Exit
		tr.Status = teststruct.Timeout
		return tr

	default:
		// Continue with Assertion
	}

	//
	// Assertion
	//
	// Even if Error : Continue and Assert
	// an Error can be the expected result
	// only if the error have been gracefully handle by the probe

	assertResult, assertSuccess := tStep.AssertProbeResult(pr)
	tr.AssertionResult = assertResult

	// TestResult after Assertion
	if assertSuccess == true {
		tr.Status = teststruct.Success
	} else {
		tr.Status = teststruct.AssertFailure
	}

	return tr
}

// getFinalResultStatus compare the TestResults of one test
// and pick the worst case as Final Status
func getFinalResultStatus(vrs []teststruct.TestResult) (finalStatus teststruct.StepStatus) {

	/*
		Success       StepStatus = 1
		NotDefined    StepStatus = 0
		AssertFailure StepStatus = -1
		Failure       StepStatus = -2
		Timeout       StepStatus = -3
		Error         StepStatus = -4
	*/

	finalStatus = teststruct.Success

	for _, vr := range vrs {

		if finalStatus == teststruct.Error {
			return teststruct.Error
		}

		if vr.Status < finalStatus {
			finalStatus = vr.Status
		}

	}
	return finalStatus
}

// logStatus simply log the testResult (Will be moved and improved)
func logStatus(tStepName string, pData teststruct.VigieResult) {

	switch pData.Status {
	case teststruct.Success:

		utils.Log.WithFields(logrus.Fields{
			"package":  "process",
			"teststep": tStepName,
		}).Debugf("TestStep OK - Assertion OK")

	case teststruct.Error:

		utils.Log.WithFields(logrus.Fields{
			"package":  "process",
			"teststep": tStepName,
		}).Debugf("TestStep KO - Probe Error %s", pData.Issue)

	case teststruct.Timeout:

		utils.Log.WithFields(logrus.Fields{
			"package":  "process",
			"teststep": tStepName,
		}).Debugf("TestStep KO - timeout %s", pData.Issue)

	case teststruct.AssertFailure:

		utils.Log.WithFields(logrus.Fields{
			"package":  "process",
			"teststep": tStepName,
		}).Debugf("TestStep KO - Assertion FAILED")

	default:

		utils.Log.WithFields(logrus.Fields{
			"package":  "process",
			"teststep": tStepName,
		}).Errorf("TestStep - %s", pData.Status)

	}

}

/*
func ProcessTask_OLD(task teststruct.Task) {

	execTimestamp := time.Now()
	procData := runTestStep(task.TestStep)

	// LogResult write probe result into TestStep
	// Then return if the TestStep ResultStatus has changed
	anyStateChange, alertEvent := task.TestStep.LogResult(procData)
	//task.WriteMetadataChanges(procData.LastAttempt)

	if anyStateChange == true {

		// Update TestSuites and TC state because a change occurred
		updateParentTestStruct(task, execTimestamp)

		task.RLockAll()
		if task.TestStep.Status == teststruct.Success {
			utils.Log.WithFields(logrus.Fields{
				"package": "process", "testcase": task.TestCase.Name, "teststep": task.TestStep.Name, "testsuite": task.TestSuite.Name,
			}).Infof("TestStep state has changed to %q.", task.TestStep.Status.String())
		} else {
			utils.Log.WithFields(logrus.Fields{
				"package": "process", "testcase": task.TestCase.Name, "teststep": task.TestStep.Name, "testsuite": task.TestSuite.Name,
			}).Warnf("TestStep state has changed to %q.", task.TestStep.Status.String())

		}
		task.RUnlockAll()

		if alertEvent && alertmanager.AM.IsEnabled() {
			_ = alertmanager.AM.AddToAlertList(task)
		}

	}

	if tsdb.TsdbMgr.Enabled {
		// Insert Task ResultStatus to DB
		tsdb.TsdbMgr.WriteOnTsdbs(task)
		// UpdateStatus TestSuite and TestCase
		tsdb.TsdbMgr.UpdateTestStateToDB(task)
	}
}
*/
