package loki

import "github.com/grafana/agent/pkg/util"

// LatestConfig is the current major config version.
type LatestConfig = ConfigV1

// Config is a versioned Config struct.
type Config struct {
	Version string
	Config  LatestConfig
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var version util.Versioned
	if err := unmarshal(&version); err != nil {
		return err
	}
	c.Version = string(version)

	switch c.Version {
	case "v0", "":
		var v0 ConfigV0
		if err := unmarshal(&v0); err != nil {
			return err
		}
		v1, err := v0.Upgrade()
		if err != nil {
			return err
		}
		c.Config = *v1
	case "v1":
		var v1 ConfigV1
		if err := unmarshal(&v1); err != nil {
			return err
		}
		c.Config = v1
	}

	return nil
}
