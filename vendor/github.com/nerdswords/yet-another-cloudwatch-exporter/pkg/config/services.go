package config

import (
	"github.com/aws/aws-sdk-go/aws"
)

type ServiceConfig struct {
	Namespace        string
	Alias            string
	ResourceFilters  []*string
	DimensionRegexps []*string
}

type serviceConfigs []ServiceConfig

func (sc serviceConfigs) GetService(serviceType string) *ServiceConfig {
	for _, sf := range sc {
		if sf.Alias == serviceType || sf.Namespace == serviceType {
			return &sf
		}
	}
	return nil
}

var SupportedServices = serviceConfigs{
	{
		Namespace: "AWS/CertificateManager",
		Alias:     "acm",
		ResourceFilters: []*string{
			aws.String("acm:certificate"),
		},
	},
	{
		Namespace: "AmazonMWAA",
		Alias:     "airflow",
		ResourceFilters: []*string{
			aws.String("airflow"),
		},
	},
	{
		Namespace: "AWS/ApplicationELB",
		Alias:     "alb",
		ResourceFilters: []*string{
			aws.String("elasticloadbalancing:loadbalancer/app"),
			aws.String("elasticloadbalancing:targetgroup"),
		},
		DimensionRegexps: []*string{
			aws.String(":(?P<TargetGroup>targetgroup/.+)"),
			aws.String(":loadbalancer/(?P<LoadBalancer>.+)$"),
		},
	},
	{
		Namespace: "AWS/AppStream",
		Alias:     "appstream",
		ResourceFilters: []*string{
			aws.String("appstream"),
		},
		DimensionRegexps: []*string{
			aws.String(":fleet/(?P<FleetName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Backup",
		Alias:     "backup",
		ResourceFilters: []*string{
			aws.String("backup"),
		},
	},
	{
		Namespace: "AWS/ApiGateway",
		Alias:     "apigateway",
		ResourceFilters: []*string{
			aws.String("apigateway"),
		},
		DimensionRegexps: []*string{
			aws.String("apis/(?P<ApiName>[^/]+)$"),
			aws.String("apis/(?P<ApiName>[^/]+)/stages/(?P<Stage>[^/]+)$"),
		},
	},
	{
		Namespace: "AWS/AmazonMQ",
		Alias:     "mq",
		ResourceFilters: []*string{
			aws.String("mq"),
		},
		DimensionRegexps: []*string{
			aws.String("broker:(?P<Broker>[^:]+)"),
		},
	},
	{
		Namespace: "AWS/AppSync",
		Alias:     "appsync",
		ResourceFilters: []*string{
			aws.String("appsync"),
		},
		DimensionRegexps: []*string{
			aws.String("apis/(?P<GraphQLAPIId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Athena",
		Alias:     "athena",
		ResourceFilters: []*string{
			aws.String("athena"),
		},
		DimensionRegexps: []*string{
			aws.String("workgroup/(?P<WorkGroup>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/AutoScaling",
		Alias:     "asg",
		DimensionRegexps: []*string{
			aws.String("autoScalingGroupName/(?P<AutoScalingGroupName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/ElasticBeanstalk",
		Alias:     "beanstalk",
	},
	{
		Namespace: "AWS/Billing",
		Alias:     "billing",
	},
	{
		Namespace: "AWS/Cassandra",
		Alias:     "cassandra",
		ResourceFilters: []*string{
			aws.String("cassandra"),
		},
	},
	{
		Namespace: "AWS/CloudFront",
		Alias:     "cloudfront",
		ResourceFilters: []*string{
			aws.String("cloudfront:distribution"),
		},
		DimensionRegexps: []*string{
			aws.String("distribution/(?P<DistributionId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Cognito",
		Alias:     "cognito-idp",
		ResourceFilters: []*string{
			aws.String("cognito-idp:userpool"),
		},
		DimensionRegexps: []*string{
			aws.String("userpool/(?P<UserPool>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/DMS",
		Alias:     "dms",
		ResourceFilters: []*string{
			aws.String("dms"),
		},
		DimensionRegexps: []*string{
			aws.String("rep:[^/]+/(?P<ReplicationInstanceIdentifier>[^/]+)"),
			aws.String("task:(?P<ReplicationTaskIdentifier>[^/]+)/(?P<ReplicationInstanceIdentifier>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/DDoSProtection",
		Alias:     "shield",
		ResourceFilters: []*string{
			aws.String("shield:protection"),
		},
	},
	{
		Namespace: "AWS/DocDB",
		Alias:     "docdb",
		ResourceFilters: []*string{
			aws.String("rds:db"),
			aws.String("rds:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String("cluster:(?P<DBClusterIdentifier>[^/]+)"),
			aws.String("db:(?P<DBInstanceIdentifier>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/DX",
		Alias:     "dx",
		ResourceFilters: []*string{
			aws.String("directconnect"),
		},
		DimensionRegexps: []*string{
			aws.String(":dxcon/(?P<ConnectionId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/DynamoDB",
		Alias:     "dynamodb",
		ResourceFilters: []*string{
			aws.String("dynamodb:table"),
		},
		DimensionRegexps: []*string{
			aws.String(":table/(?P<TableName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/EBS",
		Alias:     "ebs",
		ResourceFilters: []*string{
			aws.String("ec2:volume"),
		},
		DimensionRegexps: []*string{
			aws.String("volume/(?P<VolumeId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/ElastiCache",
		Alias:     "ec",
		ResourceFilters: []*string{
			aws.String("elasticache:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String("cluster:(?P<CacheClusterId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/EC2",
		Alias:     "ec2",
		ResourceFilters: []*string{
			aws.String("ec2:instance"),
		},
		DimensionRegexps: []*string{
			aws.String("instance/(?P<InstanceId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/EC2Spot",
		Alias:     "ec2Spot",
		DimensionRegexps: []*string{
			aws.String("(?P<FleetRequestId>.*)"),
		},
	},
	{
		Namespace: "AWS/ECS",
		Alias:     "ecs-svc",
		ResourceFilters: []*string{
			aws.String("ecs:cluster"),
			aws.String("ecs:service"),
		},
		DimensionRegexps: []*string{
			aws.String("cluster/(?P<ClusterName>[^/]+)"),
			aws.String("service/(?P<ClusterName>[^/]+)/([^/]+)"),
		},
	},
	{
		Namespace: "ECS/ContainerInsights",
		Alias:     "ecs-containerinsights",
		ResourceFilters: []*string{
			aws.String("ecs:cluster"),
			aws.String("ecs:service"),
		},
		DimensionRegexps: []*string{
			aws.String("cluster/(?P<ClusterName>[^/]+)"),
			aws.String("service/(?P<ClusterName>[^/]+)/([^/]+)"),
		},
	},
	{
		Namespace: "AWS/EFS",
		Alias:     "efs",
		ResourceFilters: []*string{
			aws.String("elasticfilesystem:file-system"),
		},
		DimensionRegexps: []*string{
			aws.String("file-system/(?P<FileSystemId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/ELB",
		Alias:     "elb",
		ResourceFilters: []*string{
			aws.String("elasticloadbalancing:loadbalancer"),
		},
		DimensionRegexps: []*string{
			aws.String(":loadbalancer/(?P<LoadBalancerName>.+)$"),
		},
	},
	{
		Namespace: "AWS/ElasticMapReduce",
		Alias:     "emr",
		ResourceFilters: []*string{
			aws.String("elasticmapreduce:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String("cluster/(?P<JobFlowId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/EMRServerless",
		Alias:     "emr-serverless",
		ResourceFilters: []*string{
			aws.String("emr-serverless:applications"),
		},
		DimensionRegexps: []*string{
			aws.String("applications/(?P<ApplicationId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/ES",
		Alias:     "es",
		ResourceFilters: []*string{
			aws.String("es:domain"),
		},
		DimensionRegexps: []*string{
			aws.String(":domain/(?P<DomainName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Firehose",
		Alias:     "firehose",
		ResourceFilters: []*string{
			aws.String("firehose"),
		},
		DimensionRegexps: []*string{
			aws.String(":deliverystream/(?P<DeliveryStreamName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/FSx",
		Alias:     "fsx",
		ResourceFilters: []*string{
			aws.String("fsx:file-system"),
		},
		DimensionRegexps: []*string{
			aws.String("file-system/(?P<FileSystemId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/GameLift",
		Alias:     "gamelift",
		ResourceFilters: []*string{
			aws.String("gamelift"),
		},
		DimensionRegexps: []*string{
			aws.String(":fleet/(?P<FleetId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/GlobalAccelerator",
		Alias:     "ga",
		ResourceFilters: []*string{
			aws.String("globalaccelerator"),
		},
		DimensionRegexps: []*string{
			aws.String("destinationEdge/(?P<DestinationEdge>[^/]+)"),
			aws.String("accelerator/(?P<Accelerator>[^/]+)"),
			aws.String("endpointGroup/(?P<EndpointGroup>[^/]+)"),
			aws.String("listener/(?P<Listener>[^/]+)"),
			aws.String("transportProtocol/(?P<TransportProtocol>[^/]+)"),
		},
	},
	{
		Namespace: "Glue",
		Alias:     "glue",
		ResourceFilters: []*string{
			aws.String("glue:job"),
		},
		DimensionRegexps: []*string{
			aws.String(":job/(?P<JobName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/IoT",
		Alias:     "iot",
		ResourceFilters: []*string{
			aws.String("iot:rule"),
			aws.String("iot:provisioningtemplate"),
		},
		DimensionRegexps: []*string{
			aws.String(":rule/(?P<RuleName>[^/]+)"),
			aws.String(":provisioningtemplate/(?P<TemplateName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Kafka",
		Alias:     "kafka",
		ResourceFilters: []*string{
			aws.String("kafka:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String(":cluster/(?P<Cluster_Name>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/KafkaConnect",
		Alias:     "kafkaconnect",
		ResourceFilters: []*string{
			aws.String("kafkaconnect"),
		},
		DimensionRegexps: []*string{
			aws.String(":connector/(?P<Connector_Name>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Kinesis",
		Alias:     "kinesis",
		ResourceFilters: []*string{
			aws.String("kinesis:stream"),
		},
		DimensionRegexps: []*string{
			aws.String(":stream/(?P<StreamName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/KinesisAnalytics",
		Alias:     "kinesis-analytics",
	},
	{
		Namespace: "AWS/Lambda",
		Alias:     "lambda",
		ResourceFilters: []*string{
			aws.String("lambda:function"),
		},
		DimensionRegexps: []*string{
			aws.String(":function:(?P<FunctionName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/MediaTailor",
		Alias:     "mediatailor",
		ResourceFilters: []*string{
			aws.String("mediatailor:playbackConfiguration"),
		},
		DimensionRegexps: []*string{
			aws.String("playbackConfiguration/(?P<ConfigurationName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Neptune",
		Alias:     "neptune",
		ResourceFilters: []*string{
			aws.String("rds:db"),
			aws.String("rds:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String(":cluster:(?P<DBClusterIdentifier>[^/]+)"),
			aws.String(":db:(?P<DBInstanceIdentifier>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/NetworkFirewall",
		Alias:     "nfw",
		ResourceFilters: []*string{
			aws.String("network-firewall:firewall"),
		},
		DimensionRegexps: []*string{
			aws.String("firewall/(?P<FirewallName>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/NATGateway",
		Alias:     "ngw",
		ResourceFilters: []*string{
			aws.String("ec2:natgateway"),
		},
		DimensionRegexps: []*string{
			aws.String("natgateway/(?P<NatGatewayId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/NetworkELB",
		Alias:     "nlb",
		ResourceFilters: []*string{
			aws.String("elasticloadbalancing:loadbalancer/net"),
			aws.String("elasticloadbalancing:targetgroup"),
		},
		DimensionRegexps: []*string{
			aws.String(":(?P<TargetGroup>targetgroup/.+)"),
			aws.String(":loadbalancer/(?P<LoadBalancer>.+)$"),
		},
	},
	{
		Namespace: "AWS/PrivateLinkEndpoints",
		Alias:     "vpc-endpoint",
		ResourceFilters: []*string{
			aws.String("ec2:vpc-endpoint"),
		},
		DimensionRegexps: []*string{
			aws.String(":vpc-endpoint/(?P<VPC_Endpoint_Id>.+)"),
		},
	},
	{
		Namespace: "AWS/PrivateLinkServices",
		Alias:     "vpc-endpoint-service",
		ResourceFilters: []*string{
			aws.String("ec2:vpc-endpoint-service"),
		},
		DimensionRegexps: []*string{
			aws.String(":vpc-endpoint-service:(?P<Service_Id>.+)"),
		},
	},
	{
		Namespace: "AWS/Prometheus",
		Alias:     "amp",
	},
	{
		Namespace: "AWS/RDS",
		Alias:     "rds",
		ResourceFilters: []*string{
			aws.String("rds:db"),
			aws.String("rds:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String(":cluster:(?P<DBClusterIdentifier>[^/]+)"),
			aws.String(":db:(?P<DBInstanceIdentifier>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Redshift",
		Alias:     "redshift",
		ResourceFilters: []*string{
			aws.String("redshift:cluster"),
		},
		DimensionRegexps: []*string{
			aws.String(":cluster:(?P<ClusterIdentifier>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Route53Resolver",
		Alias:     "route53-resolver",
		ResourceFilters: []*string{
			aws.String("route53resolver"),
		},
		DimensionRegexps: []*string{
			aws.String(":resolver-endpoint/(?P<EndpointId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/Route53",
		Alias:     "route53",
		ResourceFilters: []*string{
			aws.String("route53"),
		},
		DimensionRegexps: []*string{
			aws.String(":healthcheck/(?P<HealthCheckId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/S3",
		Alias:     "s3",
		ResourceFilters: []*string{
			aws.String("s3"),
		},
		DimensionRegexps: []*string{
			aws.String("(?P<BucketName>[^:]+)$"),
		},
	},
	{
		Namespace: "AWS/SES",
		Alias:     "ses",
	},
	{
		Namespace: "AWS/States",
		Alias:     "sfn",
		ResourceFilters: []*string{
			aws.String("states"),
		},
		DimensionRegexps: []*string{
			aws.String("(?P<StateMachineArn>.*)"),
		},
	},
	{
		Namespace: "AWS/SNS",
		Alias:     "sns",
		ResourceFilters: []*string{
			aws.String("sns"),
		},
		DimensionRegexps: []*string{
			aws.String("(?P<TopicName>[^:]+)$"),
		},
	},
	{
		Namespace: "AWS/SQS",
		Alias:     "sqs",
		ResourceFilters: []*string{
			aws.String("sqs"),
		},
		DimensionRegexps: []*string{
			aws.String("(?P<QueueName>[^:]+)$"),
		},
	},
	{
		Namespace: "AWS/StorageGateway",
		Alias:     "storagegateway",
		ResourceFilters: []*string{
			aws.String("storagegateway"),
		},
		DimensionRegexps: []*string{
			aws.String(":gateway/(?P<GatewayId>[^:]+)$"),
			aws.String(":share/(?P<ShareId>[^:]+)$"),
			aws.String("^(?P<GatewayId>[^:/]+)/(?P<GatewayName>[^:]+)$"),
		},
	},
	{
		Namespace: "AWS/TransitGateway",
		Alias:     "tgw",
		ResourceFilters: []*string{
			aws.String("ec2:transit-gateway"),
		},
		DimensionRegexps: []*string{
			aws.String(":transit-gateway/(?P<TransitGateway>[^/]+)"),
			aws.String("(?P<TransitGateway>[^/]+)/(?P<TransitGatewayAttachment>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/VPN",
		Alias:     "vpn",
		ResourceFilters: []*string{
			aws.String("ec2:vpn-connection"),
		},
		DimensionRegexps: []*string{
			aws.String(":vpn-connection/(?P<VpnId>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/WAFV2",
		Alias:     "wafv2",
		ResourceFilters: []*string{
			aws.String("wafv2"),
		},
		DimensionRegexps: []*string{
			aws.String("/webacl/(?P<WebACL>[^/]+)"),
		},
	},
	{
		Namespace: "AWS/WorkSpaces",
		Alias:     "workspaces",
		ResourceFilters: []*string{
			aws.String("workspaces:workspace"),
			aws.String("workspaces:directory"),
		},
		DimensionRegexps: []*string{
			aws.String(":workspace/(?P<WorkspaceId>[^/]+)$"),
			aws.String(":directory/(?P<DirectoryId>[^/]+)$"),
		},
	},
}
