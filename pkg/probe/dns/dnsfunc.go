package dns

import (
	"net"
	"strings"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
)

// LookupNS returns the DNStask NS records for the given domain name.
func LookupNS(probe Probe) (*ProbeAnswer, error) {
	// TODO NS
	return nil, nil
}

func lookupA(fqdn string) ProbeAnswer {

	ipsv4 := make([]string, 0)
	start := time.Now()
	returnedAnswer, err := net.LookupHost(fqdn)
	elapsed := time.Since(start)

	// Error
	if err != nil {

		pa := generateProbeErrCode(err)
		pa.ResponseTime = elapsed.Seconds()

		return pa
	}

	// Success
	pi := probe.ProbeInfo{
		Status: 1,
	}

	for _, ip := range returnedAnswer {
		if isIPv4(ip) {
			ipsv4 = append(ipsv4, ip)
		}
	}

	pa := ProbeAnswer{
		Answer:       ipsv4,
		ResponseTime: elapsed.Seconds(),
		ProbeInfo:    pi,
	}

	return pa
}

func lookupAAAA(fqdn string) ProbeAnswer {

	ipsv6 := make([]string, 0)
	start := time.Now()
	returnedAnswer, err := net.LookupHost(fqdn)
	elapsed := time.Since(start)

	// Error
	if err != nil {

		pa := generateProbeErrCode(err)
		pa.ResponseTime = elapsed.Seconds()

		return pa
	}

	// Success
	pi := probe.ProbeInfo{
		Status: 1,
	}

	for _, ip := range returnedAnswer {
		if isIPv6(ip) {
			ipsv6 = append(ipsv6, ip)
		}
	}

	pa := ProbeAnswer{
		Answer:       ipsv6,
		ResponseTime: elapsed.Seconds(),
		ProbeInfo:    pi,
	}

	return pa
}

func lookupTXT(fqdn string) ProbeAnswer {

	start := time.Now()
	returnedAnswer, err := net.LookupTXT(fqdn)
	elapsed := time.Since(start)
	// Error
	if err != nil {

		pa := generateProbeErrCode(err)
		pa.ResponseTime = elapsed.Seconds()

		return pa
	}

	pa := ProbeAnswer{
		Answer:       returnedAnswer,
		ResponseTime: elapsed.Seconds(),
		ProbeInfo:    probe.ProbeInfo{Status: 1},
	}

	return pa
}

func lookupCNAME(fqdn string) ProbeAnswer {

	cname := make([]string, 0)

	start := time.Now()
	returnedAnswer, err := net.LookupCNAME(fqdn)
	elapsed := time.Since(start)
	// Error
	if err != nil {

		pa := generateProbeErrCode(err)
		pa.ResponseTime = elapsed.Seconds()

		return pa
	}
	// Convert retAnswer (String) to []String
	cname = append(cname, returnedAnswer)

	pa := ProbeAnswer{
		Answer:       cname,
		ResponseTime: elapsed.Seconds(),
		ProbeInfo:    probe.ProbeInfo{Status: 1},
	}

	return pa
}

func generateProbeErrCode(err error) ProbeAnswer {

	pi := probe.ProbeInfo{
		Status: -3,
		Error:  err.Error(),
	}
	// Define Vigie ProbeCode Error
	switch {

	case strings.ContainsAny("no such host", err.Error()):
		pi.ProbeCode = 404
	default:
		pi.ProbeCode = 666

	}

	pa := ProbeAnswer{
		ProbeInfo: pi,
	}
	return pa
}

/*

DNS Lib

func lookupMX(fqdn string) ProbeAnswer {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.FQDN(fqdn), dns.TypeMX)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil {
		fmt.Printf("*** error: %s\n", err.Error())
	}

	if r.Rcode != dns.RcodeSuccess {
		fmt.Printf(" *** invalid answer name %s after MX query for %s\n", os.Args[1], os.Args[1])
	}
	// Stuff must be in the answer section
	for _, a := range r.Answer {
		fmt.Printf("%v\n", a)
	}
	return nil, nil
}
*/
func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

/*
func testanswerarraystring(record []string, returnedAnswer []string, strictAswr bool) bool {

	if strictAswr == true {
		if sameStringSlice(record, returnedAnswer) {
			return true
		} else {
			return false
		}
	} else {
		if containsIn(record, returnedAnswer) {
			return true
		} else {
			return false
		}
	}
}
*/
// https://golang.org/src/net/lookup.go?s=2777:4339#L92
