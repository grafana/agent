package lambda

import rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"

type Config struct {
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

func (args *Config) Convert() map[string]interface{} {
	return map[string]interface{}{
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/lambda resource attributes.
type ResourceAttributesConfig struct {
	AwsLogGroupNames  *rac.ResourceAttributeConfig `river:"aws.log.group.names,block,optional"`
	AwsLogStreamNames *rac.ResourceAttributeConfig `river:"aws.log.stream.names,block,optional"`
	CloudPlatform     *rac.ResourceAttributeConfig `river:"cloud.platform,block,optional"`
	CloudProvider     *rac.ResourceAttributeConfig `river:"cloud.provider,block,optional"`
	CloudRegion       *rac.ResourceAttributeConfig `river:"cloud.region,block,optional"`
	FaasInstance      *rac.ResourceAttributeConfig `river:"faas.instance,block,optional"`
	FaasMaxMemory     *rac.ResourceAttributeConfig `river:"faas.max_memory,block,optional"`
	FaasName          *rac.ResourceAttributeConfig `river:"faas.name,block,optional"`
	FaasVersion       *rac.ResourceAttributeConfig `river:"faas.version,block,optional"`
}

func (r *ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"aws.log.group.names":  r.AwsLogGroupNames.Convert(),
		"aws.log.stream.names": r.AwsLogStreamNames.Convert(),
		"cloud.platform":       r.CloudPlatform.Convert(),
		"cloud.provider":       r.CloudProvider.Convert(),
		"cloud.region":         r.CloudRegion.Convert(),
		"faas.instance":        r.FaasInstance.Convert(),
		"faas.max_memory":      r.FaasMaxMemory.Convert(),
		"faas.name":            r.FaasName.Convert(),
		"faas.version":         r.FaasVersion.Convert(),
	}
}
