package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jmespath/go-jmespath"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/prometheus/common/model"
)

var (
	ErrEmptyGeoIPStageConfig                = errors.New("geoip stage config cannot be empty")
	ErrEmptyDBPathGeoIPStageConfig          = errors.New("db path cannot be empty")
	ErrEmptySourceGeoIPStageConfig          = errors.New("source cannot be empty")
	ErrEmptyDBTypeGeoIPStageConfig          = errors.New("db type should be either city or asn")
	ErrEmptyDBTypeAndValuesGeoIPStageConfig = errors.New("db type or values need to be set")
)

type GeoIPFields int

const (
	CITYNAME GeoIPFields = iota
	COUNTRYNAME
	CONTINENTNAME
	CONTINENTCODE
	LOCATION
	POSTALCODE
	TIMEZONE
	SUBDIVISIONNAME
	SUBDIVISIONCODE
)

var fields = map[GeoIPFields]string{
	CITYNAME:        "geoip_city_name",
	COUNTRYNAME:     "geoip_country_name",
	CONTINENTNAME:   "geoip_continent_name",
	CONTINENTCODE:   "geoip_continent_code",
	LOCATION:        "geoip_location",
	POSTALCODE:      "geoip_postal_code",
	TIMEZONE:        "geoip_timezone",
	SUBDIVISIONNAME: "geoip_subdivision_name",
	SUBDIVISIONCODE: "geoip_subdivision_code",
}

// GeoIPConfig represents GeoIP stage config
type GeoIPConfig struct {
	DB            string            `river:"db,attr"`
	Source        *string           `river:"source,attr"`
	DBType        string            `river:"db_type,attr,optional"`
	CustomLookups map[string]string `river:"custom_lookups,attr,optional"`
}

func validateGeoIPConfig(c GeoIPConfig) (map[string]*jmespath.JMESPath, error) {
	if c.DB == "" {
		return nil, ErrEmptyDBPathGeoIPStageConfig
	}
	if c.Source != nil && *c.Source == "" {
		return nil, ErrEmptySourceGeoIPStageConfig
	}

	if c.DBType == "" && c.CustomLookups == nil {
		return nil, ErrEmptyDBTypeAndValuesGeoIPStageConfig
	}

	switch c.DBType {
	case "", "asn", "city":
	default:
		return nil, ErrEmptyDBTypeGeoIPStageConfig
	}

	if c.CustomLookups == nil {
		return nil, nil
	}

	expressions := map[string]*jmespath.JMESPath{}
	for key, expr := range c.CustomLookups {
		var err error
		jmes := expr

		// If there is no expression, use the name as the expression.
		if expr == "" {
			jmes = key
		}

		expressions[key], err = jmespath.Compile(jmes)
		if err != nil {
			return nil, errors.New(ErrCouldNotCompileJMES)
		}
	}
	return expressions, nil
}

func newGeoIPStage(logger log.Logger, config GeoIPConfig) (Stage, error) {
	valuesExpressions, err := validateGeoIPConfig(config)
	if err != nil {
		return nil, err
	}

	mmdb, err := maxminddb.Open(config.DB)
	if err != nil {
		return nil, err
	}

	return &geoIPStage{
		mmdb:              mmdb,
		logger:            logger,
		cfgs:              config,
		valuesExpressions: valuesExpressions,
	}, nil
}

type geoIPStage struct {
	logger            log.Logger
	mmdb              *maxminddb.Reader
	cfgs              GeoIPConfig
	valuesExpressions map[string]*jmespath.JMESPath
}

// Run implements Stage
func (g *geoIPStage) Run(in chan Entry) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		defer g.close()
		for e := range in {
			g.process(e.Labels, e.Extracted)
			out <- e
		}
	}()
	return out
}

// Name implements Stage
func (g *geoIPStage) Name() string {
	return StageTypeGeoIP
}

