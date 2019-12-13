package teststruct

import (
	"github.com/vincoll/vigie/pkg/assertion"
	"time"
)

type VigieResult struct {
	ProbeResult       map[string]interface{}   `json:"probe_result"`
	AssertionResult   []assertion.AssertResult `json:"assertion_result"`
	Status            StepStatus               `json:"status"`
	StatusDescription string                   `json:"status_description"`
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

type AlertMessage2 struct {
	TSname     TSHeader          `json:"ts"`
	TC         []TCHeader        `json:"tc"`
	TStepRecap []TStepAlertShort `json:"TStepAlertShort"`
}

type TotalAlertMessage struct {
	Date       time.Time
	TestSuites map[int64]TSAlertShort
}

type StepStatus uint8

const (
	NotDefined StepStatus = iota
	Success
	Failure
	AssertFailure
	Timeout
	Error
)

func (ss StepStatus) String() string {
	return [...]string{"not_defined", "success", "failure", "assert_failure", "timeout", "error"}[ss]
}

func (ss StepStatus) IsSucess() bool {
	return ss == Success
}
