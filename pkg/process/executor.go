package process

import (
	"fmt"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/teststruct"
)

// runTestStepProbe runs a probe test
// The probe can get an answer, or the probe can timeout.
//
// If multiples IPs are resolved : each IP is tested
// the probe will return multiples answser for each IP
func runTestStepProbe(pWrap *teststruct.ProbeWrap) ([]probe.ProbeReturnInterface, error) {

	// Create Channel for Probe ResultStatus
	chProbeReturn := make(chan []probe.ProbeReturnInterface, 1)
	//
	// Goroutine to run the probe teststep
	//
	go func() {
		chProbeReturn <- pWrap.Run()
	}()

	// Select dépend de l'issue de l'exec de la probe
	select {
	case probeRtrn := <-chProbeReturn: // Retour d'info de la probe
		return probeRtrn, nil

	// If no answer (timeout)
	// Failsafe : timeout
	case <-time.After(pWrap.Timeout):
		/*
			probeReturns := make([]probe.ProbeReturnInterface, 0)

			pi := probe.ProbeInfo{
				Error:  fmt.Sprintf("timeout after %s", pWrap.Frequency.String()),
				Status: probe.Timeout,
			}

			pr := probe.ProbeReturn{
				Answer:    nil,
				ProbeInfo: pi,
			}
			probeReturns = append(probeReturns, pr)
		*/
		return nil, fmt.Errorf("FailSafe: %s", pWrap.Timeout.String())
	}
}
