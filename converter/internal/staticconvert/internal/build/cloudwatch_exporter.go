package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/cloudwatch"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/cloudwatch_exporter"
)

func (b *IntegrationsV1ConfigBuilder) appendCloudwatchExporter(config *cloudwatch_exporter.Config) discovery.Exports {
	args := toCloudwatchExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "cloudwatch"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.cloudwatch.%s.targets", compLabel))
}

func toCloudwatchExporter(config *cloudwatch_exporter.Config) *cloudwatch.Arguments {
	return &cloudwatch.Arguments{
		STSRegion:             config.STSRegion,
		FIPSDisabled:          config.FIPSDisabled,
		Debug:                 config.Debug,
		DiscoveryExportedTags: config.Discovery.ExportedTags,
		Discovery:             toDiscoveryJobs(config.Discovery.Jobs),
		Static:                toStaticJobs(config.Static),
	}
}

func toDiscoveryJobs(jobs []*cloudwatch_exporter.DiscoveryJob) []cloudwatch.DiscoveryJob {
	var out []cloudwatch.DiscoveryJob
	for _, job := range jobs {
		out = append(out, toDiscoveryJob(job))
	}
	return out
}

func toDiscoveryJob(job *cloudwatch_exporter.DiscoveryJob) cloudwatch.DiscoveryJob {
	return cloudwatch.DiscoveryJob{
		Auth: cloudwatch.RegionAndRoles{
			Regions: job.InlineRegionAndRoles.Regions,
			Roles:   toRoles(job.InlineRegionAndRoles.Roles),
		},
		CustomTags:                toTags(job.CustomTags),
		SearchTags:                toTags(job.SearchTags),
		Type:                      job.Type,
		DimensionNameRequirements: job.DimensionNameRequirements,
		Metrics:                   toMetrics(job.Metrics),
		NilToZero:                 job.NilToZero,
	}
}

func toStaticJobs(jobs []cloudwatch_exporter.StaticJob) []cloudwatch.StaticJob {
	var out []cloudwatch.StaticJob
	for _, job := range jobs {
		out = append(out, toStaticJob(&job))
	}
	return out
}

func toStaticJob(job *cloudwatch_exporter.StaticJob) cloudwatch.StaticJob {
	return cloudwatch.StaticJob{
		Name: job.Name,
		Auth: cloudwatch.RegionAndRoles{
			Regions: job.Regions,
			Roles:   toRoles(job.Roles),
		},
		CustomTags: toTags(job.CustomTags),
		Namespace:  job.Namespace,
		Dimensions: toDimensions(job.Dimensions),
		Metrics:    toMetrics(job.Metrics),
		NilToZero:  job.NilToZero,
	}
}

func toRoles(roles []cloudwatch_exporter.Role) []cloudwatch.Role {
	var out []cloudwatch.Role
	for _, role := range roles {
		out = append(out, toRole(role))
	}
	return out
}

func toRole(role cloudwatch_exporter.Role) cloudwatch.Role {
	return cloudwatch.Role{
		RoleArn:    role.RoleArn,
		ExternalID: role.ExternalID,
	}
}

func toTags(tags []cloudwatch_exporter.Tag) cloudwatch.Tags {
	out := make(cloudwatch.Tags, 0)
	for _, tag := range tags {
		out[tag.Key] = tag.Value
	}
	return out
}

func toDimensions(dimensions []cloudwatch_exporter.Dimension) cloudwatch.Dimensions {
	out := make(cloudwatch.Dimensions)
	for _, dimension := range dimensions {
		out[dimension.Name] = dimension.Value
	}
	return out
}

func toMetrics(metrics []cloudwatch_exporter.Metric) []cloudwatch.Metric {
	var out []cloudwatch.Metric
	for _, metric := range metrics {
		out = append(out, toMetric(metric))
	}
	return out
}

func toMetric(metric cloudwatch_exporter.Metric) cloudwatch.Metric {
	return cloudwatch.Metric{
		Name:       metric.Name,
		Statistics: metric.Statistics,
		Period:     metric.Period,
		Length:     metric.Length,
		NilToZero:  metric.NilToZero,
	}
}
