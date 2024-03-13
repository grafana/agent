package client

import "github.com/prometheus/prometheus/model/rulefmt"

type RuleGroup struct {
	rulefmt.RuleGroup `yaml:",inline"`
	SourceTenants     []string `yaml:"source_tenants,omitempty"`
}
