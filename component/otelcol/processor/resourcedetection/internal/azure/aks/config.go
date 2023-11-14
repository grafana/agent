package aks

import rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"

type Config struct {
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

func (args *Config) Convert() map[string]interface{} {
	return map[string]interface{}{
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/aks resource attributes.
type ResourceAttributesConfig struct {
	CloudPlatform *rac.ResourceAttributeConfig `river:"cloud.platform,block,optional"`
	CloudProvider *rac.ResourceAttributeConfig `river:"cloud.provider,block,optional"`
}

func (r *ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"cloud.platform": r.CloudPlatform.Convert(),
		"cloud.provider": r.CloudProvider.Convert(),
	}
}
