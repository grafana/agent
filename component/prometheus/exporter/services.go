package exporter

import (
	"github.com/grafana/agent/service/http"
	"golang.org/x/exp/maps"
)

// RequiredServices returns the set of services needed by all
// prometheus.exporter components. Callers may optionally pass in additional
// services to add to the returned list.
func RequiredServices(additionalServices ...string) []string {
	services := map[string]struct{}{
		http.ServiceName: {},
	}
	for _, svc := range additionalServices {
		services[svc] = struct{}{}
	}

	return maps.Keys(services)
}
