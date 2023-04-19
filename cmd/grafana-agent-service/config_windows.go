package main

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// config holds configuration options to run the service.
type config struct {
	// ServicePath points to the path of the managed Grafana Agent binary.
	ServicePath string

	// Args holds arguments to pass to the Grafana Agent binary. os.Args[0] is
	// not included.
	Args []string

	// WorkingDirectory points to the working directory to run the Grafana Agent
	// binary from.
	WorkingDirectory string
}

// loadConfig loads the config from the Windows registry.
func loadConfig() (*config, error) {
	// NOTE(rfratto): the key name below shouldn't be changed without being
	// able to either migrate from the old key to the new key or supporting
	// both the old and the new key at the same time.

	agentKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `Software\Grafana\Grafana Agent Flow`, registry.READ)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}

	servicePath, _, err := agentKey.GetStringValue("")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key (Default): %w", err)
	}

	args, _, err := agentKey.GetStringsValue("Arguments")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key Arguments: %w", err)
	}

	return &config{
		ServicePath:      servicePath,
		Args:             args,
		WorkingDirectory: filepath.Dir(servicePath),
	}, nil
}
