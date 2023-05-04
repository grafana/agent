// Package advertise provide utilities to find addresses to advertise to
// cluster peers.
package advertise

import (
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
)

// DefaultInterfaces is a default list of common interfaces that are used for
// local network traffic for Unix-like platforms.
var DefaultInterfaces = []string{"eth0", "en0"}

// FirstAddress returns the first IPv4 address from the given interface names.
// Addresses used for APIPA will be ignored if possible.
func FirstAddress(interfaces []string) (net.IP, error) {
	var (
		errs      *multierror.Error
		privateIP net.IP
	)

	for _, ifaceName := range interfaces {
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			err = fmt.Errorf("interface %q: %w", ifaceName, err)
			errs = multierror.Append(errs, err)
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			err = fmt.Errorf("interface %q addrs: %w", ifaceName, err)
			errs = multierror.Append(errs, err)
			continue
		} else if len(addrs) <= 0 {
			err = fmt.Errorf("interface %q has no addresses", ifaceName)
			errs = multierror.Append(errs, err)
			continue
		}

		foundAddr := findSuitableIP(addrs)
		if foundAddr == nil {
			err = fmt.Errorf("interface %q has no suitable addresses", ifaceName)
			errs = multierror.Append(errs, err)
			continue
		}

		if !IsAutomaticPrivateIP(foundAddr) {
			return foundAddr, nil
		} else if privateIP == nil {
			privateIP = foundAddr
		}
	}

	if privateIP == nil {
		return nil, errs.ErrorOrNil()
	}
	return privateIP, nil
}

// findSuitableIP searches addrs for the first IPv4 address. IPv4 addresses
// used for APIPA will be ignored if possible.
//
// Returns nil if no suitable addresses were found.
func findSuitableIP(addrs []net.Addr) net.IP {
	var privateIP net.IP

	for _, addr := range addrs {
		addr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ipv4 := addr.IP.To4()
		if ipv4 == nil {
			// Not IPv4
			continue
		}

		// We can return non-automatic private IPs immediately, otherwise we'll
		// save the first one as a fallback if there are no better IPs.
		if !IsAutomaticPrivateIP(ipv4) {
			return ipv4
		} else if privateIP == nil {
			privateIP = ipv4
		}
	}

	return privateIP
}

// IsAutomaticPrivateIP checks whether IP represents an IP address for
// APIPA (Automatic Private IP Addressing) in the 169.254.0.0/16 range.
func IsAutomaticPrivateIP(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}

	var mask = net.IPv4Mask(255, 255, 0, 0)
	var subnet = net.IPv4(169, 254, 0, 0)
	return ip.Mask(mask).Equal(subnet)
}
