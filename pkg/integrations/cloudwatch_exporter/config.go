package cloudwatch_exporter

import (
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
)

const (
	metricsPerQuery       = 500
	cloudWatchConcurrency = 5
	tagConcurrency        = 5
	labelsSnakeCase       = false
)

// Since we are gathering metrics from CloudWatch and writing them in prometheus during each scrape, the timestamp
// used should be the scrape one
var addCloudwatchTimestamp = false

// Avoid producing absence of values in metrics
var nilToZero = true

// ToYACEConfig converts a Config into YACE's config model. Note that the conversion is not direct, some values
// have been opinionated to simplify the config model the agent exposes for this integration.
func ToYACEConfig(c *Config) (yaceConf.ScrapeConf, error) {
	discoveryJobs := []*yaceConf.Job{}
	for _, job := range c.Discovery.Jobs {
		discoveryJobs = append(discoveryJobs, toYACEDiscoveryJob(job))
	}
	staticJobs := []*yaceConf.Static{}
	for _, stat := range c.Static {
		staticJobs = append(staticJobs, toYACEStaticJob(stat))
	}
	conf := yaceConf.ScrapeConf{
		APIVersion: "v1alpha1",
		StsRegion:  c.STSRegion,
		Discovery: yaceConf.Discovery{
			ExportedTagsOnMetrics: yaceConf.ExportedTagsOnMetrics(c.Discovery.ExportedTags),
			Jobs:                  discoveryJobs,
		},
		Static: staticJobs,
	}
	// Run the exporter's config validation. Between other things, it will check that the service for which a discovery
	// job is instantiated, it's supported.
	if err := conf.Validate(); err != nil {
		return conf, err
	}
	patchYACEDefaults(&conf)

	return conf, nil
}

// patchYACEDefaults overrides some default values YACE applies after validation.
func patchYACEDefaults(yc *yaceConf.ScrapeConf) {
	// YACE doesn't allow during validation a zero-delay in each metrics scrape. Override this behaviour since it's taken
	// into account by the rounding period.
	// https://github.com/nerdswords/yet-another-cloudwatch-exporter/blob/7e5949124bb5f26353eeff298724a5897de2a2a4/pkg/config/config.go#L320
	for _, job := range yc.Discovery.Jobs {
		for _, metric := range job.Metrics {
			metric.Delay = 0
		}
	}
}

func toYACEStaticJob(job StaticJob) *yaceConf.Static {
	return &yaceConf.Static{
		Name:       job.Name,
		Regions:    job.Regions,
		Roles:      toYACERoles(job.Roles),
		Namespace:  job.Namespace,
		CustomTags: toYACETags(job.CustomTags),
		Dimensions: toYACEDimensions(job.Dimensions),
		Metrics:    toYACEMetrics(job.Metrics),
	}
}

func toYACEDimensions(dim []Dimension) []yaceConf.Dimension {
	yaceDims := []yaceConf.Dimension{}
	for _, d := range dim {
		yaceDims = append(yaceDims, yaceConf.Dimension{
			Name:  d.Name,
			Value: d.Value,
		})
	}
	return yaceDims
}

func toYACEDiscoveryJob(job *DiscoveryJob) *yaceConf.Job {
	roles := toYACERoles(job.Roles)
	yaceJob := yaceConf.Job{
		Regions:    job.Regions,
		Roles:      roles,
		CustomTags: toYACETags(job.CustomTags),
		Type:       job.Type,
		Metrics:    toYACEMetrics(job.Metrics),
		SearchTags: toYACETags(job.SearchTags),

		// By setting RoundingPeriod to nil, the exporter will align the start and end times for retrieving CloudWatch
		// metrics, with the smallest period in the retrieved batch.
		RoundingPeriod: nil,
	}
	// Set to zero job-wide scraping time settings. This should be configured at the metric level to make the data
	// being fetched more explicit.
	yaceJob.Period = 0
	yaceJob.Length = 0
	yaceJob.Delay = 0
	yaceJob.NilToZero = &nilToZero
	yaceJob.AddCloudwatchTimestamp = &addCloudwatchTimestamp
	return &yaceJob
}

func toYACEMetrics(metrics []Metric) []*yaceConf.Metric {
	yaceMetrics := []*yaceConf.Metric{}
	for _, metric := range metrics {
		periodSeconds := int64(metric.Period.Seconds())
		lengthSeconds := periodSeconds
		yaceMetrics = append(yaceMetrics, &yaceConf.Metric{
			Name:       metric.Name,
			Statistics: metric.Statistics,

			// Length dictates the size of the window for whom we request metrics, that is, endTime - startTime. Period
			// dictates the size of the buckets in which we aggregate data, inside that window. Since data will be scraped
			// by the agent every so often, dictated by the scrapedInterval, CloudWatch should return a single datapoint
			// for each requested metric. That is if Period >= Length, but is Period > Length, we will be getting not enough
			// data to fill the whole aggregation bucket. Therefore, Period == Length.
			Period: periodSeconds,
			Length: lengthSeconds,

			// Delay moves back the time window for whom CloudWatch is requested data. Since we are already adjusting
			// this with RoundingPeriod (see toYACEDiscoveryJob), we should omit this setting.
			Delay: 0,

			NilToZero:              &nilToZero,
			AddCloudwatchTimestamp: &addCloudwatchTimestamp,
		})
	}
	return yaceMetrics
}

func toYACERoles(roles []Role) []yaceConf.Role {
	yaceRoles := []yaceConf.Role{}
	// YACE defaults to an empty role, which means the environment configured role is used
	// https://github.com/nerdswords/yet-another-cloudwatch-exporter/blob/30aeceb2324763cdd024a1311045f83a09c1df36/pkg/config/config.go#L111
	if len(roles) == 0 {
		yaceRoles = append(yaceRoles, yaceConf.Role{})
	}
	for _, role := range roles {
		yaceRoles = append(yaceRoles, yaceConf.Role{
			RoleArn:    role.RoleArn,
			ExternalID: role.ExternalID,
		})
	}
	return yaceRoles
}

func toYACETags(tags []Tag) []yaceModel.Tag {
	outTags := []yaceModel.Tag{}
	for _, t := range tags {
		outTags = append(outTags, yaceModel.Tag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	return outTags
}
