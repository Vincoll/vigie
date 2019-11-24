package port

import (
	"fmt"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"

	"github.com/vincoll/vigie/pkg/probe"
)

// Name of the probe
const Name = "port"

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
	Host       string `json:"host"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	IPprotocol int    `json:"ipprotocol"`
}

// ProbeAnswer is the returned result after query
// All attributes must be Public
// ProbeInfo is Mandatory => Détail l'execution de la probe
type ProbeAnswer struct {
	Reachable bool            `json:"reachable"`
	ProbeInfo probe.ProbeInfo `json:"probeinfo"`
}

// GenerateTStepName return a tstep name if non existent
func (p *Probe) GenerateTStepName() string {
	generatedName := fmt.Sprintf("%s_%s@%s:%d", p.GetName(), p.Protocol, p.Host, p.Port)
	return generatedName
}

// Initialize Probe struct data
func (p *Probe) Initialize(step probe.StepProbe) error {

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
		return fmt.Errorf("a step is not valid: %s", step)
	}

	return nil
}

// Start the probe request
func (p *Probe) Run(timeout time.Duration) (probeReturns []probe.ProbeReturn) {

	// Start the Request
	probeAnswers := p.process(timeout)
	probeReturns = make([]probe.ProbeReturn, 0, len(probeAnswers))

	for _, pa := range probeAnswers {

		resDump, err := probe.ToMap(pa)
		if err != nil {
		}
		pr := probe.ProbeReturn{Status: pa.ProbeInfo.Status, Res: resDump, Err: pa.ProbeInfo.Error}
		probeReturns = append(probeReturns, pr)

	}

	return probeReturns

}
