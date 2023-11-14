package resourcedetection

import (
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/processor"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/aws/ec2"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/aws/ecs"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/aws/eks"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/aws/elasticbeanstalk"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/aws/lambda"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/azure"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/azure/aks"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/consul"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/docker"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/gcp"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/heroku"
	kubernetes_node "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/k8snode"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/openshift"
	"github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/system"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/grafana/river"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelextension "go.opentelemetry.io/collector/extension"
)

func init() {
	component.Register(component.Registration{
		Name:          "otelcol.processor.resourcedetection",
		Args:          Arguments{},
		Exports:       otelcol.ConsumerExports{},
		NeedsServices: []string{otel_service.ServiceName},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := resourcedetectionprocessor.NewFactory()
			return processor.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.processor.probabilistic_sampler component.
type Arguments struct {
	// Detectors is an ordered list of named detectors that should be
	// run to attempt to detect resource information.
	Detectors []string `river:"detectors,attr"`

	// Override indicates whether any existing resource attributes
	// should be overridden or preserved. Defaults to true.
	Override bool `river:"override,attr,optional"`

	// DetectorConfig is a list of settings specific to all detectors
	DetectorConfig DetectorConfig `river:",squash"`

	// HTTP client settings for the detector
	// Timeout default is 5s
	// Client otelcol.HTTPClientArguments `river:",squash"`
	Timeout time.Duration `river:"timeout,attr,optional"`
	//TODO: Uncomment this later? Can we just get away with a timeout, or do we need all the http client settings?
	//      It seems that HTTP client settings are only used in the ec2 detection via ClientFromContext.

	// Output configures where to send processed data. Required.
	Output *otelcol.ConsumerArguments `river:"output,block"`
}

// DetectorConfig contains user-specified configurations unique to all individual detectors
type DetectorConfig struct {
	// EC2Config contains user-specified configurations for the EC2 detector
	EC2Config ec2.Config `river:"ec2,block,optional"`

	// ECSConfig contains user-specified configurations for the ECS detector
	ECSConfig ecs.Config `river:"ecs,block,optional"`

	// EKSConfig contains user-specified configurations for the EKS detector
	EKSConfig eks.Config `river:"eks,block,optional"`

	// Elasticbeanstalk contains user-specified configurations for the elasticbeanstalk detector
	ElasticbeanstalkConfig elasticbeanstalk.Config `river:"elasticbeanstalk,block,optional"`

	// Lambda contains user-specified configurations for the lambda detector
	LambdaConfig lambda.Config `river:"lambda,block,optional"`

	// Azure contains user-specified configurations for the azure detector
	AzureConfig azure.Config `river:"azure,block,optional"`

	// Aks contains user-specified configurations for the aks detector
	AksConfig aks.Config `river:"aks,block,optional"`

	// ConsulConfig contains user-specified configurations for the Consul detector
	ConsulConfig consul.Config `river:"consul,block,optional"`

	// DockerConfig contains user-specified configurations for the docker detector
	DockerConfig docker.Config `river:"docker,block,optional"`

	// GcpConfig contains user-specified configurations for the gcp detector
	GcpConfig gcp.Config `river:"gcp,block,optional"`

	// HerokuConfig contains user-specified configurations for the heroku detector
	HerokuConfig heroku.Config `river:"heroku,block,optional"`

	// SystemConfig contains user-specified configurations for the System detector
	SystemConfig system.Config `river:"system,block,optional"`

	// OpenShift contains user-specified configurations for the Openshift detector
	OpenShiftConfig openshift.Config `river:"openshift,block,optional"`

	// KubernetesNode contains user-specified configurations for the K8SNode detector
	KubernetesNodeConfig kubernetes_node.Config `river:"kubernetes_node,block,optional"`
}

var (
	_ processor.Arguments = Arguments{}
	_ river.Validator     = (*Arguments)(nil)
	_ river.Defaulter     = (*Arguments)(nil)
)

// DefaultArguments holds default settings for Arguments.
var DefaultArguments = Arguments{
	Override: true,
	Timeout:  5 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if len(args.Detectors) == 0 {
		return fmt.Errorf("at least one detector must be specified")
	}
	return nil
}

func (args Arguments) ConvertDetectors() []string {
	if args.Detectors == nil {
		return nil
	}

	res := make([]string, 0, len(args.Detectors))
	for _, detector := range args.Detectors {
		//TODO(ptodev): Check if the detector name is valid
		switch detector {
		case "kubernetes_node":
			res = append(res, "k8snode")
		default:
			res = append(res, detector)
		}
	}
	return res
}

// Convert implements processor.Arguments.
func (args Arguments) Convert() (otelcomponent.Config, error) {
	input := make(map[string]interface{})

	input["detectors"] = args.ConvertDetectors()
	input["override"] = args.Override
	input["timeout"] = args.Timeout

	input["ec2"] = args.DetectorConfig.EC2Config.Convert()
	input["ecs"] = args.DetectorConfig.ECSConfig.Convert()
	input["eks"] = args.DetectorConfig.EKSConfig.Convert()
	input["elasticbeanstalk"] = args.DetectorConfig.ElasticbeanstalkConfig.Convert()
	input["lambda"] = args.DetectorConfig.LambdaConfig.Convert()
	input["azure"] = args.DetectorConfig.AzureConfig.Convert()
	input["aks"] = args.DetectorConfig.AksConfig.Convert()
	input["consul"] = args.DetectorConfig.ConsulConfig.Convert()
	input["docker"] = args.DetectorConfig.DockerConfig.Convert()
	input["gcp"] = args.DetectorConfig.GcpConfig.Convert()
	input["heroku"] = args.DetectorConfig.HerokuConfig.Convert()
	input["system"] = args.DetectorConfig.SystemConfig.Convert()
	input["openshift"] = args.DetectorConfig.OpenShiftConfig.Convert()
	input["k8snode"] = args.DetectorConfig.KubernetesNodeConfig.Convert()

	var result resourcedetectionprocessor.Config
	err := mapstructure.Decode(input, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Extensions implements processor.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelextension.Extension {
	return nil
}

// Exporters implements processor.Arguments.
func (args Arguments) Exporters() map[otelcomponent.DataType]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements processor.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}
