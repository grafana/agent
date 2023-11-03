package app_agent_receiver

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
	"net"
)

const (
	ErrEmptyGeoIPStageConfig  = "geoip stage config cannot be empty"
	ErrEmptyDBPathGeoIPConfig = "db path cannot be empty when geoip is enabled"
	ErrEmptyDBTypeGeoIPConfig = "db type should be either city or as when geoip is enabled"
)

// GeoIPProvider is interface for providing geoip information via a locally mounted GeoIP database.
type GeoIPProvider interface {
	GetGeoIPData(sourceIP net.IP) (string, error)
}

// GeoIPProvider is a wrapper for the MaxMind geoip2.Reader
type GeoIP2 struct {
	logger log.Logger
	db     *geoip2.Reader
	cfgs   *GeoIPConfig
}

// NewGeoIPProvider creates an instance of GeoIPProvider.
// httpClient and fileService will be instantiated to defaults if nil is provided
func NewGeoIPProvider(l log.Logger, config GeoIPConfig, reg prometheus.Registerer) GeoIPProvider {

	l.Log("msg", "Initializing GeoIPProvider")
	err := validateGeoIPConfig(&config)
	if err != nil {
		panic(err)
	}

	var db *geoip2.Reader

	if config.Enabled {
		db, err = geoip2.Open(config.DB)
		if err != nil {
			panic(err)
		}
	}

	return &GeoIP2{
		logger: l,
		db:     db,
		cfgs:   &config,
	}
}

func validateGeoIPConfig(c *GeoIPConfig) error {
	if c != nil && c.Enabled {
		if c.DB == "" {
			return fmt.Errorf(ErrEmptyDBPathGeoIPConfig)
		}

		if c.DBType == "" {
			return fmt.Errorf(ErrEmptyDBTypeGeoIPConfig)
		}
	}

	return nil
}

func (geo *GeoIP2) GetGeoIPData(sourceIP net.IP) (string, error) {
	return sourceIP.String(), nil
}
