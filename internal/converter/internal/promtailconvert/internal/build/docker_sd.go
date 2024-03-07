package build

import (
	"time"

	"github.com/grafana/agent/internal/component/common/loki"
	flow_relabel "github.com/grafana/agent/internal/component/common/relabel"
	"github.com/grafana/agent/internal/component/discovery"
	"github.com/grafana/agent/internal/component/discovery/docker"
	loki_docker "github.com/grafana/agent/internal/component/loki/source/docker"
	"github.com/grafana/agent/internal/converter/internal/common"
	"github.com/prometheus/prometheus/discovery/moby"
)

func (s *ScrapeConfigBuilder) AppendDockerPipeline() {
	if len(s.cfg.DockerSDConfigs) == 0 {
		return
	}

	for i, sd := range s.cfg.DockerSDConfigs {
		compLabel := common.LabelWithIndex(i, s.globalCtx.LabelPrefix, s.cfg.JobName)

		// Add discovery.docker
		s.f.Body().AppendBlock(common.NewBlockWithOverride(
			[]string{"discovery", "docker"},
			compLabel,
			toDiscoveryDocker(sd),
		))

		// The targets output from above component
		targets := "discovery.docker." + compLabel + ".targets"

		// Add loki.source.docker
		overrideHook := func(val interface{}) interface{} {
			switch val.(type) {
			case []discovery.Target: // override targets expression to our string
				return common.CustomTokenizer{Expr: targets}
			case flow_relabel.Rules: // use the relabel rules defined for this pipeline
				return common.CustomTokenizer{Expr: s.getOrNewDiscoveryRelabelRules()}
			}
			return val
		}

		forwardTo := s.getOrNewProcessStageReceivers() // forward to process stage, which forwards to writers
		s.f.Body().AppendBlock(common.NewBlockWithOverrideFn(
			[]string{"loki", "source", "docker"},
			compLabel,
			toLokiSourceDocker(sd, forwardTo),
			overrideHook,
		))
	}
}

func toLokiSourceDocker(sd *moby.DockerSDConfig, forwardTo []loki.LogsReceiver) *loki_docker.Arguments {
	return &loki_docker.Arguments{
		Host:             sd.Host,
		Targets:          nil,
		ForwardTo:        forwardTo,
		Labels:           nil,
		RelabelRules:     flow_relabel.Rules{},
		HTTPClientConfig: common.ToHttpClientConfig(&sd.HTTPClientConfig),
		RefreshInterval:  time.Duration(sd.RefreshInterval),
	}
}

func toDiscoveryDocker(sdConfig *moby.DockerSDConfig) *docker.Arguments {
	if sdConfig == nil {
		return nil
	}

	return &docker.Arguments{
		Host:               sdConfig.Host,
		Port:               sdConfig.Port,
		HostNetworkingHost: sdConfig.HostNetworkingHost,
		RefreshInterval:    time.Duration(sdConfig.RefreshInterval),
		Filters:            toFlowDockerSDFilters(sdConfig.Filters),
		HTTPClientConfig:   *common.ToHttpClientConfig(&sdConfig.HTTPClientConfig),
	}
}

func toFlowDockerSDFilters(filters []moby.Filter) []docker.Filter {
	if len(filters) == 0 {
		return nil
	}

	flowFilters := make([]docker.Filter, len(filters))
	for i, filter := range filters {
		flowFilters[i] = docker.Filter{
			Name:   filter.Name,
			Values: filter.Values,
		}
	}

	return flowFilters
}