func (g *geoIPStage) process(_ model.LabelSet, extracted map[string]interface{}) {
	var ip net.IP
	if g.cfgs.Source != nil {
		if _, ok := extracted[*g.cfgs.Source]; !ok {
			if Debug {
				level.Debug(g.logger).Log("msg", "source does not exist in the set of extracted values", "source", *g.cfgs.Source)
			}
			return
		}

		value, err := getString(extracted[*g.cfgs.Source])
		if err != nil {
			if Debug {
				level.Debug(g.logger).Log("msg", "failed to convert source value to string", "source", *g.cfgs.Source, "err", err, "type", reflect.TypeOf(extracted[*g.cfgs.Source]))
			}
			return
		}
		ip = net.ParseIP(value)
		if ip == nil {
			level.Error(g.logger).Log("msg", "source is not an ip", "source", value)
			return
		}
	}
	if g.cfgs.DBType != "" {
		switch g.cfgs.DBType {
		case "city":
			var record geoip2.City
			err := g.mmdb.Lookup(ip, &record)
			if err != nil {
				level.Error(g.logger).Log("msg", "unable to get City record for the ip", "err", err, "ip", ip)
				return
			}
			g.populateExtractedWithCityData(extracted, &record)
		case "asn":
			var record geoip2.ASN
			err := g.mmdb.Lookup(ip, &record)
			if err != nil {
				level.Error(g.logger).Log("msg", "unable to get ASN record for the ip", "err", err, "ip", ip)
				return
			}
			g.populateExtractedWithASNData(extracted, &record)
		default:
			level.Error(g.logger).Log("msg", "unknown database type")
		}
	}
	if g.valuesExpressions != nil {
		g.populateExtractedWithCustomFields(ip, extracted)
	}
}

func (g *geoIPStage) close() {
	if err := g.mmdb.Close(); err != nil {
		level.Error(g.logger).Log("msg", "error while closing mmdb", "err", err)
	}
}

func (g *geoIPStage) populateExtractedWithCityData(extracted map[string]interface{}, record *geoip2.City) {
	for field, label := range fields {
		switch field {
		case CITYNAME:
			cityName := record.City.Names["en"]
			if cityName != "" {
				extracted[label] = cityName
			}
		case COUNTRYNAME:
			contryName := record.Country.Names["en"]
			if contryName != "" {
				extracted[label] = contryName
			}
		case CONTINENTNAME:
			continentName := record.Continent.Names["en"]
			if continentName != "" {
				extracted[label] = continentName
			}
		case CONTINENTCODE:
			continentCode := record.Continent.Code
			if continentCode != "" {
				extracted[label] = continentCode
			}
		case POSTALCODE:
			postalCode := record.Postal.Code
			if postalCode != "" {
				extracted[label] = postalCode
			}
		case TIMEZONE:
			timezone := record.Location.TimeZone
			if timezone != "" {
				extracted[label] = timezone
			}
		case LOCATION:
			latitude := record.Location.Latitude
			longitude := record.Location.Longitude
			if latitude != 0 || longitude != 0 {
				extracted[fmt.Sprintf("%s_latitude", label)] = latitude
				extracted[fmt.Sprintf("%s_longitude", label)] = longitude
			}
		case SUBDIVISIONNAME:
			if len(record.Subdivisions) > 0 {
				// we get most specific subdivision https://dev.maxmind.com/release-note/most-specific-subdivision-attribute-added/
				subdivisionName := record.Subdivisions[len(record.Subdivisions)-1].Names["en"]
				if subdivisionName != "" {
					extracted[label] = subdivisionName
				}
			}
		case SUBDIVISIONCODE:
			if len(record.Subdivisions) > 0 {
				subdivisionCode := record.Subdivisions[len(record.Subdivisions)-1].IsoCode
				if subdivisionCode != "" {
					extracted[label] = subdivisionCode
				}
			}
		default:
			level.Error(g.logger).Log("msg", "unknown geoip field")
		}
	}
}

func (g *geoIPStage) populateExtractedWithASNData(extracted map[string]interface{}, record *geoip2.ASN) {
	autonomousSystemNumber := record.AutonomousSystemNumber
	autonomousSystemOrganization := record.AutonomousSystemOrganization
	if autonomousSystemNumber != 0 {
		extracted["geoip_autonomous_system_number"] = autonomousSystemNumber
	}
	if autonomousSystemOrganization != "" {
		extracted["geoip_autonomous_system_organization"] = autonomousSystemOrganization
	}
}

func (g *geoIPStage) populateExtractedWithCustomFields(ip net.IP, extracted map[string]interface{}) {
	var record any
	if err := g.mmdb.Lookup(ip, &record); err != nil {
		level.Error(g.logger).Log("msg", "unable to lookup record for the ip", "err", err, "ip", ip)
		return
	}

	for key, expr := range g.valuesExpressions {
		r, err := expr.Search(record)
		if err != nil {
			level.Error(g.logger).Log("msg", "failed to search JMES expression", "err", err)
			continue
		}
		if r == nil {
			if Debug {
				level.Debug(g.logger).Log("msg", "failed find a result with JMES expression", "key", key)
			}
			continue
		}
		extracted[key] = r
	}
}
