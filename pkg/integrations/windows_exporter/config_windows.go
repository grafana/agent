package windows_exporter //nolint:golint

import (
	"reflect"

	"github.com/prometheus-community/windows_exporter/collector"
)

func (c *Config) applyConfig(exporterConfigs map[string]collector.Config) {
	agentConfigs := []translatableConfig{
		&c.Exchange,
		&c.IIS,
		&c.LogicalDisk,
		&c.MSMQ,
		&c.MSSQL,
		&c.Network,
		&c.Process,
		&c.Service,
		&c.SMTP,
		&c.TextFile,
	}
	// Brute force the syncing, its a bounded set and reduces the code footprint
	for _, ac := range agentConfigs {
		if ac == nil || reflect.ValueOf(ac).IsNil() {
			continue
		}
		for _, ec := range exporterConfigs {
			// Sync will return true if it can handle the exporter config
			// which means we can break early
			if ac.sync(ec) {
				break
			}
		}
	}
}
