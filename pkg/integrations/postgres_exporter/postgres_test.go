package postgres_exporter //nolint:golint

import (
	"testing"

	"github.com/grafana/agent/pkg/config"
)

func TestConfig_SecretPostgres(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  postgres_exporter:
    enabled: true
    data_source_names: ["secret_password_in_uri","secret_password_in_uri_2"]
`
	config.CheckSecret(t, stringCfg, "secret_password_in_uri")
	config.CheckSecret(t, stringCfg, "secret_password_in_uri_2")
}
