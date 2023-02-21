package redis

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverUnmarshal(t *testing.T) {
	riverConfig := `
		redis_addr                  = "localhost:6379"
		redis_user                  = "redis_user"
		redis_password_file         = "/tmp/pass"
		namespace                   = "namespace"
		config_command              = "TEST_CONFIG"
		check_keys                  = ["key1*", "cache_*"]
		check_key_groups            = ["other_key%d+"]
		check_key_groups_batch_size = 5000
		max_distinct_key_groups     = 50
		check_single_keys           = ["particular_key"]
		check_streams               = ["stream1*"]
		check_single_streams        = ["particular_stream"]
		count_keys                  = ["count_key1", "count_key2"]
		script_path                 = "/tmp/metrics-script.lua"
		connection_timeout          = "7s"
		tls_client_key_file         = "/tmp/client-key.pem"
		tls_client_cert_file        = "/tmp/client-cert.pem"
		tls_ca_cert_file            = "/tmp/ca-cert.pem"
		set_client_name             = false
		is_tile38                   = true
		export_client_list          = false
		export_client_port          = true
		redis_metrics_only          = false
		ping_on_connect             = true
		incl_system_metrics         = true
		skip_tls_verification       = false
	
	`
	var cfg Config
	err := river.Unmarshal([]byte(riverConfig), &cfg)

	require.NoError(t, err)
	require.Equal(t, "localhost:6379", cfg.RedisAddr)
	require.Equal(t, "redis_user", cfg.RedisUser)
	require.Equal(t, "/tmp/pass", cfg.RedisPasswordFile)
	require.Equal(t, "namespace", cfg.Namespace)
	require.Equal(t, "TEST_CONFIG", cfg.ConfigCommand)

	require.Equal(t, []string{"key1*", "cache_*"}, cfg.CheckKeys)
	require.Equal(t, []string{"other_key%d+"}, cfg.CheckKeyGroups)
	require.Equal(t, []string{"particular_key"}, cfg.CheckSingleKeys)
	require.Equal(t, int64(5000), cfg.CheckKeyGroupsBatchSize)
	require.Equal(t, int64(50), cfg.MaxDistinctKeyGroups)

	require.Equal(t, []string{"stream1*"}, cfg.CheckStreams)
	require.Equal(t, []string{"particular_stream"}, cfg.CheckSingleStreams)
	require.Equal(t, []string{"count_key1", "count_key2"}, cfg.CountKeys)

	require.Equal(t, "/tmp/metrics-script.lua", cfg.ScriptPath)
	require.Equal(t, 7*time.Second, cfg.ConnectionTimeout)

	require.Equal(t, "/tmp/client-key.pem", cfg.TLSClientKeyFile)
	require.Equal(t, "/tmp/client-cert.pem", cfg.TLSClientCertFile)
	require.Equal(t, "/tmp/ca-cert.pem", cfg.TLSCaCertFile)

	require.False(t, cfg.SetClientName)
	require.True(t, cfg.IsTile38)
	require.False(t, cfg.ExportClientList)
	require.True(t, cfg.ExportClientPort)
	require.False(t, cfg.RedisMetricsOnly)
	require.True(t, cfg.PingOnConnect)
	require.True(t, cfg.InclSystemMetrics)
	require.False(t, cfg.SkipTLSVerification)
}

func TestRiverConvert(t *testing.T) {
	orig := Config{
		RedisAddr:         "localhost:6379",
		RedisUser:         "redis_user",
		RedisPasswordFile: "/tmp/pass",
		Namespace:         "namespace",
		ConfigCommand:     "TEST_CONFIG",

		CheckKeys:               []string{"key1*", "cache_*"},
		CheckKeyGroups:          []string{"other_key%d+"},
		CheckSingleKeys:         []string{"particular_key"},
		CountKeys:               []string{"count_key1", "count_key2"},
		CheckKeyGroupsBatchSize: 5000,
		MaxDistinctKeyGroups:    50,

		CheckStreams:       []string{"stream1*", "stream2*"},
		CheckSingleStreams: []string{"particular_stream"},

		ScriptPath:        "/tmp/metrics-script.lua",
		ConnectionTimeout: 7 * time.Second,

		TLSClientKeyFile:  "/tmp/client-key.pem",
		TLSClientCertFile: "/tmp/client-cert.pem",
		TLSCaCertFile:     "/tmp/ca-cert.pem",

		SetClientName:       false,
		IsTile38:            true,
		ExportClientList:    false,
		ExportClientPort:    true,
		RedisMetricsOnly:    false,
		PingOnConnect:       true,
		InclSystemMetrics:   true,
		SkipTLSVerification: false,
	}
	converted := orig.Convert()

	require.Equal(t, "localhost:6379", converted.RedisAddr)
	require.Equal(t, "redis_user", converted.RedisUser)
	require.Equal(t, "/tmp/pass", converted.RedisPasswordFile)
	require.Equal(t, "namespace", converted.Namespace)
	require.Equal(t, "TEST_CONFIG", converted.ConfigCommand)

	require.Equal(t, "key1*,cache_*", converted.CheckKeys)
	require.Equal(t, "other_key%d+", converted.CheckKeyGroups)
	require.Equal(t, "particular_key", converted.CheckSingleKeys)
	require.Equal(t, "count_key1,count_key2", converted.CountKeys)
	require.Equal(t, int64(5000), converted.CheckKeyGroupsBatchSize)
	require.Equal(t, int64(50), converted.MaxDistinctKeyGroups)

	require.Equal(t, "stream1*,stream2*", converted.CheckStreams)
	require.Equal(t, "particular_stream", converted.CheckSingleStreams)

	require.Equal(t, "/tmp/metrics-script.lua", converted.ScriptPath)
	require.Equal(t, 7*time.Second, converted.ConnectionTimeout)

	require.Equal(t, "/tmp/client-key.pem", converted.TLSClientKeyFile)
	require.Equal(t, "/tmp/client-cert.pem", converted.TLSClientCertFile)
	require.Equal(t, "/tmp/ca-cert.pem", converted.TLSCaCertFile)

	require.False(t, converted.SetClientName)
	require.True(t, converted.IsTile38)
	require.False(t, converted.ExportClientList)
	require.True(t, converted.ExportClientPort)
	require.False(t, converted.RedisMetricsOnly)
	require.True(t, converted.PingOnConnect)
	require.True(t, converted.InclSystemMetrics)
	require.False(t, converted.SkipTLSVerification)
}
