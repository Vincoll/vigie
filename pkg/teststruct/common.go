package teststruct

import (
	"github.com/vincoll/vigie/pkg/assertion"
	"github.com/vincoll/vigie/pkg/probe"
	"time"
)

// TestResult combines the ProbeReturn and the AssertionResult for a resolved IP
type TestResult struct {
	ProbeReturn     probe.ProbeReturnInterface `json:"probe_return"`
	AssertionResult []assertion.AssertResult   `json:"assertion_result"`
	Status          StepStatus                 `json:"status"`
}

type UIDTest struct {
	TestSuite TSHeader      `json:"testsuite"`
	TestCase  TCHeader      `json:"testcase"`
	TestStep  TStepDescribe `json:"teststep"`
}

type AlertMessage struct {
	TSname     string     `json:"tsname"`
	TSfile     string     `json:"tsfile"`
	TCname     string     `json:"tcname"`
	TCstatus   bool       `json:"tcres"`
	TStepfails []TestStep `json:"tstepfails"`
}

type TotalAlertMessage struct {
	Date       time.Time
	TestSuites map[uint64]TSAlertShort
}

type StepStatus int

const (
	Success       StepStatus = 1
	NotDefined    StepStatus = 0
	AssertFailure StepStatus = -1
	Failure       StepStatus = -2
	Timeout       StepStatus = -3
	Error         StepStatus = -4
)

func (ss StepStatus) String() string {

	m := map[StepStatus]string{
		NotDefined:    "not_defined",
		Success:       "success",
		AssertFailure: "assert_failure",
		Failure:       "failure",
		Timeout:       "timeout",
		Error:         "error",
	}

	return m[ss]
}

func (ss StepStatus) Int() int {
	return int(ss)
}

func (ss StepStatus) IsSucess() bool {
	return ss == Success
}

func (ss StepStatus) IsTimeMesureable() bool {
	if ss == Success || ss == AssertFailure {
		return true
	} else {
		return false
	}
}
