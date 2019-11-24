package port

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
)

func (p *Probe) process(timeout time.Duration) (probeAnswers []ProbeAnswer) {

	addrsPort, err := probe.ADVGetIPsfromHostname_port(p.Host, p.Port, p.IPprotocol)
	if err != nil {
		pi := probe.ProbeInfo{Status: probe.Error, Error: err.Error()}
		probeAnswers = make([]ProbeAnswer, 0)
		probeAnswers = append(probeAnswers, ProbeAnswer{Reachable: false, ProbeInfo: pi})
		return probeAnswers
	}

	if len(addrsPort) == 0 {
		errNoIP := fmt.Errorf("No IP for %s with ipv%d found.", p.Host, p.IPprotocol)

		pi := probe.ProbeInfo{Status: probe.Error, Error: errNoIP.Error()}
		probeAnswers = make([]ProbeAnswer, 0)
		probeAnswers = append(probeAnswers, ProbeAnswer{Reachable: false, ProbeInfo: pi})
		return probeAnswers
	}

	// Loop for each ip behind a DNS record
	// probePIs store the results for each IP
	probeAnswers = make([]ProbeAnswer, 0, len(addrsPort))
	var wg sync.WaitGroup
	wg.Add(len(addrsPort))

	// Check for each IP
	for _, hp := range addrsPort {

		go func() {
			pa, errReq := sendPortRequest(hp, p.Protocol, timeout)
			if errReq != nil {
				//print(errReq)
			}
			probeAnswers = append(probeAnswers, pa)
			wg.Done()
		}()

	}
	wg.Wait()

	return probeAnswers

}

func sendPortRequest(hostport, protocol string, timeout time.Duration) (ProbeAnswer, error) {

	start := time.Now()
	_, err := net.DialTimeout(protocol, hostport, timeout)
	elapsed := time.Since(start)

	// Error
	if err != nil {

		var pi probe.ProbeInfo
		probeErr := fmt.Sprintf("(%s@%s) %s", hostport, protocol, err)

		// Define Vigie ProbeCode Error
		switch er := err.Error(); {

		case strings.Contains(er, "no such host"):

			pi = probe.ProbeInfo{
				ResponseTime: elapsed,
				ProbeCode:    8749,
				Error:        probeErr,
				Status:       probe.Error,
			}

		case strings.Contains(er, "connect: connection refused"):

			pi = probe.ProbeInfo{
				ResponseTime: elapsed,

				ProbeCode: 6863,
				Error:     probeErr,
				Status:    probe.Error,
			}

		case strings.Contains(er, "i/o timeout"):
			// Iptable DROP is done silently => Timeout
			pi = probe.ProbeInfo{
				ResponseTime: elapsed,

				ProbeCode: 2074,
				Error:     probeErr,
				Status:    probe.Error,
			}

		case strings.Contains(er, "connect: network is unreachable"):

			pi = probe.ProbeInfo{
				ResponseTime: elapsed,

				ProbeCode: 666,
				Error:     probeErr,
				Status:    probe.Error,
			}

		default:
			pi = probe.ProbeInfo{
				ResponseTime: elapsed,

				ProbeCode: -1,
				Error:     err.Error(),
				Status:    probe.Error,
			}

		}
		// Fail
		return ProbeAnswer{
			Reachable: false,
			ProbeInfo: pi,
		}, nil

	}

	// OK
	pi := probe.ProbeInfo{
		ResponseTime: elapsed,
		Error:        "",
		Status:       probe.Success,
	}

	// Success
	return ProbeAnswer{
		Reachable: true,
		ProbeInfo: pi,
	}, nil

}
