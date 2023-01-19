package cloudflare

import (
	"fmt"
	"time"

	cft "github.com/grafana/agent/component/loki/source/cloudflare/internal/cloudflaretarget"
	"github.com/grafana/agent/pkg/river"
	"github.com/prometheus/common/model"
)

// Config defines how to create a Cloudflare logs target.
type Config struct {
	APIToken   string            `river:"api_token,attr"`
	ZoneID     string            `river:"zone_id,attr"`
	Labels     map[string]string `river:"labels,attr,optional"`
	Workers    int               `river:"workers,attr,optional"`
	PullRange  time.Duration     `river:"pull_range,attr,optional"`
	FieldsType string            `river:"fields_type,attr,optional"`
}

// Convert bridges the River and cloudflaretarget Config structs.
func (c Config) Convert() *cft.Config {
	lbls := make(model.LabelSet, len(c.Labels))
	for k, v := range c.Labels {
		lbls[model.LabelName(k)] = model.LabelValue(v)
	}
	return &cft.Config{
		APIToken:   c.APIToken,
		ZoneID:     c.APIToken,
		Labels:     lbls,
		Workers:    c.Workers,
		PullRange:  model.Duration(c.PullRange),
		FieldsType: c.FieldsType,
	}
}

// DefaultConfig sets the configuration defaults.
var DefaultConfig = Config{
	Workers:    3,
	PullRange:  1 * time.Minute,
	FieldsType: string(cft.FieldsTypeDefault),
}

var _ river.Unmarshaler = (*Config)(nil)

// UnmarshalRiver implements the unmarshaller
func (c *Config) UnmarshalRiver(f func(v interface{}) error) error {
	*c = DefaultConfig
	type config Config
	err := f((*config)(c))
	if err != nil {
		return err
	}
	if c.PullRange < 0 {
		return fmt.Errorf("pull_range must be a positive duration")
	}
	_, err = cft.Fields(cft.FieldsType(c.FieldsType))
	if err != nil {
		return fmt.Errorf("invalid fields_type set; the available values are 'default', 'minimal', 'extended' and 'all'")
	}
	return nil
}
