package cloudwatch_exporter

import (
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	yaceSvc "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
)

const (
	metricsPerQuery       = 500
	cloudWatchConcurrency = 5
	tagConcurrency        = 5
	labelsSnakeCase       = false
)

var addCloudwatchTimestamp = false
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
		ApiVersion: "v1alpha1",
		StsRegion:  c.STSRegion,
		Discovery: yaceConf.Discovery{
			ExportedTagsOnMetrics: yaceConf.ExportedTagsOnMetrics(c.Discovery.ExportedTags),
			Jobs:                  discoveryJobs,
		},
		Static: staticJobs,
	}
	if err := conf.Validate(yaceSvc.CheckServiceName); err != nil {
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
		CustomTags: toYACECustomTags(job.CustomTags),
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
	return &yaceConf.Job{
		Regions:                job.Regions,
		Roles:                  roles,
		CustomTags:             toYACECustomTags(job.CustomTags),
		Type:                   job.Type,
		NilToZero:              &nilToZero,
		Metrics:                toYACEMetrics(job.Metrics),
		Period:                 0,
		Length:                 0,
		Delay:                  0,
		RoundingPeriod:         nil,
		AddCloudwatchTimestamp: &addCloudwatchTimestamp,
	}
}

func toYACEMetrics(metrics []Metric) []*yaceConf.Metric {
	yaceMetrics := []*yaceConf.Metric{}
	for _, metric := range metrics {
		periodSeconds := int64(metric.Period.Seconds())
		lengthSeconds := periodSeconds
		yaceMetrics = append(yaceMetrics, &yaceConf.Metric{
			Name:                   metric.Name,
			Statistics:             metric.Statistics,
			Period:                 periodSeconds,
			Length:                 lengthSeconds,
			NilToZero:              &nilToZero,
			AddCloudwatchTimestamp: &addCloudwatchTimestamp,
			Delay:                  0,
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

func toYACECustomTags(tags []Tag) []yaceModel.Tag {
	outTags := []yaceModel.Tag{}
	for _, t := range tags {
		outTags = append(outTags, yaceModel.Tag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	return outTags
}
