package metrics

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/azuresdk/armclient"
	stringsCommon "github.com/webdevops/go-common/strings"
	"github.com/webdevops/go-common/utils/to"
)

var (
	metricNamePlaceholders     = regexp.MustCompile(`{([^}]+)}`)
	metricNameNotAllowedChars  = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	metricLabelNotAllowedChars = regexp.MustCompile(`[^a-zA-Z0-9_]`)
	metricNameReplacer         = strings.NewReplacer("-", "_", " ", "_", "/", "_", ".", "_")
)

type (
	AzureInsightMetrics struct {
	}

	AzureInsightMetricsResult struct {
		settings *RequestMetricSettings
		target   *MetricProbeTarget
		Result   *armmonitor.MetricsClientListResponse
	}

	PrometheusMetricResult struct {
		Name   string
		Labels prometheus.Labels
		Value  float64
		Help   string
	}
)

func (p *MetricProber) MetricsClient(subscriptionId string) (*armmonitor.MetricsClient, error) {
	clientOpts := p.AzureClient.NewArmClientOptions()
	clientOpts.PerRetryPolicies = append(
		clientOpts.PerRetryPolicies,
		noCachePolicy{},
	)
	return armmonitor.NewMetricsClient(p.AzureClient.GetCred(), clientOpts)
}

func (p *MetricProber) FetchMetricsFromTarget(client *armmonitor.MetricsClient, target MetricProbeTarget, metrics, aggregations []string) (AzureInsightMetricsResult, error) {
	ret := AzureInsightMetricsResult{
		settings: p.settings,
		target:   &target,
	}

	resultType := armmonitor.ResultTypeData
	opts := armmonitor.MetricsClientListOptions{
		Interval:    p.settings.Interval,
		ResultType:  &resultType,
		Timespan:    to.StringPtr(p.settings.Timespan),
		Metricnames: to.StringPtr(strings.Join(metrics, ",")),
		Top:         p.settings.MetricTop,
	}

	if len(aggregations) >= 1 {
		opts.Aggregation = to.StringPtr(strings.Join(aggregations, ","))
	}

	if len(p.settings.MetricFilter) >= 1 {
		opts.Filter = to.StringPtr(p.settings.MetricFilter)
	}

	if len(p.settings.MetricNamespace) >= 1 {
		opts.Metricnamespace = to.StringPtr(p.settings.MetricNamespace)
	}

	if len(p.settings.MetricOrderBy) >= 1 {
		opts.Orderby = to.StringPtr(p.settings.MetricOrderBy)
	}

	resourceURI := target.ResourceId
	if strings.HasPrefix(strings.ToLower(p.settings.MetricNamespace), "microsoft.storage/storageaccounts/") {
		splitNamespace := strings.Split(p.settings.MetricNamespace, "/")
		// Storage accounts have an extra requirement that their ResourceURI include <type>/default
		storageAccountType := splitNamespace[len(splitNamespace)-1]
		resourceURI = resourceURI + fmt.Sprintf("/%s/default", storageAccountType)
	}

	result, err := client.List(
		p.ctx,
		resourceURI,
		&opts,
	)

	if err == nil {
		ret.Result = &result
	}

	return ret, err
}

func (r *AzureInsightMetricsResult) buildMetric(labels prometheus.Labels, value float64) (metric PrometheusMetricResult) {
	// copy map to ensure we don't keep references
	metricLabels := prometheus.Labels{}
	for labelName, labelValue := range labels {
		metricLabels[labelName] = labelValue
	}

	metric = PrometheusMetricResult{
		Name:   r.settings.MetricTemplate,
		Labels: metricLabels,
		Value:  value,
	}

	// fallback if template is empty (should not be)
	if r.settings.MetricTemplate == "" {
		metric.Name = r.settings.Name
	}

	resourceType := r.settings.ResourceType
	// MetricNamespace is more descriptive than type
	if r.settings.MetricNamespace != "" {
		resourceType = r.settings.MetricNamespace
	}

	// set help
	metric.Help = r.settings.HelpTemplate
	if metricNamePlaceholders.MatchString(metric.Help) {
		metric.Help = metricNamePlaceholders.ReplaceAllStringFunc(
			metric.Help,
			func(fieldName string) string {
				fieldName = strings.Trim(fieldName, "{}")
				switch fieldName {
				case "name":
					return r.settings.Name
				case "type":
					return resourceType
				default:
					if fieldValue, exists := metric.Labels[fieldName]; exists {
						return fieldValue
					}
				}
				return ""
			},
		)
	}

	if metricNamePlaceholders.MatchString(metric.Name) {
		metric.Name = metricNamePlaceholders.ReplaceAllStringFunc(
			metric.Name,
			func(fieldName string) string {
				fieldName = strings.Trim(fieldName, "{}")
				switch fieldName {
				case "name":
					return r.settings.Name
				case "type":
					return resourceType
				default:
					if fieldValue, exists := metric.Labels[fieldName]; exists {
						// remove label, when we add it to metric name
						delete(metric.Labels, fieldName)
						return fieldValue
					}
				}
				return ""
			},
		)
	}

	// sanitize metric name
	metric.Name = metricNameReplacer.Replace(metric.Name)
	metric.Name = strings.ToLower(metric.Name)
	metric.Name = metricNameNotAllowedChars.ReplaceAllString(metric.Name, "")

	return
}

