package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	iso8601 "github.com/ChannelMeter/iso8601duration"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/azure-metrics-exporter/config"
)

const (
	PrometheusMetricNameDefault = "azurerm_resource_metric"
)

type (
	RequestMetricSettings struct {
		Name            string
		Subscriptions   []string
		ResourceType    string
		Filter          string
		Timespan        string
		Interval        *string
		Metrics         []string
		MetricNamespace string
		Aggregations    []string

		// needed for dimension support
		MetricTop     *int32
		MetricFilter  string
		MetricOrderBy string

		MetricTemplate string
		HelpTemplate   string

		TagLabels []string

		// cache
		Cache *time.Duration
	}
)

func NewRequestMetricSettingsForAzureResourceApi(r *http.Request, opts config.Opts) (RequestMetricSettings, error) {
	settings, err := NewRequestMetricSettings(r, opts)
	if err != nil {
		return settings, err
	}

	if r.URL.Path == config.ProbeMetricsResourceUrl {
		return settings, nil
	} else if settings.ResourceType != "" && settings.Filter != "" {
		return settings, fmt.Errorf("parameter \"resourceType\" and \"filter\" are mutually exclusive")
	} else if settings.ResourceType != "" {
		settings.Filter = fmt.Sprintf(
			"resourceType eq '%s'",
			strings.ReplaceAll(settings.ResourceType, "'", "\\'"),
		)
	} else if settings.Filter == "" {
		return settings, fmt.Errorf("parameter \"resourceType\" or \"filter\" is missing")
	}

	return settings, nil
}

func NewRequestMetricSettings(r *http.Request, opts config.Opts) (RequestMetricSettings, error) {
	ret := RequestMetricSettings{}
	params := r.URL.Query()

	ret.TagLabels = opts.Azure.ResourceTags

	// param name
	ret.Name = paramsGetWithDefault(params, "name", PrometheusMetricNameDefault)

	// param subscription
	if subscriptionList, err := paramsGetListRequired(params, "subscription"); err == nil {
		for _, subscription := range subscriptionList {
			subscription = strings.TrimSpace(subscription)
			ret.Subscriptions = append(ret.Subscriptions, subscription)
		}
	} else {
		return ret, err
	}

	// param filter
	ret.ResourceType = paramsGetWithDefault(params, "resourceType", "")
	ret.Filter = paramsGetWithDefault(params, "filter", "")

	// param timespan
	ret.Timespan = paramsGetWithDefault(params, "timespan", "PT1M")

	// param interval
	if val := params.Get("interval"); val != "" {
		ret.Interval = &val
	}

	// param metric
	if val, err := paramsGetList(params, "metric"); err == nil {
		ret.Metrics = val
	} else {
		return ret, err
	}

	// param metricNamespace
	ret.MetricNamespace = paramsGetWithDefault(params, "metricNamespace", "")

	// param aggregation
	if val, err := paramsGetList(params, "aggregation"); err == nil {
		ret.Aggregations = val
	} else {
		return ret, err
	}

	// param metricTop
	if val := params.Get("metricTop"); val != "" {
		valInt64, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return ret, err
		}
		valInt32 := int32(valInt64)
		ret.MetricTop = &valInt32
	}

	// param metricFilter
	ret.MetricFilter = paramsGetWithDefault(params, "metricFilter", "")

	// param metricOrderBy
	ret.MetricOrderBy = paramsGetWithDefault(params, "metricOrderBy", "")

	// param template
	ret.MetricTemplate = paramsGetWithDefault(params, "template", opts.Metrics.Template)

	// param help
	ret.HelpTemplate = paramsGetWithDefault(params, "help", opts.Metrics.Help)

	// param cache (timespan as default)
	if opts.Prober.Cache {
		cacheDefaultDuration, err := iso8601.FromString(ret.Timespan)
		cacheDefaultDurationString := ""
		if err == nil {
			cacheDefaultDurationString = cacheDefaultDuration.ToDuration().String()
		}

		// get value from query (with default from timespan)
		cacheDurationString := paramsGetWithDefault(params, "cache", cacheDefaultDurationString)
		// only enable caching if value is set
		if cacheDurationString != "" {
			if val, err := time.ParseDuration(cacheDurationString); err == nil {
				ret.Cache = &val
			} else {
				log.Error(err.Error())
				return ret, err
			}
		}
	}

	return ret, nil
}

func (s *RequestMetricSettings) CacheDuration(requestTime time.Time) (ret *time.Duration) {
	if s.Cache != nil {
		bufferDuration := 2 * time.Second
		cachedUntilTime := requestTime.Add(*s.Cache).Add(-bufferDuration)
		cacheDuration := time.Until(cachedUntilTime)
		if cacheDuration.Seconds() > 0 {
			ret = &cacheDuration
		}
	}
	return
}

func (s *RequestMetricSettings) SetMetrics(val string) {
	s.Metrics = stringToStringList(val, ",")
}

func (s *RequestMetricSettings) SetAggregations(val string) {
	s.Aggregations = stringToStringList(val, ",")
}
