package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/aws/aws-sdk-go/service/storagegateway"

	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/config"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/model"
	"github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/promutil"
)

type serviceFilter struct {
	// ResourceFunc can be used to fetch additional resources
	ResourceFunc func(context.Context, TagsInterface, *config.Job, string) ([]*TaggedResource, error)

	// FilterFunc can be used to the input resources or to drop based on some condition
	FilterFunc func(context.Context, TagsInterface, []*TaggedResource) ([]*TaggedResource, error)
}

// serviceFilters maps a service namespace to (optional) serviceFilter
var serviceFilters = map[string]serviceFilter{
	"AWS/ApiGateway": {
		FilterFunc: func(ctx context.Context, iface TagsInterface, inputResources []*TaggedResource) (outputResources []*TaggedResource, err error) {
			promutil.APIGatewayAPICounter.Inc()
			var limit int64 = 500 // max number of results per page. default=25, max=500
			const maxPages = 10
			input := apigateway.GetRestApisInput{Limit: &limit}
			output := apigateway.GetRestApisOutput{}
			var pageNum int
			err = iface.APIGatewayClient.GetRestApisPagesWithContext(ctx, &input, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
				pageNum++
				output.Items = append(output.Items, page.Items...)
				return pageNum <= maxPages
			})
			for _, resource := range inputResources {
				for i, gw := range output.Items {
					searchString := regexp.MustCompile(fmt.Sprintf(".*apis/%s$", *gw.Id))
					if searchString.MatchString(resource.ARN) {
						r := resource
						r.ARN = strings.ReplaceAll(resource.ARN, *gw.Id, *gw.Name)
						outputResources = append(outputResources, r)
						output.Items = append(output.Items[:i], output.Items[i+1:]...)
						break
					}
				}
			}
			return outputResources, err
		},
	},
	"AWS/AutoScaling": {
		ResourceFunc: func(ctx context.Context, iface TagsInterface, job *config.Job, region string) (resources []*TaggedResource, err error) {
			pageNum := 0
			return resources, iface.AsgClient.DescribeAutoScalingGroupsPagesWithContext(ctx, &autoscaling.DescribeAutoScalingGroupsInput{},
				func(page *autoscaling.DescribeAutoScalingGroupsOutput, more bool) bool {
					pageNum++
					promutil.AutoScalingAPICounter.Inc()

					for _, asg := range page.AutoScalingGroups {
						resource := TaggedResource{
							ARN:       aws.StringValue(asg.AutoScalingGroupARN),
							Namespace: job.Type,
							Region:    region,
						}

						for _, t := range asg.Tags {
							resource.Tags = append(resource.Tags, model.Tag{Key: *t.Key, Value: *t.Value})
						}

						if resource.FilterThroughTags(job.SearchTags) {
							resources = append(resources, &resource)
						}
					}
					return pageNum < 100
				},
			)
		},
	},
	"AWS/DMS": {
		// Append the replication instance identifier to DMS task and instance ARNs
		FilterFunc: func(ctx context.Context, iface TagsInterface, inputResources []*TaggedResource) (outputResources []*TaggedResource, err error) {
			if len(inputResources) == 0 {
				return inputResources, nil
			}

			replicationInstanceIdentifiers := make(map[string]string)
			pageNum := 0
			if err := iface.DmsClient.DescribeReplicationInstancesPagesWithContext(ctx, nil,
				func(page *databasemigrationservice.DescribeReplicationInstancesOutput, lastPage bool) bool {
					pageNum++
					promutil.DmsAPICounter.Inc()

					for _, instance := range page.ReplicationInstances {
						replicationInstanceIdentifiers[aws.StringValue(instance.ReplicationInstanceArn)] = aws.StringValue(instance.ReplicationInstanceIdentifier)
					}

					return pageNum < 100
				},
			); err != nil {
				return nil, err
			}
			pageNum = 0
			if err := iface.DmsClient.DescribeReplicationTasksPagesWithContext(ctx, nil,
				func(page *databasemigrationservice.DescribeReplicationTasksOutput, lastPage bool) bool {
					pageNum++
					promutil.DmsAPICounter.Inc()

					for _, task := range page.ReplicationTasks {
						taskInstanceArn := aws.StringValue(task.ReplicationInstanceArn)
						if instanceIdentifier, ok := replicationInstanceIdentifiers[taskInstanceArn]; ok {
							replicationInstanceIdentifiers[aws.StringValue(task.ReplicationTaskArn)] = instanceIdentifier
						}
					}

					return pageNum < 100
				},
			); err != nil {
				return nil, err
			}

			for _, resource := range inputResources {
				r := resource
				// Append the replication instance identifier to replication instance and task ARNs
				if instanceIdentifier, ok := replicationInstanceIdentifiers[r.ARN]; ok {
					r.ARN = fmt.Sprintf("%s/%s", r.ARN, instanceIdentifier)
				}
				outputResources = append(outputResources, r)
			}
			return
		},
	},
	"AWS/EC2Spot": {
		ResourceFunc: func(ctx context.Context, iface TagsInterface, job *config.Job, region string) (resources []*TaggedResource, err error) {
			pageNum := 0
			return resources, iface.Ec2Client.DescribeSpotFleetRequestsPagesWithContext(ctx, &ec2.DescribeSpotFleetRequestsInput{},
				func(page *ec2.DescribeSpotFleetRequestsOutput, more bool) bool {
					pageNum++
					promutil.Ec2APICounter.Inc()

					for _, ec2Spot := range page.SpotFleetRequestConfigs {
						resource := TaggedResource{
							ARN:       aws.StringValue(ec2Spot.SpotFleetRequestId),
							Namespace: job.Type,
							Region:    region,
						}

						for _, t := range ec2Spot.Tags {
							resource.Tags = append(resource.Tags, model.Tag{Key: *t.Key, Value: *t.Value})
						}

						if resource.FilterThroughTags(job.SearchTags) {
							resources = append(resources, &resource)
						}
					}
					return pageNum < 100
				},
			)
		},
	},
	"AWS/Prometheus": {
		ResourceFunc: func(ctx context.Context, iface TagsInterface, job *config.Job, region string) (resources []*TaggedResource, err error) {
			pageNum := 0
			return resources, iface.PrometheusClient.ListWorkspacesPagesWithContext(ctx, &prometheusservice.ListWorkspacesInput{},
				func(page *prometheusservice.ListWorkspacesOutput, more bool) bool {
					pageNum++
					promutil.ManagedPrometheusAPICounter.Inc()

					for _, ws := range page.Workspaces {
						resource := TaggedResource{
							ARN:       aws.StringValue(ws.Arn),
							Namespace: job.Type,
							Region:    region,
						}

						for key, value := range ws.Tags {
							resource.Tags = append(resource.Tags, model.Tag{Key: key, Value: *value})
						}

						if resource.FilterThroughTags(job.SearchTags) {
							resources = append(resources, &resource)
						}
					}
					return pageNum < 100
				},
			)
		},
	},
	"AWS/StorageGateway": {
		ResourceFunc: func(ctx context.Context, iface TagsInterface, job *config.Job, region string) (resources []*TaggedResource, err error) {
			pageNum := 0
			return resources, iface.StoragegatewayClient.ListGatewaysPagesWithContext(ctx, &storagegateway.ListGatewaysInput{},
				func(page *storagegateway.ListGatewaysOutput, more bool) bool {
					pageNum++
					promutil.StoragegatewayAPICounter.Inc()

					for _, gwa := range page.Gateways {
						resource := TaggedResource{
							ARN:       fmt.Sprintf("%s/%s", *gwa.GatewayId, *gwa.GatewayName),
							Namespace: job.Type,
							Region:    region,
						}

						tagsRequest := &storagegateway.ListTagsForResourceInput{
							ResourceARN: gwa.GatewayARN,
						}
						tagsResponse, _ := iface.StoragegatewayClient.ListTagsForResource(tagsRequest)
						promutil.StoragegatewayAPICounter.Inc()

						for _, t := range tagsResponse.Tags {
							resource.Tags = append(resource.Tags, model.Tag{Key: *t.Key, Value: *t.Value})
						}

						if resource.FilterThroughTags(job.SearchTags) {
							resources = append(resources, &resource)
						}
					}

					return pageNum < 100
				},
			)
		},
	},
	"AWS/TransitGateway": {
		ResourceFunc: func(ctx context.Context, iface TagsInterface, job *config.Job, region string) (resources []*TaggedResource, err error) {
			pageNum := 0
			return resources, iface.Ec2Client.DescribeTransitGatewayAttachmentsPagesWithContext(ctx, &ec2.DescribeTransitGatewayAttachmentsInput{},
				func(page *ec2.DescribeTransitGatewayAttachmentsOutput, more bool) bool {
					pageNum++
					promutil.Ec2APICounter.Inc()

					for _, tgwa := range page.TransitGatewayAttachments {
						resource := TaggedResource{
							ARN:       fmt.Sprintf("%s/%s", *tgwa.TransitGatewayId, *tgwa.TransitGatewayAttachmentId),
							Namespace: job.Type,
							Region:    region,
						}

						for _, t := range tgwa.Tags {
							resource.Tags = append(resource.Tags, model.Tag{Key: *t.Key, Value: *t.Value})
						}

						if resource.FilterThroughTags(job.SearchTags) {
							resources = append(resources, &resource)
						}
					}
					return pageNum < 100
				},
			)
		},
	},
}
