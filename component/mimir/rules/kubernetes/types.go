package rules

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component/common/config"
)

type Arguments struct {
	Address              string                  `river:"address,attr"`
	TenantID             string                  `river:"tenant_id,attr,optional"`
	UseLegacyRoutes      bool                    `river:"use_legacy_routes,attr,optional"`
	HTTPClientConfig     config.HTTPClientConfig `river:",squash"`
	SyncInterval         time.Duration           `river:"sync_interval,attr,optional"`
	MimirNameSpacePrefix string                  `river:"mimir_namespace_prefix,attr,optional"`

	RuleSelector          LabelSelector `river:"rule_selector,block,optional"`
	RuleNamespaceSelector LabelSelector `river:"rule_namespace_selector,block,optional"`
}

var DefaultArguments = Arguments{
	SyncInterval:         30 * time.Second,
	MimirNameSpacePrefix: "agent",
	HTTPClientConfig:     config.DefaultHTTPClientConfig,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if args.SyncInterval <= 0 {
		return fmt.Errorf("sync_interval must be greater than 0")
	}
	if args.MimirNameSpacePrefix == "" {
		return fmt.Errorf("mimir_namespace_prefix must not be empty")
	}

	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	return args.HTTPClientConfig.Validate()
}

type LabelSelector struct {
	MatchLabels      map[string]string `river:"match_labels,attr,optional"`
	MatchExpressions []MatchExpression `river:"match_expression,block,optional"`
}

type MatchExpression struct {
	Key      string   `river:"key,attr"`
	Operator string   `river:"operator,attr"`
	Values   []string `river:"values,attr,optional"`
}
