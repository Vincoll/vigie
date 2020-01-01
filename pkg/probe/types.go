package probe

import (
	"encoding/json"
	"time"
)

type StepProbe map[string]interface{}
type ProbeResult map[string]interface{}

type Status int

const (
	Success Status = 1
	Timeout Status = -2
	Error   Status = -3
)

// ProbeInfo details
// DO NOT EDIT
type ProbeInfo struct {
	Error        string        `json:"error"`
	Status       Status        `json:"status"`
	ProbeCode    int           `json:"probecode"`
	ResponseTime time.Duration `json:"responsetime"`
	SubTest      string        `json:"subtest"`
}

func (pi ProbeInfo) MarshalJSON() ([]byte, error) {

	type Copy ProbeInfo
	return json.Marshal(&struct {
		ResponseTime string `json:"responsetime"`
		Copy
	}{
		ResponseTime: pi.ResponseTime.String(),
		Copy:         (Copy)(pi),
	})
}

type ProbeDuration struct {
	time.Duration
}

func (d ProbeDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// ProbeReturn represents an probe result on a test step
type ProbeReturn struct {
	Res    ProbeResult
	Err    string
	Status Status
}

// Probe execute a testStep.
type Probe interface {
	// Start run a Step TStep
	Run(timeout time.Duration) []ProbeReturn
	GetName() string
	Initialize(StepProbe) error
	GenerateTStepName() string
}
