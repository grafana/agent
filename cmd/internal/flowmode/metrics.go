package flowmode

import (
	"net/url"
	"strings"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/loki/write"
	"github.com/grafana/agent/component/prometheus/remotewrite"
	pyro "github.com/grafana/agent/component/pyroscope/write"
	"github.com/grafana/agent/pkg/flow"
	"golang.org/x/exp/maps"
)

func getReporterMetricsFunc(f *flow.Flow) func() map[string]interface{} {
	return func() map[string]interface{} {
		components := component.GetAllComponents(f, component.InfoOptions{})
		componentNames := map[string]struct{}{}
		for _, c := range components {
			componentNames[c.Registration.Name] = struct{}{}
		}
		metrics := map[string]interface{}{
			"enabled-components": getEnabledComponents(f),
		}
		if instances := getMetricsInstances(f); instances != nil {
			metrics["metrics-instances"] = instances
		}
		if instances := getLogsInstances(f); instances != nil {
			metrics["logs-instances"] = instances
		}
		if instances := getProfileInstances(f); instances != nil {
			metrics["profile-instances"] = instances
		}
		// todo: trace instances
		return metrics
	}
}

// getMetricsInstances returns all Grafana Cloud Hosted Metrics instances referenced by prometheus.remote_write components
func getMetricsInstances(f *flow.Flow) interface{} {
	components := component.GetAllComponents(f, component.InfoOptions{GetHealth: true, GetArguments: true})
	instances := map[string]bool{}
	for _, c := range components {
		if c.Registration.Name != "prometheus.remote_write" {
			continue
		}
		args, ok := c.Arguments.(remotewrite.Arguments)
		if !ok {
			continue
		}
		for _, e := range args.Endpoints {
			if id := getGrafanaNetInstance(e.URL, e.HTTPClientConfig); id != "" {
				instances[id] = true
			}
		}
	}
	if len(instances) == 0 {
		return nil
	}
	return maps.Keys(instances)
}

// getLogsInstances returns all Grafana Cloud Hosted Logs instances referenced by loki.write components
func getLogsInstances(f *flow.Flow) interface{} {
	components := component.GetAllComponents(f, component.InfoOptions{GetHealth: true, GetArguments: true})
	instances := map[string]bool{}
	for _, c := range components {
		if c.Registration.Name != "loki.write" {
			continue
		}
		args, ok := c.Arguments.(write.Arguments)
		if !ok {
			continue
		}
		for _, e := range args.Endpoints {
			if id := getGrafanaNetInstance(e.URL, e.HTTPClientConfig); id != "" {
				instances[id] = true
			}
		}
	}
	if len(instances) == 0 {
		return nil
	}
	return maps.Keys(instances)
}

// getProfileInstances returns all Grafana Cloud Profiles instances referenced by pyroscope.write components
func getProfileInstances(f *flow.Flow) interface{} {
	components := component.GetAllComponents(f, component.InfoOptions{GetHealth: true, GetArguments: true})
	instances := map[string]bool{}
	for _, c := range components {
		if c.Registration.Name != "pyroscope.write" {
			continue
		}
		args, ok := c.Arguments.(pyro.Arguments)
		if !ok {
			continue
		}
		for _, e := range args.Endpoints {
			if id := getGrafanaNetInstance(e.URL, e.HTTPClientConfig); id != "" {
				instances[id] = true
			}
		}
	}
	if len(instances) == 0 {
		return nil
	}
	return maps.Keys(instances)
}

func getGrafanaNetInstance(u string, h *config.HTTPClientConfig) string {
	// only look at endpoints that are very clearly pointed at grafana.net, with basic auth
	if h == nil || h.BasicAuth == nil {
		return ""
	}
	uParsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	if !strings.HasSuffix(uParsed.Hostname(), ".grafana.net") {
		return ""
	}
	return h.BasicAuth.Username
}

// getEnabledComponents returns the current enabled components
func getEnabledComponents(f *flow.Flow) interface{} {
	components := component.GetAllComponents(f, component.InfoOptions{})
	componentNames := map[string]struct{}{}
	for _, c := range components {
		componentNames[c.Registration.Name] = struct{}{}
	}
	return maps.Keys(componentNames)
}
