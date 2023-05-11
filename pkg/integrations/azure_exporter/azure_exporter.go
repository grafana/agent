package azure_exporter

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	azure_config "github.com/webdevops/azure-metrics-exporter/config"
	"github.com/webdevops/azure-metrics-exporter/metrics"
	"github.com/webdevops/go-common/azuresdk/armclient"

	"github.com/grafana/agent/pkg/integrations/config"
)

type Exporter struct {
	cfg               Config
	logger            *zap.SugaredLogger // used by azure client
	ConcurrencyConfig azure_config.Opts
}

func (e Exporter) MetricsHandler() (http.Handler, error) {
	//Safe to re-use as it doesn't connect to anything directly
	client, err := armclient.NewArmClientWithCloudName(e.cfg.AzureCloudEnvironment, e.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure client, %v", err)
	}

	h := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		reg := prometheus.NewRegistry()
		ctx := context.Background()

		params := req.URL.Query()
		mergedConfig := MergeConfigWithQueryParams(e.cfg, params)

		if err := mergedConfig.Validate(); err != nil {
			err = fmt.Errorf("config to be used for scraping was invalid, %v", err)
			e.logger.Error(err)
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		tagManager, err := client.TagManager.ParseTagConfig(mergedConfig.IncludedResourceTags)
		if err != nil {
			err = fmt.Errorf("unable to create azure tag manager from included_resource_tags %s, %v", strings.Join(mergedConfig.IncludedResourceTags, ","), err)
			e.logger.Error(err)
			http.Error(resp, err.Error(), http.StatusBadRequest)
		}

		logEntry := e.logger.With(
			"resource_type", mergedConfig.ResourceType,
			"resource_graph_query_filter", mergedConfig.ResourceGraphQueryFilter,
			"subscriptions", strings.Join(mergedConfig.Subscriptions, ","),
			"metric_namespace", mergedConfig.MetricNamespace,
			"metrics", strings.Join(mergedConfig.Metrics, ","),
		)

		settings, err := mergedConfig.ToScrapeSettings()
		if err != nil {
			e.logger.Error(fmt.Errorf("unexpected error mapping config to scrape settings, %v", err))
			http.Error(resp, "unexpected scrape error", http.StatusInternalServerError)
			return
		}

		prober := metrics.NewMetricProber(ctx, logEntry, nil, settings, e.ConcurrencyConfig)
		prober.SetAzureClient(client)
		prober.SetPrometheusRegistry(reg)
		prober.SetAzureResourceTagManager(tagManager)

		err = prober.ServiceDiscovery.FindResourceGraph(ctx, settings.Subscriptions, settings.ResourceType, settings.Filter)
		if err != nil {
			e.logger.Error(fmt.Errorf("service discovery failed, %v", err))
			http.Error(resp, "Failed to discovery azure resources", http.StatusInternalServerError)
			return
		}

		prober.Run()

		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(resp, req)
	})
	return h, nil
}

func (e Exporter) ScrapeConfigs() []config.ScrapeConfig {
	return []config.ScrapeConfig{{JobName: e.cfg.Name(), MetricsPath: "/metrics"}}
}

func (e Exporter) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
