package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"
	"github.com/webdevops/go-common/azuresdk/armclient"

	"github.com/webdevops/azure-metrics-exporter/config"
)

const (
	AzureMetricApiMaxMetricNumber = 20
)

type (
	MetricProber struct {
		Conf config.Opts

		AzureClient *armclient.ArmClient

		userAgent string

		settings *RequestMetricSettings

		request  *http.Request
		response http.ResponseWriter

		ctx context.Context

		logger *log.Entry

		metricsCache struct {
			cache         *cache.Cache
			cacheKey      *string
			cacheDuration *time.Duration
		}

		serviceDiscoveryCache struct {
			cache         *cache.Cache
			cacheDuration *time.Duration
		}

		targets map[string][]MetricProbeTarget

		metricList *MetricList

		prometheus struct {
			registry *prometheus.Registry
		}

		callbackSubscriptionFishish func(subscriptionId string)

		ServiceDiscovery AzureServiceDiscovery
	}

	MetricProbeTarget struct {
		ResourceId   string
		Metrics      []string
		Aggregations []string
		Tags         map[string]string
	}
)

func NewMetricProber(ctx context.Context, logger *log.Entry, w http.ResponseWriter, settings *RequestMetricSettings, conf config.Opts) *MetricProber {
	prober := MetricProber{}
	prober.ctx = ctx
	prober.response = w
	prober.logger = logger
	prober.settings = settings
	prober.Conf = conf
	prober.ServiceDiscovery = AzureServiceDiscovery{prober: &prober}
	prober.Init()
	return &prober
}

func (p *MetricProber) Init() {
	p.targets = map[string][]MetricProbeTarget{}

	p.metricList = NewMetricList()
}
func (p *MetricProber) RegisterSubscriptionCollectFinishCallback(callback func(subscriptionId string)) {
	p.callbackSubscriptionFishish = callback
}

func (p *MetricProber) SetUserAgent(value string) {
	p.userAgent = value
}

func (p *MetricProber) SetPrometheusRegistry(registry *prometheus.Registry) {
	p.prometheus.registry = registry
}

func (p *MetricProber) SetAzureClient(client *armclient.ArmClient) {
	p.AzureClient = client
}

func (p *MetricProber) EnableMetricsCache(cache *cache.Cache, cacheKey string, cacheDuration *time.Duration) {
	p.metricsCache.cache = cache
	p.metricsCache.cacheKey = &cacheKey
	p.metricsCache.cacheDuration = cacheDuration
}

func (p *MetricProber) EnableServiceDiscoveryCache(cache *cache.Cache, cacheDuration *time.Duration) {
	p.serviceDiscoveryCache.cache = cache
	p.serviceDiscoveryCache.cacheDuration = cacheDuration
}

func (p *MetricProber) AddTarget(targets ...MetricProbeTarget) {
	for _, target := range targets {
		resourceInfo, err := azure.ParseResourceID(target.ResourceId)
		if err != nil {
			p.logger.Warnf("unable to parse resource id: %s", err.Error())
			continue
		}

		subscriptionId := resourceInfo.SubscriptionID
		if _, exists := p.targets[subscriptionId]; !exists {
			p.targets[subscriptionId] = []MetricProbeTarget{}
		}

		p.targets[subscriptionId] = append(p.targets[subscriptionId], target)
	}
}

func (p *MetricProber) FetchFromCache() bool {
	if p.metricsCache.cache == nil {
		return false
	}

	if val, ok := p.metricsCache.cache.Get(*p.metricsCache.cacheKey); ok {
		p.metricList = val.(*MetricList)
		p.publishMetricList()
		return true
	}

	return false
}

func (p *MetricProber) SaveToCache() {
	if p.metricsCache.cache == nil {
		return
	}

	if p.metricsCache.cacheDuration != nil {
		_ = p.metricsCache.cache.Add(*p.metricsCache.cacheKey, p.metricList, *p.metricsCache.cacheDuration)
		p.response.Header().Add("X-metrics-cached-until", time.Now().Add(*p.metricsCache.cacheDuration).Format(time.RFC3339))
	}
}

func (p *MetricProber) Run() {
	p.collectMetricsFromTargets()
	p.SaveToCache()
	p.publishMetricList()
}

func (p *MetricProber) collectMetricsFromTargets() {
	metricsChannel := make(chan PrometheusMetricResult)

	wgSubscription := sizedwaitgroup.New(p.Conf.Prober.ConcurrencySubscription)

	go func() {
		for subscriptionId, resourceList := range p.targets {
			wgSubscription.Add()
			go func(subscriptionId string, targetList []MetricProbeTarget) {
				defer wgSubscription.Done()

				wgSubscriptionResource := sizedwaitgroup.New(p.Conf.Prober.ConcurrencySubscriptionResource)
				client, err := p.MetricsClient(subscriptionId)
				if err != nil {
					// FIXME: find a better way to report errors
					p.logger.Error(err)
					return
				}

				for _, target := range targetList {
					wgSubscriptionResource.Add()
					go func(target MetricProbeTarget) {
						defer wgSubscriptionResource.Done()

						// request metrics in 20 metrics chunks (azure metric api limitation)
						for i := 0; i < len(target.Metrics); i += AzureMetricApiMaxMetricNumber {
							end := i + AzureMetricApiMaxMetricNumber
							if end > len(target.Metrics) {
								end = len(target.Metrics)
							}
							metricList := target.Metrics[i:end]

							if result, err := p.FetchMetricsFromTarget(client, target, metricList, target.Aggregations); err == nil {
								result.SendMetricToChannel(metricsChannel)
							} else {
								p.logger.WithField("resourceID", target.ResourceId).Warn(err)
							}
						}
					}(target)
				}
				wgSubscriptionResource.Wait()

				if p.callbackSubscriptionFishish != nil {
					p.callbackSubscriptionFishish(subscriptionId)
				}
			}(subscriptionId, resourceList)
		}
		wgSubscription.Wait()
		close(metricsChannel)
	}()

	for result := range metricsChannel {
		metric := MetricRow{
			Labels: result.Labels,
			Value:  result.Value,
		}
		p.metricList.Add(result.Name, metric)
		p.metricList.SetMetricHelp(result.Name, result.Help)
	}
}

func (p *MetricProber) publishMetricList() {
	if p.metricList == nil {
		return
	}

	// create prometheus metrics and set rows
	for _, metricName := range p.metricList.GetMetricNames() {
		gauge := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: metricName,
				Help: p.metricList.GetMetricHelp(metricName),
			},
			p.metricList.GetMetricLabelNames(metricName),
		)
		p.prometheus.registry.MustRegister(gauge)

		for _, row := range p.metricList.GetMetricList(metricName) {
			gauge.With(row.Labels).Set(row.Value)
		}
	}
}
