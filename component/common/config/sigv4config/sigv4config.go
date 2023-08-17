package sigv4config

import (
	"github.com/grafana/agent/pkg/river/rivertypes"
	"github.com/prometheus/common/config"
	internal "github.com/prometheus/common/sigv4"
)

// SigV4Config is the configuration for signing remote write requests with
// AWS's SigV4 verification process. Empty values will be retrieved using the
// AWS default credentials chain.
type SigV4Config struct {
	Region    string            `river:"region,string,optional"`
	AccessKey string            `river:"access_key,string,optional"`
	SecretKey rivertypes.Secret `river:"secret_key,attr,optional"`
	Profile   string            `river:"profile,string,optional"`
	RoleARN   string            `river:"role_arn,string,optional"`
}

func (c *SigV4Config) ToInternal() internal.SigV4Config {
	return internal.SigV4Config{
		Region:    c.Region,
		AccessKey: c.AccessKey,
		SecretKey: config.Secret(c.SecretKey),
		Profile:   c.Profile,
		RoleARN:   c.RoleARN,
	}
}
