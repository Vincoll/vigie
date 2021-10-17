package probetable

import (
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/http"
)

var AvailableProbes = map[string]probe.Probe{

	/*
		x509.Name:  x509.New(),
		hash.Name:  hash.New(),
		dns.Name:   dns.New(),
		port.Name:  port.New(),
		debug.Name: debug.New(),
		icmp.Name:  icmp.New(),

	*/
	// NEW VIGIE TIME SERIES SYSTEM
	http.Name: http.New(),
}
