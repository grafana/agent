//go:build !linux
// +build !linux

package cadvisor //nolint:golint

import "github.com/grafana/agent/pkg/integrations"

func init() {
	integrations.RegisterIntegration(integrations.NewStubConfig(name, "the cadvisor integration only works on linux; enabling it on other platforms will do nothing"))
}
