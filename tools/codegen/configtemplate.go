package main

import "github.com/grafana/agent/pkg/integrations/shared"

// ConfigurationTemplate is used for the code generator to generate the config
type ConfigurationTemplate struct {
	Config        interface{}
	DefaultConfig interface{}
	Type          shared.Type
	IsV1          bool
}
