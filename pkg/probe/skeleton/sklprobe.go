package skeleton

import (
	"context"
	"fmt"
	"strings"

	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/teststruct"
)

// Name of the probe
const Name = "base_skeleton"

// New returns a new Probe
func New() probe.Probe {
	return &Probe{}
}

// Return Probe Name
func (Probe) GetName() string {
	return Name
}

// Probe struct : Informations necessaires à l'execution de la probe
// All attributes must be Public
type Probe struct {
	Foo  string `json:"foo"`
	Type int    `json:"foo"`
}

// ProbeAnswer is the returned result after query
// All attributes must be Public
// ProbeInfo is Mandatory => Détail l'execution de la probe
type ProbeAnswer struct {
	Answer       int             `json:"anwser"`
	ProbeInfo    probe.ProbeInfo `json:"probeinfo"`
	ResponseTime float64
}

// GenerateTStepName return a tstep name if non existent
func (p *Probe) GenerateTStepName() string {
	generatedName := fmt.Sprintf("%s_%s", p.GetName(), p.Foo)
	return generatedName
}

// Initialize Probe struct data
func (p *Probe) Initialize(step probe.StepProbe) error {

	// Decode Probe Struct from TestStep
	if err := mapstructure.Decode(step, p); err != nil {
		return err
	}

	// Lower the case for vigie valid SHA1 => sha1
	p.Foo = strings.ToLower(p.Foo)

	// Check if TestStep is Valid with asaskevich/govalidator
	ok, err := valid.ValidateStruct(p)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("a step is not valid: ", step)
	}

	return nil
}

// Initialize Probe struct data
func (p *Probe) Validate(step teststruct.Step) error {

	// Decode Probe Struct from TestStep
	if err := mapstructure.Decode(step, p); err != nil {
		return err
	}

	// Check if TestStep is Valid with asaskevich/govalidator
	ok, err := valid.ValidateStruct(p)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("a step is not valid: ", step)
	}

	return nil
}

// Start the probe request
func (p *Probe) Run(timeout time.Duration) (probeReturn []probe.ProbeReturn) {

	chResult := make(chan ProbeAnswer, 1) // Chan for ResultStatus
	//chFail := make(chan ProbeFail, 1)     // Chan for Err
	ctxExecTimeout, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()
	// Start the Request
	go p.work(chResult)

	// Select wait for a incoming result from any channel.
	select {

	case workRes := <-chResult:
		//close(chFail)
		resDump, _ := probe.ToMap(workRes)
		x := probe.ProbeReturn{Status: workRes.ProbeInfo.Status, Res: resDump, Err: workRes.ProbeInfo.Error}
		return x

	// Timeout set by TestStep
	case <-ctxExecTimeout.Done():
		return probe.ProbeReturn{Status: probe.Timeout, Res: nil, Err: fmt.Sprint("Timeout after %d ms", timeout)}
	}

}

// work déclenche l'appel "metier" de la probe.
// Le switch sert à appeller une fonction particuliére en fonction des info de la probe.
func (p *Probe) work(r chan<- ProbeAnswer) {

	switch p.Type {

	case 1:
		r <- func1(p.Foo)
	default:
		pi := probe.ProbeInfo{Error: fmt.Sprintf("Unknown Type"), Status: probe.Error}
		r <- ProbeAnswer{ProbeInfo: pi}
	}

	return
}
