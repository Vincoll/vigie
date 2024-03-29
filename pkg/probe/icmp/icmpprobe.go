package icmp

import (
	"fmt"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
	// 	ping "github.com/digineo/go-ping"
	valid "github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"
)

// Name of executor
const Name = "icmp"
const TSDBmetric = "icmp"

// New returns a new Probe
func New() probe.Probe {
	return &Probe{}
}

// Return Probe Name
func (Probe) GetName() string {
	return Name
}

// Return Metric Key associated with this probe
func (Probe) GetTSDBmetric() string {
	return TSDBmetric
}

func (Probe) GetDefaultTimeout() time.Duration {
	return time.Second * 10
}

func (Probe) GetDefaultFrequency() time.Duration {
	return time.Second * 10
}

// Probe struct. Json and yaml descriptor are used for json output
type Probe struct {
	Name        string        `json:"name"`
	Host        string        `json:"host"`
	IPversion   int           `json:"ipversion"` // valid:"equal(0|4|6)"
	PayloadSize int           `json:"payloadsize" valid:"nonnegative"`
	Count       int           `json:"count" valid:"nonnegative"`
	Interval    time.Duration `json:"interval" valid:"nonnegative"`
}

func (p *Probe) Labels() map[string]string {

	lbl := map[string]string{

		"host":        p.Host,
		"ipversion":   fmt.Sprint(p.IPversion),
		"payloadsize": fmt.Sprint(p.PayloadSize),
	}
	return lbl

}

// ResultStatus represents a step result. Json and yaml descriptor are used for json output
type ProbeICMPReturnInterface struct {
	ProbeInfo  probe.ProbeInfo `json:"probeinfo"`
	Reacheable string          `json:"reacheable"`
	Rtt        time.Duration   `json:"rtt"`
}

func (pa *ProbeICMPReturnInterface) StructAnswer() interface{} {
	return pa
}
func (pa *ProbeICMPReturnInterface) GetProbeInfo() probe.ProbeInfo {
	return pa.ProbeInfo
}
func (pa *ProbeICMPReturnInterface) DumpAnswer() map[string]interface{} {

	aswDump, err := probe.ToMap(pa)
	if err != nil {
	}
	return aswDump

}
func (pa *ProbeICMPReturnInterface) Labels() map[string]string {

	labels := map[string]string{

		"ip": pa.ProbeInfo.IPresolved,
	}

	return labels
}
func (pa *ProbeICMPReturnInterface) Values() map[string]interface{} {

	values := map[string]interface{}{

		"reacheable": pa.Reacheable,
		"rtt":        pa.Rtt,
		"status":     pa.ProbeInfo.Status,
	}

	return values

}

// GetDefaultAssertions return default assertions for this executor
// Optional
/*
func (Probe) GetDefaultAssertions() teststruct.StepAssertions {
	return teststruct.StepAssertions{Assertions: []string{"result.code == Success"}}
}
*/
// GenerateTStepName return a tstep name if non existent
func (p *Probe) GenerateTStepName() string {
	generatedName := fmt.Sprintf("%s_ipv%d_%s_pl%d", p.GetName(), p.IPversion, p.Host, p.PayloadSize)
	return generatedName
}

func (p *Probe) applyDefaultValues() {

	if p.IPversion == 0 {
		p.IPversion = 4
	}

	if p.PayloadSize == 0 {
		p.PayloadSize = 64
	}

	// Default Count if empty
	if p.Count == 0 {
		p.Count = 2
	}

	if p.Interval == 0 {
		p.Interval = time.Millisecond * 10
	}

}

// Initialize Probe struct data
func (p *Probe) Initialize(step probe.StepProbe) error {

	// Decode Probe Sruct from TestStep
	//var e Probe
	if err := mapstructure.Decode(step, &p); err != nil {
		return err
	}
	// Check if Users's TestStep is Valid
	_, err := valid.ValidateStruct(p)
	if err != nil {
		return err
	}

	p.applyDefaultValues()

	// Advanced Validation
	okip := map[int]bool{10: true, 4: true, 6: true}
	if !okip[p.IPversion] {
		// if ipVersion is not set so 0 , both ipv4 and ipv6 addresses will be resolved
		return fmt.Errorf("Ip version %d is unknown", p.IPversion)
	}

	// Set Payload if empty

	// Check Payload limits
	if p.IPversion == 4 || p.IPversion == 0 {
		if p.PayloadSize > 8960 {
			return fmt.Errorf("The Maximal payload for ipv4 is 8960 bytes. Your value %d", p.PayloadSize)
		}
	}
	if p.IPversion == 6 {
		if p.PayloadSize > 65535 {
			return fmt.Errorf("The maximal payload for ipv6 is 65535 bytes. Your value %d", p.PayloadSize)
		}
	}

	if p.Host == "" {
		return fmt.Errorf("Probe host value is not defined")
	}

	// Test if ICMP Capable
	_, errICMP := p.sendICMP("127.0.0.1", time.Second)
	if errICMP != nil {
		return fmt.Errorf("No icmp packet can't be sent (tested on localhost). Linux required some system tweak to send icmp (cf: https://github.com/sparrc/go-ping#note-on-linux-support).")
	}

	// Return Valid and Loaded Probe
	return nil
}

func (p *Probe) Run(timeout time.Duration) (probeReturns []probe.ProbeReturnInterface) {

	// Start the Request
	probeAnswers := p.process(timeout)
	return probeAnswers

}
