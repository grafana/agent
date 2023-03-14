package docker

import (
	"fmt"
	"net"

	"github.com/grafana/agent/pkg/autodiscovery"
)

// Run ...
func Run() (*autodiscovery.Result, error) {
	// Check unix:///var/run/docker.sock for the docker daemon.
	_, err := net.Dial("unix", "/var/run/docker.sock")
	if err != nil {
		return nil, fmt.Errorf("could not reach docker daemon: %w", err)
	}
	return nil, nil
}
