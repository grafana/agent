package main

import (
	"testing"

	"github.com/grafana/agent/pkg/integrations/v1/node_exporter"

	"github.com/grafana/agent/pkg/integrations/shared"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestCodeGenV1(t *testing.T) {
	codegen := codeGen{}
	gen := codegen.createV1Config(v1Configs)
	assert.True(t, len(gen) > 0)
}

func TestCodeGenV1Unmarshal(t *testing.T) {
	configStr := `
node_exporter:
  enabled: true
  metric_relabel_configs: 
  - source_labels: [__address__]
    target_label: "banana"
    replacement: "apple"
`
	cfg := &testWinExporter{}
	err := yaml.UnmarshalStrict([]byte(configStr), cfg)
	assert.NoError(t, err)
	assert.True(t, cfg != nil)
	assert.True(t, cfg.NodeExporter != nil)
	assert.True(t, cfg.NodeExporter.Enabled)
	assert.Len(t, cfg.NodeExporter.MetricRelabelConfigs, 1)
	assert.True(t, cfg.NodeExporter.MetricRelabelConfigs[0].TargetLabel == "banana")
}

type testWinExporter struct {
	NodeExporter *NodeExporter `yaml:"node_exporter"`
}

type NodeExporter struct {
	node_exporter.Config `yaml:",omitempty,inline"`
	shared.Common        `yaml:",omitempty,inline"`
}

func (c *NodeExporter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Config = node_exporter.DefaultConfig
	type plain NodeExporter
	return unmarshal((*plain)(c))
}