func (r *AzureInsightMetricsResult) SendMetricToChannel(channel chan<- PrometheusMetricResult) {
	if r.Result.Value != nil {
		// DEBUGGING
		// data, _ := json.Marshal(r.Result)
		// fmt.Println(string(data))

		for _, metric := range r.Result.Value {
			if metric.Timeseries != nil {
				for _, timeseries := range metric.Timeseries {
					if timeseries.Data != nil {
						// get dimension name (optional)
						dimensions := map[string]string{}
						if timeseries.Metadatavalues != nil {
							for _, dimensionRow := range timeseries.Metadatavalues {
								dimensions[to.String(dimensionRow.Name.Value)] = to.String(dimensionRow.Value)
							}
						}

						resourceId := r.target.ResourceId
						azureResource, _ := armclient.ParseResourceId(resourceId)

						metricUnit := ""
						if metric.Unit != nil {
							metricUnit = string(*metric.Unit)
						}

						metricLabels := prometheus.Labels{
							"resourceID":     strings.ToLower(resourceId),
							"subscriptionID": azureResource.Subscription,
							"resourceGroup":  azureResource.ResourceGroup,
							"resourceName":   azureResource.ResourceName,
							"metric":         to.String(metric.Name.Value),
							"unit":           metricUnit,
							"interval":       to.String(r.settings.Interval),
							"timespan":       r.settings.Timespan,
							"aggregation":    "",
						}

						// add resource tags as labels
						metricLabels = armclient.AddResourceTagsToPrometheusLabels(
							metricLabels,
							r.target.Tags,
							r.settings.TagLabels,
						)

						if len(dimensions) == 1 {
							// we have only one dimension
							// add one dimension="foobar" label (backward compatibility)
							for _, dimensionValue := range dimensions {
								metricLabels["dimension"] = dimensionValue
							}
						} else if len(dimensions) >= 2 {
							// we have multiple dimensions
							// add each dimension as dimensionXzy="foobar" label
							for dimensionName, dimensionValue := range dimensions {
								labelName := "dimension" + stringsCommon.UppercaseFirst(dimensionName)
								labelName = metricLabelNotAllowedChars.ReplaceAllString(labelName, "")
								metricLabels[labelName] = dimensionValue
							}
						}

						for _, timeseriesData := range timeseries.Data {
							if timeseriesData.Total != nil {
								metricLabels["aggregation"] = "total"
								channel <- r.buildMetric(
									metricLabels,
									*timeseriesData.Total,
								)
							}

							if timeseriesData.Minimum != nil {
								metricLabels["aggregation"] = "minimum"
								channel <- r.buildMetric(
									metricLabels,
									*timeseriesData.Minimum,
								)
							}

							if timeseriesData.Maximum != nil {
								metricLabels["aggregation"] = "maximum"
								channel <- r.buildMetric(
									metricLabels,
									*timeseriesData.Maximum,
								)
							}

							if timeseriesData.Average != nil {
								metricLabels["aggregation"] = "average"
								channel <- r.buildMetric(
									metricLabels,
									*timeseriesData.Average,
								)
							}

							if timeseriesData.Count != nil {
								metricLabels["aggregation"] = "count"
								channel <- r.buildMetric(
									metricLabels,
									*timeseriesData.Count,
								)
							}
						}
					}
				}
			}
		}
	}
}
