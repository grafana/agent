package config

import "os"

type AppConfig interface {
	MetricsPath() string
	ListenPort() string
	LogLevel() string
	ApplicationName() string
}

type BaseConfig struct {
	metricsPath string
	listenPort  string
	logLevel    string
	appName     string
}

func (c BaseConfig) MetricsPath() string {
	return c.metricsPath
}

func (c BaseConfig) ListenPort() string {
	return c.listenPort
}

func (c BaseConfig) LogLevel() string {
	return c.logLevel
}

func (c BaseConfig) ApplicationName() string {
	return c.appName
}

func Init() BaseConfig {

	appConfig := BaseConfig{
		metricsPath: GetEnv("METRICS_PATH", "/metrics"),
		listenPort:  GetEnv("LISTEN_PORT", "8080"),
		logLevel:    GetEnv("LOG_LEVEL", "debug"),
		appName:     GetEnv("APP_NAME", "app"),
	}
	return appConfig
}

// GetEnv - Allows us to supply a fallback option if nothing specified
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
