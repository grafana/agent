package app_agent_receiver

import (
	"fmt"
	"net"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ErrEmptyDBPathGeoIPConfig = "db path cannot be empty when geoip is enabled"
	ErrEmptyDBTypeGeoIPConfig = "db type should be either city or as when geoip is enabled"
)

// GeoIPProvider is interface for providing geoip information via a locally mounted GeoIP database.
type GeoIPProvider interface {
	TransformMetas(mt *Meta, sourceIP net.IP) *Meta
}

// GeoIPProvider is a wrapper for the MaxMind geoip2.Reader
type GeoIP2 struct {
	logger log.Logger
	db     *geoip2.Reader
	cfgs   *GeoIPConfig
}

// NewGeoIPProvider creates an instance of GeoIPProvider.
func NewGeoIPProvider(l log.Logger, config GeoIPConfig, reg prometheus.Registerer) GeoIPProvider {

	err := validateGeoIPConfig(&config)
	if err != nil {
		panic(err) //TODO Is panicing the correct way to handle this?
	}

	var db *geoip2.Reader

	if config.Enabled {
		db, err = geoip2.Open(config.DB)
		if err != nil {
			panic(err) //TODO Is panicing the correct way to handle this?
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

// getGeoIPData will query the geoip2 database for the given IP address and return the geoip2.City record.
func (gp *GeoIP2) getGeoIPData(sourceIP net.IP) (*geoip2.City, error) {

	record, err := gp.db.City(sourceIP)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// mapGeoIP2CityToMetas will map the geoip2.City record to the app_agent_receiver.Metas.GeoIP struct.
func mapGeoIP2CityToMetas(mt *Meta, record *geoip2.City, clientIP net.IP) (*Meta, error) {
	country, ok := record.Country.Names["en"]
	if !ok {
		return nil, fmt.Errorf("no English name for country")
	}

	city, ok := record.City.Names["en"]
	if !ok {
		return nil, fmt.Errorf("no English name for city")
	}

	continent, ok := record.Continent.Names["en"]
	if !ok {
		return nil, fmt.Errorf("no English name for continent")
	}

	subdivisionName, subdivisionCode := "", ""
	if len(record.Subdivisions) > 0 {
		subdivisionName, ok = record.Subdivisions[0].Names["en"] // TODO: Copilot generated first. Example has last index.
		if !ok {
			return nil, fmt.Errorf("no English name for subdivision")
		}
		subdivisionCode = record.Subdivisions[0].IsoCode
	}

	mt.GeoIP = GeoIP{
		ClientIP:        clientIP, // Set this value from the client's IP
		LocationLat:     record.Location.Latitude,
		LocationLong:    record.Location.Longitude,
		CityName:        city,
		CountryName:     country,
		ContinentName:   continent,
		ContinentCode:   record.Continent.Code,
		PostalCode:      record.Postal.Code,
		Timezone:        record.Location.TimeZone,
		SubdivisionName: subdivisionName,
		SubdivisionCode: subdivisionCode,
	}

	return mt, nil
}

// TransformException will attempt to populate the metas with geo IP data. If the geo IP data is not available, the
// metas will be returned as is.
func (gp *GeoIP2) TransformMetas(mt *Meta, clientIP net.IP) *Meta {
	if clientIP == nil {
		level.Warn(gp.logger).Log("msg", "Client IP is nil")
		return mt
	}

	// Query GeoIP db
	geoIpCityRecord, err := gp.getGeoIPData(clientIP)
	if err != nil {
		level.Error(gp.logger).Log("msg", "Error querying geo IP2 database", "err", err)
		return mt
	}

	//  Populate metas with geo IP data
	transformedMeta, err := mapGeoIP2CityToMetas(mt, geoIpCityRecord, clientIP)
	if err != nil {
		level.Error(gp.logger).Log("msg", "Error populating metas with geo IP data", "err", err)
		return mt
	}

	return transformedMeta
}
