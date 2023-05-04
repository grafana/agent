package job

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/session"
)

func ScrapeAwsData(
	ctx context.Context,
	cfg config.ScrapeConf,
	metricsPerQuery int,
	cloudwatchSemaphore,
	tagSemaphore chan struct{},
	cache session.SessionCache,
	logger logger.Logger,
) ([]*services.TaggedResource, []*cloudwatchData) { //nolint:revive
	mux := &sync.Mutex{}

	cwData := make([]*cloudwatchData, 0)
	awsInfoData := make([]*services.TaggedResource, 0)
	var wg sync.WaitGroup

	// since we have called refresh, we have loaded all the credentials
	// into the clients and it is now safe to call concurrently. Defer the
	// clearing, so we always clear credentials before the next scrape
	cache.Refresh()
	defer cache.Clear()

	for _, discoveryJob := range cfg.Discovery.Jobs {
		for _, role := range discoveryJob.Roles {
			for _, region := range discoveryJob.Regions {
				wg.Add(1)
				go func(discoveryJob *config.Job, region string, role config.Role) {
					defer wg.Done()
					jobLogger := logger.With("job_type", discoveryJob.Type, "region", region, "arn", role.RoleArn)
					result, err := cache.GetSTS(role).GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
					if err != nil || result.Account == nil {
						jobLogger.Error(err, "Couldn't get account Id")
						return
					}
					jobLogger = jobLogger.With("account", *result.Account)

					clientCloudwatch := cloudwatchInterface{
						client: cache.GetCloudwatch(&region, role),
						logger: jobLogger,
					}

					clientTag := services.TagsInterface{
						Client:               cache.GetTagging(&region, role),
						APIGatewayClient:     cache.GetAPIGateway(&region, role),
						AsgClient:            cache.GetASG(&region, role),
						DmsClient:            cache.GetDMS(&region, role),
						Ec2Client:            cache.GetEC2(&region, role),
						StoragegatewayClient: cache.GetStorageGateway(&region, role),
						PrometheusClient:     cache.GetPrometheus(&region, role),
						Logger:               jobLogger,
					}

					resources, metrics := scrapeDiscoveryJobUsingMetricData(ctx, discoveryJob, region, result.Account, cfg.Discovery.ExportedTagsOnMetrics, clientTag, clientCloudwatch, metricsPerQuery, discoveryJob.RoundingPeriod, tagSemaphore, jobLogger)
					if len(resources) != 0 && len(metrics) != 0 {
						mux.Lock()
						awsInfoData = append(awsInfoData, resources...)
						cwData = append(cwData, metrics...)
						mux.Unlock()
					}
				}(discoveryJob, region, role)
			}
		}
	}

	for _, staticJob := range cfg.Static {
		for _, role := range staticJob.Roles {
			for _, region := range staticJob.Regions {
				wg.Add(1)
				go func(staticJob *config.Static, region string, role config.Role) {
					defer wg.Done()
					jobLogger := logger.With("static_job_name", staticJob.Name, "region", region, "arn", role.RoleArn)
					result, err := cache.GetSTS(role).GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
					if err != nil || result.Account == nil {
						jobLogger.Error(err, "Couldn't get account Id")
						return
					}
					jobLogger = jobLogger.With("account", *result.Account)

					clientCloudwatch := cloudwatchInterface{
						client: cache.GetCloudwatch(&region, role),
						logger: jobLogger,
					}

					metrics := scrapeStaticJob(ctx, staticJob, region, result.Account, clientCloudwatch, cloudwatchSemaphore, jobLogger)

					mux.Lock()
					cwData = append(cwData, metrics...)
					mux.Unlock()
				}(staticJob, region, role)
			}
		}
	}

	for _, customNamespaceJob := range cfg.CustomNamespace {
		for _, role := range customNamespaceJob.Roles {
			for _, region := range customNamespaceJob.Regions {
				wg.Add(1)
				go func(customNamespaceJob *config.CustomNamespace, region string, role config.Role) {
					defer wg.Done()
					jobLogger := logger.With("custom_metric_namespace", customNamespaceJob.Namespace, "region", region, "arn", role.RoleArn)
					result, err := cache.GetSTS(role).GetCallerIdentityWithContext(ctx, &sts.GetCallerIdentityInput{})
					if err != nil || result.Account == nil {
						jobLogger.Error(err, "Couldn't get account Id")
						return
					}
					jobLogger = jobLogger.With("account", *result.Account)

					clientCloudwatch := cloudwatchInterface{
						client: cache.GetCloudwatch(&region, role),
						logger: jobLogger,
					}

					metrics := scrapeCustomNamespaceJobUsingMetricData(
						ctx,
						customNamespaceJob,
						region,
						result.Account,
						clientCloudwatch,
						cloudwatchSemaphore,
						tagSemaphore,
						jobLogger,
						metricsPerQuery,
					)

					mux.Lock()
					cwData = append(cwData, metrics...)
					mux.Unlock()
				}(customNamespaceJob, region, role)
			}
		}
	}
	wg.Wait()
	return awsInfoData, cwData
}

