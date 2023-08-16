package build

import "github.com/grafana/agent/converter/diag"

func (s *ScrapeConfigBuilder) AppendConsulAgentSDs() {
	// TODO: implement this
	if s.cfg.ServiceDiscoveryConfig.ConsulAgentSDConfigs != nil {
		s.diags.Add(
			diag.SeverityLevelError,
			"consul_agent SDs are not currently supported in Grafana Agent Flow - see https://github.com/grafana/agent/issues/2261",
		)
	}
}
