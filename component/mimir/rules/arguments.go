package rules

import (
	"time"

	"github.com/grafana/agent/pkg/flow/rivertypes"
)

type Arguments struct {
	ClientParams ClientArguments `river:"client,block"`
	SyncInterval time.Duration   `river:"sync_interval,attr,optional"`

	RuleSelector          LabelSelector `river:"rule_selector,block,optional"`
	RuleNamespaceSelector LabelSelector `river:"rule_namespace_selector,block,optional"`
}

type LabelSelector struct {
	MatchLabels      map[string]string `river:"match_labels,attr,optional"`
	MatchExpressions []MatchExpression `river:"match_expressions,attr,optional"`
}

type MatchExpression struct {
	Key      string   `river:"key,attr"`
	Operator string   `river:"operator,attr"`
	Values   []string `river:"values,attr"`
}

type ClientArguments struct {
	User            string            `river:"user,attr,optional"`
	Key             rivertypes.Secret `river:"key,attr,optional"`
	Address         string            `river:"address,attr"`
	ID              string            `river:"id,attr,optional"`
	TLS             TLSArguments      `river:"tls,block,optional"`
	UseLegacyRoutes bool              `river:"use_legacy_routes,attr,optional"`
	AuthToken       rivertypes.Secret `river:"auth_token,attr,optional"`
}

type TLSArguments struct {
	CertPath           string `river:"tls_cert_path,attr,optional"`
	KeyPath            string `river:"tls_key_path,attr,optional"`
	CAPath             string `river:"tls_ca_path,attr,optional"`
	ServerName         string `river:"tls_server_name,attr,optional"`
	InsecureSkipVerify bool   `river:"tls_insecure_skip_verify,attr,optional"`
	CipherSuites       string `river:"tls_cipher_suites,attr,optional"`
	MinVersion         string `river:"tls_min_version,attr,optional"`
}
