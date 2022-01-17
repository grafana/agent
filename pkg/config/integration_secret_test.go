package config

import "testing"

func TestConfig_SecretGithub(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  github_exporter:
    enabled: true
    api_token: secret_api`
	CheckSecret(t, stringCfg, "secret_api")
}

func TestConfig_SecretKafkaPassword(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  kafka_exporter:
    enabled: true
    sasl_password: secret_password
`
	CheckSecret(t, stringCfg, "secret_password")
}

func TestConfig_SecretMongoDB(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  mongodb_exporter:
    enabled: true
    mongodb_uri: secret_password_in_uri
`
	CheckSecret(t, stringCfg, "secret_password_in_uri")
}

func TestConfig_SecretMysqlD(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  mysqld_exporter:
    enabled: true
    data_source_name: root:secret_password@myserver:3306`
	CheckSecret(t, stringCfg, "secret_password")
}

func TestConfig_SecretPostgres(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  postgres_exporter:
    enabled: true
    data_source_names: ["secret_password_in_uri","secret_password_in_uri_2"]
`
	CheckSecret(t, stringCfg, "secret_password_in_uri")
	CheckSecret(t, stringCfg, "secret_password_in_uri_2")

}

func TestConfig_SecretRedisPassword(t *testing.T) {
	stringCfg := `
prometheus:
  wal_directory: /tmp/agent
integrations:
  redis_exporter:
    enabled: true
    redis_password: secret_password
`
	CheckSecret(t, stringCfg, "secret_password")
}
