package install

import (
	v1 "github.com/grafana/agent/pkg/integrations"
	v2 "github.com/grafana/agent/pkg/integrations/v2"
	metricsutils "github.com/grafana/agent/pkg/integrations/v2/metricsutils"
)

// Perform a migration of v1 integrations which do not yet have a v2
// counterpart. These integrations will be registered as Singletons
// to maintain existing behavior.
//
// To migrate a v1 integration to a v2 migration with support for multiple
// instances, v1 integrations must manually migrate themselves by calling
// v2.RegisterDynamic directly.
func init() {
	for _, v1Integration := range v1.RegisteredIntegrations() {
		// Look to see if there's a v2 integration with the same name.
		var found bool
		for _, v2Integration := range v2.Registered() {
			if v2Integration.Name() == v1Integration.Name() {
				found = true
				break
			}
		}
		if !found {
			v2.RegisterDynamic(v1Integration, v1Integration.Name(), v2.TypeSingleton, func(in interface{}) v2.WrappedConfig {
				return metricsutils.CreateShim(in.(v1.Config))
			})
		}
	}
}
