package process

import (
	"github.com/vincoll/vigie/pkg/alertmanager"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/tsdb"
	"github.com/vincoll/vigie/pkg/utils"
)

func ProcessTask(task teststruct.Task) {

	procData := runTestStep(task.TestStep)

	// WriteResult write probe result into TestStep
	// Then return if the TestStep ResultStatus has changed
	anyStateChange := task.TestStep.WriteResult(procData)

	if anyStateChange == true {

		// Update TestSuites and TC state because a change occurred
		updateParentTestStruct(task.TestSuite, task.TestCase)

		task.WriteMetadataChanges(procData.LastAttempt)

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

		if alertmanager.AM.IsEnable() {
			_ = alertmanager.AM.AddToAlertList(task)
		}

	}

	if tsdb.InfluxInst.IsEnable() {
		// Insert Task ResultStatus to DB
		insertTaskToDB(&task)
		// UpdateStatus TestSuite and TestCase
		updateTestStateToDB(&task)
	}
}

// runTestStep runs Probe and Check Assertions (if no timeout)
func runTestStep(tStep *teststruct.TestStep) *teststruct.Processing {

	tStep.Mutex.RLock()
	utils.Log.WithFields(logrus.Fields{
		"package":  "process",
		"teststep": tStep.Name,
	}).Trace("Start test")
	tStep.Mutex.RUnlock()

	// pData to avoid multiple lock / unlock, processing data will be written
	// in tStep at the end.
	var pData teststruct.Processing
	pData.LastAttempt = time.Now()

	// Run the Probe
	probeReturns, issue := runTestStepProbe(&tStep.ProbeWrap)
	if issue != nil {
		// Timeout: No Probe result => No need to assert
		pData.Status = teststruct.Timeout
		pData.Issue = issue.Error()
		return &pData
	}

	// Loop and check on all the ProbeResults
	// If one of the check fails =>
	// Set the worst case as TestStep Status (Err>Timeout>AssertFail>Success)

	// vigieResults contains (VigieResults, AssertionResult, Final Status)
	vigieResults := make([]teststruct.VigieResult, 0, len(probeReturns))

	for _, pr := range probeReturns {
		vr := processProbeResult(tStep, pr)
		vigieResults = append(vigieResults, vr)
	}

	pData.VigieResults = vigieResults
	pData.Status = getFinalResultStatus(vigieResults)

	return &pData
}

func processProbeResult(tStep *teststruct.TestStep, pr probe.ProbeReturn) (vr teststruct.VigieResult) {

	// Add the VigieResults
	vr.ProbeResult = pr.Res

	// Look for any error, to avoid Assertion if not needed
	switch pr.Status {

	case probe.Error:

		prbCode := pr.Res["probeinfo"].(map[string]interface{})["probecode"]

		vr.Status = teststruct.Error
		vr.StatusDescription = pr.Err

		if prbCode == 0.0 {
			// Unhandled error: no result to assert properly => Exit
			return vr
		}

	case probe.Timeout:
		// Timeout: no result to assert properly => Exit
		vr.Status = teststruct.Timeout
		vr.StatusDescription = pr.Err
		return vr
	}

	//
	// Assertion
	//
	// Even if Error : Continue and Assert
	// an Error can be the expected result
	// but only Error gracefully handle by the probe

	assertResult, success := tStep.AssertProbeResult(&pr.Res)
	vr.AssertionResult = assertResult

	// VigieResult after Assertion
	if success == true {
		vr.Status = teststruct.Success
	} else {
		vr.Status = teststruct.AssertFailure
	}

	return vr
}

// getFinalResultStatus stack compare the VigieResults,
// of one test and pick the worst case as Final Status
func getFinalResultStatus(vrs []teststruct.VigieResult) (finalStatus teststruct.StepStatus) {

	// Error>Timeout>AssertFailure>Success

	for _, vr := range vrs {
		if vr.Status > finalStatus {
			finalStatus = vr.Status
		}
	}
	return finalStatus
}