func scrapeStaticJob(ctx context.Context, resource *config.Static, region string, accountID *string, clientCloudwatch cloudwatchInterface, cloudwatchSemaphore chan struct{}, logger logger.Logger) (cw []*cloudwatchData) {
	mux := &sync.Mutex{}
	var wg sync.WaitGroup

	for j := range resource.Metrics {
		metric := resource.Metrics[j]
		wg.Add(1)
		go func() {
			defer wg.Done()

			cloudwatchSemaphore <- struct{}{}
			defer func() {
				<-cloudwatchSemaphore
			}()

			id := resource.Name
			data := cloudwatchData{
				ID:                     &id,
				Metric:                 &metric.Name,
				Namespace:              &resource.Namespace,
				Statistics:             metric.Statistics,
				NilToZero:              metric.NilToZero,
				AddCloudwatchTimestamp: metric.AddCloudwatchTimestamp,
				CustomTags:             resource.CustomTags,
				Dimensions:             createStaticDimensions(resource.Dimensions),
				Region:                 &region,
				AccountID:              accountID,
			}

			filter := createGetMetricStatisticsInput(
				data.Dimensions,
				&resource.Namespace,
				metric,
				logger,
			)

			data.Points = clientCloudwatch.get(ctx, filter)

			if data.Points != nil {
				mux.Lock()
				cw = append(cw, &data)
				mux.Unlock()
			}
		}()
	}
	wg.Wait()
	return cw
}

func getMetricDataInputLength(metrics []*config.Metric) int64 {
	var length int64
	for _, metric := range metrics {
		if metric.Length > length {
			length = metric.Length
		}
	}
	return length
}

func getMetricDataForQueries(
	ctx context.Context,
	discoveryJob *config.Job,
	svc *config.ServiceConfig,
	region string,
	accountID *string,
	tagsOnMetrics config.ExportedTagsOnMetrics,
	clientCloudwatch cloudwatchInterface,
	resources []*services.TaggedResource,
	tagSemaphore chan struct{},
	logger logger.Logger,
) []cloudwatchData {
	var getMetricDatas []cloudwatchData

	// For every metric of the job
	for _, metric := range discoveryJob.Metrics {
		// Get the full list of metrics
		// This includes, for this metric the possible combinations
		// of dimensions and value of dimensions with data
		tagSemaphore <- struct{}{}

		metricsList, err := getFullMetricsList(ctx, svc.Namespace, metric, clientCloudwatch)
		<-tagSemaphore

		if err != nil {
			logger.Error(err, "Failed to get full metric list", "metric_name", metric.Name, "namespace", svc.Namespace)
			continue
		}

		if len(resources) == 0 {
			logger.Debug("No resources for metric", "metric_name", metric.Name, "namespace", svc.Namespace)
		}
		getMetricDatas = append(getMetricDatas, getFilteredMetricDatas(region, accountID, discoveryJob.Type, discoveryJob.CustomTags, tagsOnMetrics, svc.DimensionRegexps, resources, metricsList.Metrics, discoveryJob.DimensionNameRequirements, metric)...)
	}
	return getMetricDatas
}

