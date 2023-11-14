package azure

import rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"

type Config struct {
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

func (args *Config) Convert() map[string]interface{} {
	return map[string]interface{}{
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/azure resource attributes.
type ResourceAttributesConfig struct {
	AzureResourcegroupName *rac.ResourceAttributeConfig `river:"azure.resourcegroup.name,block,optional"`
	AzureVMName            *rac.ResourceAttributeConfig `river:"azure.vm.name,block,optional"`
	AzureVMScalesetName    *rac.ResourceAttributeConfig `river:"azure.vm.scaleset.name,block,optional"`
	AzureVMSize            *rac.ResourceAttributeConfig `river:"azure.vm.size,block,optional"`
	CloudAccountID         *rac.ResourceAttributeConfig `river:"cloud.account.id,block,optional"`
	CloudPlatform          *rac.ResourceAttributeConfig `river:"cloud.platform,block,optional"`
	CloudProvider          *rac.ResourceAttributeConfig `river:"cloud.provider,block,optional"`
	CloudRegion            *rac.ResourceAttributeConfig `river:"cloud.region,block,optional"`
	HostID                 *rac.ResourceAttributeConfig `river:"host.id,block,optional"`
	HostName               *rac.ResourceAttributeConfig `river:"host.name,block,optional"`
}

func (r *ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"azure.resourcegroup.name": r.AzureResourcegroupName.Convert(),
		"azure.vm.name":            r.AzureVMName.Convert(),
		"azure.vm.scaleset.name":   r.AzureVMScalesetName.Convert(),
		"azure.vm.size":            r.AzureVMSize.Convert(),
		"cloud.account.id":         r.CloudAccountID.Convert(),
		"cloud.platform":           r.CloudPlatform.Convert(),
		"cloud.provider":           r.CloudProvider.Convert(),
		"cloud.region":             r.CloudRegion.Convert(),
		"host.id":                  r.HostID.Convert(),
		"host.name":                r.HostName.Convert(),
	}
}
