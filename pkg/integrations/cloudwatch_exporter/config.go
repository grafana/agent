package cloudwatch_exporter

import (
	yaceConf "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	yaceModel "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	yaceSvc "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
	"time"
)

const (
	metricsPerQuery       = 500
	cloudWatchConcurrency = 5
	tagConcurrency        = 5
	labelsSnakeCase       = false
)

var addCloudwatchTimestamp = false
var nilToZero = true

func ToYACEConfig(c *Config) (yaceConf.ScrapeConf, error) {
	discoveryJobs := []*yaceConf.Job{}
	for _, job := range c.Discovery.Jobs {
		discoveryJobs = append(discoveryJobs, toYACEDiscoveryJob(job))
	}
	staticJobs := []*yaceConf.Static{}
	for _, stat := range c.Static {
		staticJobs = append(staticJobs, toYACEStaticJob(stat))
	}
	customNSJobs := []*yaceConf.CustomNamespace{}
	for _, ns := range c.CustomNamespace {
		customNSJobs = append(customNSJobs, toYACECustomNSJob(ns))
	}
	conf := yaceConf.ScrapeConf{
		StsRegion: c.STSRegion,
		Discovery: yaceConf.Discovery{
			ExportedTagsOnMetrics: nil,
			Jobs:                  discoveryJobs,
		},
		Static:          staticJobs,
		CustomNamespace: customNSJobs,
	}
	return conf, conf.Validate(yaceSvc.CheckServiceName)
}

func toYACECustomNSJob(job CustomNamespaceJob) *yaceConf.CustomNamespace {
	length, period, roundingPeriod := toYACETimeParameters(job.Period)
	return &yaceConf.CustomNamespace{
		Regions:                   job.Regions,
		Name:                      job.Name,
		Namespace:                 job.Namespace,
		Roles:                     toYACERoles(job.Roles),
		Metrics:                   toYACEMetrics(job.Metrics),
		Statistics:                nil,
		NilToZero:                 &nilToZero,
		Period:                    period,
		Length:                    length,
		Delay:                     0,
		AddCloudwatchTimestamp:    &addCloudwatchTimestamp,
		CustomTags:                toYACECustomTags(job.CustomTags),
		DimensionNameRequirements: nil,
		RoundingPeriod:            &roundingPeriod,
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

func toYACETimeParameters(period time.Duration) (length, yacePeriod, roundingPeriod int64) {
	yacePeriod = int64(period.Seconds())
	length = yacePeriod
	roundingPeriod = yacePeriod
	return
}

func toYACEDiscoveryJob(job *DiscoveryJob) *yaceConf.Job {
	roles := toYACERoles(job.Roles)
	periodSeconds := int64(job.Period.Seconds())
	lengthSeconds := periodSeconds
	roundingPeriod := periodSeconds
	return &yaceConf.Job{
		Regions:                job.Regions,
		Roles:                  roles,
		CustomTags:             toYACECustomTags(job.CustomTags),
		Type:                   job.Type,
		NilToZero:              &nilToZero,
		Metrics:                toYACEMetrics(job.Metrics),
		Period:                 periodSeconds,
		Length:                 lengthSeconds,
		Delay:                  0,
		RoundingPeriod:         &roundingPeriod,
		AddCloudwatchTimestamp: &addCloudwatchTimestamp,
	}
}

func toYACEMetrics(metrics []Metric) []*yaceConf.Metric {
	yaceMetrics := []*yaceConf.Metric{}
	for _, metric := range metrics {
		yaceMetrics = append(yaceMetrics, &yaceConf.Metric{
			Name:       metric.Name,
			Statistics: metric.Statistics,
		})
	}
	return yaceMetrics
}

func toYACERoles(roles []Role) []yaceConf.Role {
	yaceRoles := []yaceConf.Role{}
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