func scrapeDiscoveryJobUsingMetricData(
	ctx context.Context,
	job *config.Job,
	region string,
	accountID *string,
	tagsOnMetrics config.ExportedTagsOnMetrics,
	clientTag services.TagsInterface,
	clientCloudwatch cloudwatchInterface,
	metricsPerQuery int,
	roundingPeriod *int64,
	tagSemaphore chan struct{},
	logger logger.Logger,
) (resources []*services.TaggedResource, cw []*cloudwatchData) {
	// Add the info tags of all the resources
	tagSemaphore <- struct{}{}
	resources, err := clientTag.Get(ctx, job, region)
	<-tagSemaphore
	if err != nil {
		logger.Error(err, "Couldn't describe resources")
		return
	}

	if len(resources) == 0 {
		logger.Info("No tagged resources made it through filtering")
		return
	}

	svc := config.SupportedServices.GetService(job.Type)
	getMetricDatas := getMetricDataForQueries(ctx, job, svc, region, accountID, tagsOnMetrics, clientCloudwatch, resources, tagSemaphore, logger)
	metricDataLength := len(getMetricDatas)
	if metricDataLength == 0 {
		logger.Debug("No metrics data found")
		return
	}

	maxMetricCount := metricsPerQuery
	length := getMetricDataInputLength(job.Metrics)
	partition := int(math.Ceil(float64(metricDataLength) / float64(maxMetricCount)))

	mux := &sync.Mutex{}
	var wg sync.WaitGroup
	wg.Add(partition)

	for i := 0; i < metricDataLength; i += maxMetricCount {
		go func(i int) {
			defer wg.Done()
			end := i + maxMetricCount
			if end > metricDataLength {
				end = metricDataLength
			}
			input := getMetricDatas[i:end]
			filter := createGetMetricDataInput(input, &svc.Namespace, length, job.Delay, roundingPeriod, logger)
			data := clientCloudwatch.getMetricData(ctx, filter)
			if data != nil {
				output := make([]*cloudwatchData, 0)
				for _, MetricDataResult := range data.MetricDataResults {
					getMetricData, err := findGetMetricDataByID(input, *MetricDataResult.Id)
					if err == nil {
						if len(MetricDataResult.Values) != 0 {
							getMetricData.GetMetricDataPoint = MetricDataResult.Values[0]
							getMetricData.GetMetricDataTimestamps = MetricDataResult.Timestamps[0]
						}
						output = append(output, &getMetricData)
					}
				}
				mux.Lock()
				cw = append(cw, output...)
				mux.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return resources, cw
}

func scrapeCustomNamespaceJobUsingMetricData(
	ctx context.Context,
	customNamespaceJob *config.CustomNamespace,
	region string,
	accountID *string,
	clientCloudwatch cloudwatchInterface,
	cloudwatchSemaphore chan struct{},
	tagSemaphore chan struct{},
	logger logger.Logger,
	metricsPerQuery int,
) (cw []*cloudwatchData) {
	mux := &sync.Mutex{}
	var wg sync.WaitGroup

	getMetricDatas := getMetricDataForQueriesForCustomNamespace(ctx, customNamespaceJob, region, accountID, clientCloudwatch, tagSemaphore, logger)
	metricDataLength := len(getMetricDatas)
	if metricDataLength == 0 {
		logger.Debug("No metrics data found")
		return
	}

	maxMetricCount := metricsPerQuery
	length := getMetricDataInputLength(customNamespaceJob.Metrics)
	partition := int(math.Ceil(float64(metricDataLength) / float64(maxMetricCount)))

	wg.Add(partition)

	for i := 0; i < metricDataLength; i += maxMetricCount {
		go func(i int) {
			cloudwatchSemaphore <- struct{}{}

			defer func() {
				defer wg.Done()
				<-cloudwatchSemaphore
			}()

			end := i + maxMetricCount
			if end > metricDataLength {
				end = metricDataLength
			}
			input := getMetricDatas[i:end]
			filter := createGetMetricDataInput(input, &customNamespaceJob.Namespace, length, customNamespaceJob.Delay, customNamespaceJob.RoundingPeriod, logger)
			data := clientCloudwatch.getMetricData(ctx, filter)
			if data != nil {
				output := make([]*cloudwatchData, 0)
				for _, MetricDataResult := range data.MetricDataResults {
					getMetricData, err := findGetMetricDataByID(input, *MetricDataResult.Id)
					if err == nil {
						if len(MetricDataResult.Values) != 0 {
							getMetricData.GetMetricDataPoint = MetricDataResult.Values[0]
							getMetricData.GetMetricDataTimestamps = MetricDataResult.Timestamps[0]
						}
						output = append(output, &getMetricData)
					}
				}
				mux.Lock()
				cw = append(cw, output...)
				mux.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return cw
}

func getMetricDataForQueriesForCustomNamespace(
	ctx context.Context,
	customNamespaceJob *config.CustomNamespace,
	region string,
	accountID *string,
	clientCloudwatch cloudwatchInterface,
	tagSemaphore chan struct{},
	logger logger.Logger,
) []cloudwatchData {
	var getMetricDatas []cloudwatchData

	// For every metric of the job
	for _, metric := range customNamespaceJob.Metrics {
		// Get the full list of metrics
		// This includes, for this metric the possible combinations
		// of dimensions and value of dimensions with data
		tagSemaphore <- struct{}{}

		metricsList, err := getFullMetricsList(ctx, customNamespaceJob.Namespace, metric, clientCloudwatch)
		<-tagSemaphore

		if err != nil {
			logger.Error(err, "Failed to get full metric list", "metric_name", metric.Name, "namespace", customNamespaceJob.Namespace)
			continue
		}

		for _, cwMetric := range metricsList.Metrics {
			if len(customNamespaceJob.DimensionNameRequirements) > 0 && !metricDimensionsMatchNames(cwMetric, customNamespaceJob.DimensionNameRequirements) {
				continue
			}

			for _, stats := range metric.Statistics {
				id := fmt.Sprintf("id_%d", rand.Int())
				getMetricDatas = append(getMetricDatas, cloudwatchData{
					ID:                     &customNamespaceJob.Name,
					MetricID:               &id,
					Metric:                 &metric.Name,
					Namespace:              &customNamespaceJob.Namespace,
					Statistics:             []string{stats},
					NilToZero:              metric.NilToZero,
					AddCloudwatchTimestamp: metric.AddCloudwatchTimestamp,
					CustomTags:             customNamespaceJob.CustomTags,
					Dimensions:             cwMetric.Dimensions,
					Region:                 &region,
					AccountID:              accountID,
					Period:                 metric.Period,
				})
			}
		}
	}
	return getMetricDatas
}
