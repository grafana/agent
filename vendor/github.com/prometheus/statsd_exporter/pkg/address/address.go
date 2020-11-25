// Copyright 2013 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package address

import (
	"fmt"
	"net"
	"strconv"
)

func IPPortFromString(addr string) (*net.IPAddr, int, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, 0, fmt.Errorf("bad StatsD listening address: %s", addr)
	}

	if host == "" {
		host = "0.0.0.0"
	}
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to resolve %s: %s", host, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return nil, 0, fmt.Errorf("bad port %s: %s", portStr, err)
	}

	return ip, port, nil
}

func UDPAddrFromString(addr string) (*net.UDPAddr, error) {
	ip, port, err := IPPortFromString(addr)
	if err != nil {
		return nil, err
	}
	return &net.UDPAddr{
		IP:   ip.IP,
		Port: port,
		Zone: ip.Zone,
	}, nil
}

func TCPAddrFromString(addr string) (*net.TCPAddr, error) {
	ip, port, err := IPPortFromString(addr)
	if err != nil {
		return nil, err
	}
	return &net.TCPAddr{
		IP:   ip.IP,
		Port: port,
		Zone: ip.Zone,
	}, nil
}
