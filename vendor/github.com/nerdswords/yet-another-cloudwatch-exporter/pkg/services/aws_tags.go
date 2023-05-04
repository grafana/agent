package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice/databasemigrationserviceiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/prometheusservice/prometheusserviceiface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/aws/aws-sdk-go/service/storagegateway/storagegatewayiface"

	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/logger"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/promutil"
)

// TaggedResource is an AWS resource with tags
type TaggedResource struct {
	// ARN is the unique AWS ARN (Amazon Resource Name) of the resource
	ARN string

	// Namespace identifies the resource type (e.g. EC2)
	Namespace string

	// Region is the AWS regions that the resource belongs to
	Region string

	// Tags is a set of tags associated to the resource
	Tags []model.Tag
}

// filterThroughTags returns true if all filterTags match
// with tags of the TaggedResource, returns false otherwise.
func (r TaggedResource) FilterThroughTags(filterTags []model.Tag) bool {
	tagMatches := 0

	for _, resourceTag := range r.Tags {
		for _, filterTag := range filterTags {
			if resourceTag.Key == filterTag.Key {
				r, _ := regexp.Compile(filterTag.Value)
				if r.MatchString(resourceTag.Value) {
					tagMatches++
				}
			}
		}
	}

	return tagMatches == len(filterTags)
}

// MetricTags returns a list of tags built from the tags of
// TaggedResource, if there's a definition for its namespace
// in tagsOnMetrics.
//
// Returned tags have as key the key from tagsOnMetrics, and
// as value the value from the corresponding tag of the resource,
// if it exists (otherwise an empty string).
func (r TaggedResource) MetricTags(tagsOnMetrics config.ExportedTagsOnMetrics) []model.Tag {
	tags := make([]model.Tag, 0)
	for _, tagName := range tagsOnMetrics[r.Namespace] {
		tag := model.Tag{
			Key: tagName,
		}
		for _, resourceTag := range r.Tags {
			if resourceTag.Key == tagName {
				tag.Value = resourceTag.Value
				break
			}
		}

		// Always add the tag, even if it's empty, to ensure the same labels are present on all metrics for a single service
		tags = append(tags, tag)
	}
	return tags
}

// https://docs.aws.amazon.com/sdk-for-go/api/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface/
type TagsInterface struct {
	Client               resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	AsgClient            autoscalingiface.AutoScalingAPI
	APIGatewayClient     apigatewayiface.APIGatewayAPI
	Ec2Client            ec2iface.EC2API
	DmsClient            databasemigrationserviceiface.DatabaseMigrationServiceAPI
	PrometheusClient     prometheusserviceiface.PrometheusServiceAPI
	StoragegatewayClient storagegatewayiface.StorageGatewayAPI
	Logger               logger.Logger
}

func (iface TagsInterface) Get(ctx context.Context, job *config.Job, region string) ([]*TaggedResource, error) {
	svc := config.SupportedServices.GetService(job.Type)
	var resources []*TaggedResource

	if len(svc.ResourceFilters) > 0 {
		inputparams := &resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: svc.ResourceFilters,
			ResourcesPerPage:    aws.Int64(100), // max allowed value according to API docs
		}
		c := iface.Client
		pageNum := 0

		err := c.GetResourcesPagesWithContext(ctx, inputparams, func(page *resourcegroupstaggingapi.GetResourcesOutput, lastPage bool) bool {
			pageNum++
			promutil.ResourceGroupTaggingAPICounter.Inc()

			if len(page.ResourceTagMappingList) == 0 {
				iface.Logger.Error(errors.New("resource tag list is empty"), "Account contained no tagged resource. Tags must be defined for resources to be discovered.")
			}

			for _, resourceTagMapping := range page.ResourceTagMappingList {
				resource := TaggedResource{
					ARN:       aws.StringValue(resourceTagMapping.ResourceARN),
					Namespace: job.Type,
					Region:    region,
				}

				for _, t := range resourceTagMapping.Tags {
					resource.Tags = append(resource.Tags, model.Tag{Key: *t.Key, Value: *t.Value})
				}

				if resource.FilterThroughTags(job.SearchTags) {
					resources = append(resources, &resource)
				} else {
					iface.Logger.Debug("Skipping resource because search tags do not match", "arn", resource.ARN)
				}
			}
			return !lastPage
		})
		if err != nil {
			return nil, err
		}
	}

	if ext, ok := serviceFilters[svc.Namespace]; ok {
		if ext.ResourceFunc != nil {
			newResources, err := ext.ResourceFunc(ctx, iface, job, region)
			if err != nil {
				return nil, err
			}
			resources = append(resources, newResources...)
		}

		if ext.FilterFunc != nil {
			filteredResources, err := ext.FilterFunc(ctx, iface, resources)
			if err != nil {
				return nil, err
			}
			resources = filteredResources
		}
	}

	return resources, nil
}

func MigrateTagsToPrometheus(tagData []*TaggedResource, labelsSnakeCase bool, logger logger.Logger) []*promutil.PrometheusMetric {
	output := make([]*promutil.PrometheusMetric, 0)

	tagList := make(map[string][]string)

	for _, d := range tagData {
		for _, entry := range d.Tags {
			if !stringInSlice(entry.Key, tagList[d.Namespace]) {
				tagList[d.Namespace] = append(tagList[d.Namespace], entry.Key)
			}
		}
	}

	for _, d := range tagData {
		promNs := strings.ToLower(d.Namespace)
		if !strings.HasPrefix(promNs, "aws") {
			promNs = "aws_" + promNs
		}
		name := promutil.PromString(promNs) + "_info"
		promLabels := make(map[string]string)
		promLabels["name"] = d.ARN

		for _, entry := range tagList[d.Namespace] {
			ok, promTag := promutil.PromStringTag(entry, labelsSnakeCase)
			if !ok {
				logger.Warn("tag name is an invalid prometheus label name", "tag", entry)
				continue
			}

			labelKey := "tag_" + promTag
			promLabels[labelKey] = ""

			for _, rTag := range d.Tags {
				if entry == rTag.Key {
					promLabels[labelKey] = rTag.Value
				}
			}
		}

		var i int
		f := float64(i)

		p := promutil.PrometheusMetric{
			Name:   &name,
			Labels: promLabels,
			Value:  &f,
		}

		output = append(output, &p)
	}

	return output
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
