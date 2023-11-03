package redis

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/redis_exporter"
	"github.com/grafana/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.redis",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.New(createExporter, "redis"),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

// DefaultArguments holds non-zero default options for Arguments when it is
// unmarshaled from river.
var DefaultArguments = Arguments{
	IncludeExporterMetrics:  false,
	Namespace:               "redis",
	ConfigCommand:           "CONFIG",
	ConnectionTimeout:       15 * time.Second,
	SetClientName:           true,
	CheckKeyGroupsBatchSize: 10000,
	MaxDistinctKeyGroups:    100,
	ExportKeyValues:         true,
}

type Arguments struct {
	IncludeExporterMetrics bool `river:"include_exporter_metrics,attr,optional"`

	// exporter-specific config.
	//
	// The exporter binary config differs to this, but these
	// are the only fields that are relevant to the exporter struct.
	RedisAddr               string            `river:"redis_addr,attr"`
	RedisUser               string            `river:"redis_user,attr,optional"`
	RedisPassword           rivertypes.Secret `river:"redis_password,attr,optional"`
	RedisPasswordFile       string            `river:"redis_password_file,attr,optional"`
	RedisPasswordMapFile    string            `river:"redis_password_map_file,attr,optional"`
	Namespace               string            `river:"namespace,attr,optional"`
	ConfigCommand           string            `river:"config_command,attr,optional"`
	CheckKeys               []string          `river:"check_keys,attr,optional"`
	CheckKeyGroups          []string          `river:"check_key_groups,attr,optional"`
	CheckKeyGroupsBatchSize int64             `river:"check_key_groups_batch_size,attr,optional"`
	MaxDistinctKeyGroups    int64             `river:"max_distinct_key_groups,attr,optional"`
	CheckSingleKeys         []string          `river:"check_single_keys,attr,optional"`
	CheckStreams            []string          `river:"check_streams,attr,optional"`
	CheckSingleStreams      []string          `river:"check_single_streams,attr,optional"`
	ExportKeyValues         bool              `river:"export_key_values,attr,optional"`
	CountKeys               []string          `river:"count_keys,attr,optional"`
	ScriptPath              string            `river:"script_path,attr,optional"`
	ScriptPaths             []string          `river:"script_paths,attr,optional"`
	ConnectionTimeout       time.Duration     `river:"connection_timeout,attr,optional"`
	TLSClientKeyFile        string            `river:"tls_client_key_file,attr,optional"`
	TLSClientCertFile       string            `river:"tls_client_cert_file,attr,optional"`
	TLSCaCertFile           string            `river:"tls_ca_cert_file,attr,optional"`
	SetClientName           bool              `river:"set_client_name,attr,optional"`
	IsTile38                bool              `river:"is_tile38,attr,optional"`
	IsCluster               bool              `river:"is_cluster,attr,optional"`
	ExportClientList        bool              `river:"export_client_list,attr,optional"`
	ExportClientPort        bool              `river:"export_client_port,attr,optional"`
	RedisMetricsOnly        bool              `river:"redis_metrics_only,attr,optional"`
	PingOnConnect           bool              `river:"ping_on_connect,attr,optional"`
	InclSystemMetrics       bool              `river:"incl_system_metrics,attr,optional"`
	SkipTLSVerification     bool              `river:"skip_tls_verification,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	if a.ScriptPath != "" && len(a.ScriptPaths) > 0 {
		return fmt.Errorf("only one of script_path and script_paths should be specified")
	}
	return nil
}

func (a *Arguments) Convert() *redis_exporter.Config {
	var scriptPath string
	if a.ScriptPath != "" {
		scriptPath = a.ScriptPath
	} else if len(a.ScriptPaths) > 0 {
		scriptPath = strings.Join(a.ScriptPaths, ",")
	}

	return &redis_exporter.Config{
		IncludeExporterMetrics:  a.IncludeExporterMetrics,
		RedisAddr:               a.RedisAddr,
		RedisUser:               a.RedisUser,
		RedisPassword:           config_util.Secret(a.RedisPassword),
		RedisPasswordFile:       a.RedisPasswordFile,
		RedisPasswordMapFile:    a.RedisPasswordMapFile,
		Namespace:               a.Namespace,
		ConfigCommand:           a.ConfigCommand,
		CheckKeys:               strings.Join(a.CheckKeys, ","),
		CheckKeyGroups:          strings.Join(a.CheckKeyGroups, ","),
		CheckKeyGroupsBatchSize: a.CheckKeyGroupsBatchSize,
		MaxDistinctKeyGroups:    a.MaxDistinctKeyGroups,
		CheckSingleKeys:         strings.Join(a.CheckSingleKeys, ","),
		CheckStreams:            strings.Join(a.CheckStreams, ","),
		CheckSingleStreams:      strings.Join(a.CheckSingleStreams, ","),
		ExportKeyValues:         a.ExportKeyValues,
		CountKeys:               strings.Join(a.CountKeys, ","),
		ScriptPath:              scriptPath,
		ConnectionTimeout:       a.ConnectionTimeout,
		TLSClientKeyFile:        a.TLSClientKeyFile,
		TLSClientCertFile:       a.TLSClientCertFile,
		TLSCaCertFile:           a.TLSCaCertFile,
		SetClientName:           a.SetClientName,
		IsTile38:                a.IsTile38,
		IsCluster:               a.IsCluster,
		ExportClientList:        a.ExportClientList,
		ExportClientPort:        a.ExportClientPort,
		RedisMetricsOnly:        a.RedisMetricsOnly,
		PingOnConnect:           a.PingOnConnect,
		InclSystemMetrics:       a.InclSystemMetrics,
		SkipTLSVerification:     a.SkipTLSVerification,
	}
}
