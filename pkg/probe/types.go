package probe

import (
	"encoding/json"
	"time"
)

type StepProbe map[string]interface{}
type ProbeAnswer map[string]interface{}

type Status int

const (
	Success Status = 1  // Success
	Timeout Status = -2 // The probe request encountered a timeout
	Error   Status = -3 // The probe request encountered a error (can be considered as a desired state)
	Failure Status = -4 // The probe cannot create the request nor send it
)

// ProbeInfo details
// DO NOT EDIT
type ProbeInfo struct {
	Error        string        `json:"error"`        // Error Details
	Status       Status        `json:"status"`       // Probe Status (OK, KO, TO..)
	ProbeCode    int           `json:"probecode"`    // ProbeCode for some specific error handling
	ResponseTime time.Duration `json:"responsetime"` // ResponseTime of the request
	IPresolved   string        `json:"ipresolved"`   // IPresolved (if multiples A / AAAA behind a FQDN)
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
	ProbeInfo ProbeInfo
	Answer    ProbeAnswer
}

// ProbeReturnInterface
type ProbeReturnInterface interface {
	StructAnswer() interface{}
	DumpAnswer() map[string]interface{}
	GetProbeInfo() ProbeInfo
	Labels() map[string]string
	Values() map[string]interface{}
}

// Probe execute a testStep.
type Probe interface {
	// Start run a Step TStep
	Run(timeout time.Duration) []ProbeReturnInterface
	GetName() string
	Initialize(StepProbe) error
	GenerateTStepName() string
	GetDefaultTimeout() time.Duration
	GetDefaultFrequency() time.Duration
	Labels() map[string]string
}

////////////////////////////////

/*
CYCLE IMPORT FOR NOW
func (x *ProbeComplete) ToVigieTest(probeType string) probe.VigieTest {

	var prbType proto.Message
	switch probeType {
	case "icmp":
		prbType = &icmp.Probe{}
	case "bar":
		prbType = &icmp.Probe{}
	}
	err := proto.Unmarshal(x.Spec.Value, prbType)
	if err != nil {

	}

	vt := probe.VigieTest{
		Metadata:   *x.Metadata,
		Spec:       x.Spec,
		Assertions: x.Assertions,
	}

	return vt
}
*/
