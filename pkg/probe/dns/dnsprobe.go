package dns

import (
	"context"
	"fmt"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"

	"github.com/vincoll/vigie/pkg/probe"
)

// Name of the probe
const Name = "dns"

// New returns a new Probe
func New() probe.Probe {
	return &Probe{}
}

// Return Probe Name
func (Probe) GetName() string {
	return Name
}

// Probe struct. Json and yaml descriptor are used for json output
type Probe struct {
	FQDN         string `json:"fqdn"`         // IP or Hostname
	RecordType   string `json:"recordtype" `  // Record Type to Lookup
	StrictAnswer bool   `json:"strictanswer"` // Attend uniquement la valeur
	NameServer   string `json:"nameserver" `  // Send Request to a specified Nameserver
}

// ProbeAnswer is the returned result after query
type ProbeAnswer struct {
	Answer       []string        `json:"answer"`
	TTL          int             `json:"ttl"`
	ResponseTime float64         `json:"responsetime"` // Reference Time Unit = Second
	Class        string          `json:"class"`
	Priority     int             `json:"priority"`
	Weight       int             `json:"weight"`
	Port         int             `json:"port"`
	Target       string          `json:"target"`
	Proto        string          `json:"proto"`
	ProbeInfo    probe.ProbeInfo `json:"probeinfo"`
}

// GenerateTStepName return a tstep name if non existent
func (p *Probe) GenerateTStepName() string {
	generatedName := fmt.Sprintf("%s_%s-%s", p.GetName(), p.FQDN, p.RecordType)
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
func (p *Probe) Run(timeout time.Duration) (probeReturn []probe.ProbeReturn) {

	chResult := make(chan ProbeAnswer, 1) // Chan for ResultStatus
	ctxExecTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var pr probe.ProbeReturn
	probeReturns := make([]probe.ProbeReturn, 0, 1)

	// Start the Request
	go p.work(chResult)

	// Select wait for a incoming result from any channel.
	select {

	case probeAnswer := <-chResult:
		//close(chFail)
		resDump, _ := probe.ToMap(probeAnswer)
		pr = probe.ProbeReturn{Status: probeAnswer.ProbeInfo.Status, Res: resDump, Err: probeAnswer.ProbeInfo.Error}

	// Timeout set by TestStep
	case <-ctxExecTimeout.Done():
		pr = probe.ProbeReturn{Status: probe.Timeout, Res: nil, Err: fmt.Sprintf("Timeout after %s", timeout)}
	}

	probeReturns = append(probeReturns, pr)
	return probeReturns
}

func (p *Probe) work(r chan<- ProbeAnswer) {

	switch p.RecordType {

	case "A":
		r <- lookupA(p.FQDN)
	case "AAAA":
		r <- lookupAAAA(p.FQDN)
	case "CNAME":
		r <- lookupCNAME(p.FQDN)
	case "TXT":
		r <- lookupTXT(p.FQDN)
	default:
		pi := probe.ProbeInfo{Error: fmt.Sprintf("%q is not a supported DNS Record Type.", p.RecordType), Status: probe.Error}
		r <- ProbeAnswer{ProbeInfo: pi}
	}

	return
}
