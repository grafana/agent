package shared

import (
	"net/url"

	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"

	common_config "github.com/prometheus/common/config"

	"github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/traces"
)

// Globals are used to pass around subsystem-wide settings that integrations
// can take advantage of.
type Globals struct {
	// AgentIdentifier provides an identifier for the running agent. This can
	// be used for labelling whenever appropriate.
	//
	// AgentIdentifier will be set to the hostname:port of the running agent.
	// TODO(rfratto): flag to override identifier at agent level?
	AgentIdentifier string

	// Some integrations may wish to interact with various subsystems for their
	// implementation if the desired behavior is not supported natively by the
	// integration manager.

	Metrics *metrics.Agent // Metrics subsystem
	Logs    *logs.Logs     // Logs subsystem
	Tracing *traces.Traces // Traces subsystem

	// BaseURL to use to invoke methods against the embedded HTTP server.
	AgentBaseURL *url.URL

	ClientConfig common_config.HTTPClientConfig

	Autoscrape autoscrape.Global
}

// CloneAgentBaseURL returns a copy of AgentBaseURL that can be modified.
func (g Globals) CloneAgentBaseURL() *url.URL {
	if g.AgentBaseURL == nil {
		return nil
	}
	rawURL := g.AgentBaseURL.String()
	u, err := url.Parse(rawURL)
	if err != nil {
		// The URL shouldn't be invalid at this point
		panic(err)
	}
	return u
}
