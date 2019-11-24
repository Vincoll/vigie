package process

import (
	"fmt"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/teststruct"
)

func stringifyExecutorResult(e probe.ProbeResult) map[string]string {
	out := make(map[string]string)
	for k, v := range e {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

// Lance la step avec la bonne probe
// Un compte à rebourd (time.after) est crée à partir de la fréquence :
// Il stoppera une requéte trop longue non géré par la probe qui dispose lui d'un context plus fin: TimeOut
func runTestStepProbe(pWrap *teststruct.ProbeWrap) ([]probe.ProbeReturn, error) {

	// Create Channel for Probe ResultStatus
	chProbeReturn := make(chan []probe.ProbeReturn, 1)
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
	/*

		switch pStatus := probeRtrn.Status; {

		case pStatus == probe.Success:
			// Probe Success
			return &probeRtrn, nil

		case pStatus == probe.Error:
			// Probe Error
			return &probeRtrn, fmt.Errorf("Probe error: %s", probeRtrn.Err)

		case pStatus == probe.Timeout:
			// Timeout return by the probe
			return &probeRtrn, fmt.Errorf("Timeout after: %s", pWrap.Timeout.String())

		default:
			utils.Log.WithFields(logrus.Fields{
				"package":  "process",
				"teststep": pWrap.Probe.GenerateTStepName(),
			}).Errorf("Probe return status unknown")
			return &probeRtrn, fmt.Errorf("internal error: unknown status")

		} // switch



	*/
	// If no answer (Timeout)
	// Failsafe : Timeout before the other iteration.
	// In case of the Probe timeout have failed
	case <-time.After(pWrap.Frequency):

		probeReturns := make([]probe.ProbeReturn, 0)

		pr := probe.ProbeReturn{
			Res:    nil,
			Err:    fmt.Sprintf("timeout after %s", pWrap.Frequency.String()),
			Status: probe.Timeout,
		}
		probeReturns = append(probeReturns, pr)

		return probeReturns, fmt.Errorf("FailSafe: %s", pWrap.Timeout.String())
	} // select
}
