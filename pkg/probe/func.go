package probe

import (
	"fmt"
	"net"
	"strings"
)

func MixPIs(pis []*ProbeInfo) ProbeInfo {

	blendedPI := ProbeInfo{
		Status:       Success,
		Error:        "",
		ProbeCode:    pis[0].ProbeCode,
		ResponseTime: pis[0].ResponseTime,
	}

	var sb strings.Builder

	for _, pi := range pis {

		if pi.Status != Success {
			// If 1 Error => All Error
			if pi.Status == Error {
				blendedPI.Status = Error
			} else {
				blendedPI.Status = pi.Status
			}

		}

		if pi.ProbeCode != blendedPI.ProbeCode {
			blendedPI.ProbeCode = 666
		}
		if pi.ResponseTime > blendedPI.ResponseTime {
			blendedPI.ResponseTime = pi.ResponseTime
		}

		if pi.Error != "" {
			// Add each Error
			if sb.Len() == 0 {
				sb.WriteString(fmt.Sprintf("%s", pi.Error))
			} else {
				sb.WriteString(fmt.Sprintf(", %s", pi.Error))
			}
		}
	}

	blendedPI.Error = sb.String()

	return blendedPI

}

func GetIPsfromHostname(host string, port int) ([]string, error) {

	// Good example :
	// https: //golang.org/src/net/ipsock.go

	// If a DNS is provided:
	// The probe will check every IPs behind this DNS record
	// Only the worst ResponseTime will be keep
	// is Host an IP ?
	res := net.ParseIP(host)
	if res != nil {
		// Simply add IP
		simpleIP := []string{fmt.Sprintf("%s:%d", host, port)}
		return simpleIP, nil
	} else {
		// Get all the IPs behind a DNS record
		addrs, err := net.LookupHost(host)
		if err != nil {
			return nil, err
		}

		hostsPort := make([]string, 0, len(addrs))

		for _, addr := range addrs {
			if port == 0 {
				hostsPort = append(hostsPort, addr)
			} else {
				hostsPort = append(hostsPort, net.JoinHostPort(addr, fmt.Sprintf("%d", port)))

			}

		}
		return hostsPort, nil

	}

}

// ADVGetIPsfromHostname_port returns a array of IPs resolved by a Hostname
// Ipv4 or Ipv6
// This func must be divided WIP
func ADVGetIPsfromHostname_port(host string, port int, ipv int) ([]string, error) {

	// Good example :
	// https://golang.org/src/net/ipsock.go

	// If a DNS is provided:
	// The probe will check every IPs behind this DNS record
	// Only the worst ResponseTime will be keep
	// is Host an IP ?

	// Test if host is a network hotname or an IP
	// If IP : checks and return,
	// If NetworkName : Get all IP behind this name check and return.
	res := net.ParseIP(host)
	// If IP
	if res != nil {

		switch {

		/* All case (disable for now)
		case ipv == 0:

			if port == 0 {
				simpleIP := []string{res.String()}
				return simpleIP, nil
			} else {
				simpleIP := []string{net.JoinHostPort(res.String(), fmt.Sprintf("%d", port))}
				return simpleIP, nil
			}
		*/
		case isIPv4(res) && ipv <= 4:

			if port == 0 {
				simpleIP := []string{res.String()}
				return simpleIP, nil
			} else {
				simpleIP := []string{net.JoinHostPort(res.String(), fmt.Sprintf("%d", port))}
				return simpleIP, nil
			}

		case isIPv6(res) && ipv >= 6:

			if port == 0 {
				simpleIP := []string{res.String()}
				return simpleIP, nil
			} else {
				simpleIP := []string{net.JoinHostPort(res.String(), fmt.Sprintf("%d", port))}
				return simpleIP, nil
			}

		default:
			return nil, fmt.Errorf("the one IP %q is not the expected ip version (%d)", host, ipv)

		}
		// Simply add IP

	}

	//
	// Network Hostname
	// Get all the IPs behind a DNS record
	//

	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}

	hostsPort := make([]string, 0, len(addrs))

	for _, addr := range addrs {

		switch {

		case isIPv6(net.ParseIP(addr)) && ipv >= 6:

			if port == 0 {
				hostsPort = append(hostsPort, addr)
			} else {
				hostsPort = append(hostsPort, net.JoinHostPort(addr, fmt.Sprintf("%d", port)))

			}

		case isIPv4(net.ParseIP(addr)) && ipv <= 4:
			if port == 0 {
				hostsPort = append(hostsPort, addr)
			} else {
				hostsPort = append(hostsPort, net.JoinHostPort(addr, fmt.Sprintf("%d", port)))

			}

		default:
			/*
				if port == 0 {
					hostsPort = append(hostsPort, addr)
				} else {
					hostsPort = append(hostsPort, net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
				}
			*/

		}

	}
	return hostsPort, nil

}

// ADVGetIPsfromHostname returns a array of IPs resolved by a Hostname
// Ipv4 or Ipv6
// This func must be divided WIP
func ADVGetIPsfromHostname(host string, ipv int) ([]string, error) {

	// Good example :
	// https://golang.org/src/net/ipsock.go

	// If a DNS is provided:
	// The probe will check every IPs behind this DNS record
	// Only the worst ResponseTime will be keep
	// is Host an IP ?

	// Test if host is a network hotname or an IP
	// If IP : checks and return,
	// If NetworkName : Get all IP behind this name check and return.
	res := net.ParseIP(host)

	// If IP
	if res != nil {
		// Simply return the IP if matching ipv
		switch {

		case ipv == 0:

			simpleIP := []string{res.String()}
			return simpleIP, nil

		case isIPv4(res) && ipv <= 4:

			simpleIP := []string{res.String()}
			return simpleIP, nil

		case isIPv6(res) && ipv >= 6:

			simpleIP := []string{res.String()}
			return simpleIP, nil

		default:
			return nil, fmt.Errorf("the one IP %q is not the expected ip version (%d)", host, ipv)

		}

	}

	//
	// Network Hostname
	// Get all the IPs behind a DNS record
	//

	addrs, err := net.LookupHost(host)
	if err != nil {
		return nil, fmt.Errorf("error while DNS resolution of %q : %s", host, err)
	}

	hosts := make([]string, 0, len(addrs))

	for _, addr := range addrs {

		switch {

		case isIPv6(net.ParseIP(addr)) && ipv == 6:
			hosts = append(hosts, addr)

		case isIPv4(net.ParseIP(addr)) && ipv == 4:
			hosts = append(hosts, addr)

		default:
			hosts = append(hosts, addr)

		}

	}
	return hosts, nil

}

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}
