package network

import (
	"errors"
	"net"
)

// IPVersion represents the IP protocol version
type IPVersion string

const (
	// IPv4 represents IPv4 protocol
	IPv4 IPVersion = "ipv4"
	// IPv6 represents IPv6 protocol
	IPv6 IPVersion = "ipv6"
	// IPAny represents either IPv4 or IPv6 protocol (prefers IPv4)
	IPAny IPVersion = "any"
)

// LocalIP get the host machine local IP address with specified version
func LocalIP(version IPVersion) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip nil IPs
			if ip == nil {
				continue
			}

			// Check if it's a private IP
			if !isPrivateIP(ip) {
				continue
			}

			// Apply IP version filter
			switch version {
			case IPv4:
				if ip4 := ip.To4(); ip4 != nil {
					return ip4, nil
				}
			case IPv6:
				if ip4 := ip.To4(); ip4 == nil && ip.To16() != nil {
					return ip, nil
				}
			case IPAny:
				// Prefer IPv4 over IPv6
				if ip4 := ip.To4(); ip4 != nil {
					return ip4, nil
				} else if ip.To16() != nil {
					return ip, nil
				}
			}
		}
	}

	switch version {
	case IPv4:
		return nil, errors.New("no private IPv4 address found")
	case IPv6:
		return nil, errors.New("no private IPv6 address found")
	default:
		return nil, errors.New("no private IP address found")
	}
}

func isPrivateIP(ip net.IP) bool {
	var privateIPBlocks []*net.IPNet
	for _, cidr := range []string{
		// don't check loopback ips
		//"127.0.0.0/8",    // IPv4 loopback
		//"::1/128",        // IPv6 loopback
		//"fe80::/10",      // IPv6 link-local
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}
