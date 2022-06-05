package probetable

import (
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/icmp"
)

var AvailableProbes = map[string]probe.Probe{

	/*
		x509.Name:  x509.New(),
		hash.Name:  hash.New(),
		dns.Name:   dns.New(),
		port.Name:  port.New(),
		debug.Name: debug.New(),
		//http.Name: http.New(),

	*/

	// VIGIE PROTOBUF

	icmp.Name: icmp.New(),
}

/*
var AvailableProbes = map[string]probe.ProbeInfo{

	/*
		x509.Name:  x509.New(),
		hash.Name:  hash.New(),
		dns.Name:   dns.New(),
		port.Name:  port.New(),
		debug.Name: debug.New(),
		//http.Name: http.New(),



	// VIGIE PROTOBUF

	icmp.Name: icmp.New(),

}
*/
