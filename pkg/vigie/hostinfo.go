package vigie

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type HostInfo struct {
	Name        string            // Familiar Name
	URL         string            // Web URL to Vigie Instance
	Tags        map[string]string // Descriptives Tags of this host
	IPv6Capable bool              // IPv6Capable
}

// AddHostSytemInfo adds info about the specification
// or capabilities of a host.
func (hi *HostInfo) AddHostSytemInfo() {

	if hi.Name == "" {
		hostname, err := os.Hostname()
		if err == nil {
			hi.Name = "cannotbedetermined"
		}
		hi.Name = hostname
	}

	if hi.URL == "" {

		_, err := url.Parse(hi.URL)
		if err != nil {
			hi.URL = "http://badurlformat"
		}

		hostname, err := os.Hostname()
		if err == nil {
		}
		hi.URL = fmt.Sprintf("http://%s", hostname)
	}

	// Add cloud metadata
	hi.addCloudMetadata()

	// Can deal with IPv6 ?
	hi.isIPV6capable()

}

func (hi *HostInfo) addCloudMetadata() {}

func (hi *HostInfo) isIPV6capable() {

	// ipv6 Detection : Quick & Dirty
	interfaces, err := net.Interfaces()
	if err != nil {

	}
	for _, i := range interfaces {

		byNameInterface, err := net.InterfaceByName(i.Name)
		if err != nil {
			fmt.Println(err)
		}
		addresses, err := byNameInterface.Addrs()
		for _, v := range addresses {
			ip := v.String()
			switch {

			case strings.Contains(ip, "2000:"):
				hi.IPv6Capable = true

			case strings.Contains(ip, "fc00:"):
				hi.IPv6Capable = true

			case strings.Contains(ip, "fe80:"):
				hi.IPv6Capable = true

			case strings.Contains(ip, "2000:"):
				hi.IPv6Capable = true

			}
		}
	}
}
