package ec2

import rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"

// Config defines user-specified configurations unique to the EC2 detector
type Config struct {
	// Tags is a list of regex's to match ec2 instance tag keys that users want
	// to add as resource attributes to processed data
	Tags               []string                 `river:"tags,attr,optional"`
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

func (args *Config) Convert() map[string]interface{} {
	var tags []string
	if args.Tags != nil {
		tags = append([]string{}, args.Tags...)
	}

	return map[string]interface{}{
		"tags":                tags,
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/ec2 resource attributes.
type ResourceAttributesConfig struct {
	CloudAccountID        *rac.ResourceAttributeConfig `river:"cloud.account.id,block,optional"`
	CloudAvailabilityZone *rac.ResourceAttributeConfig `river:"cloud.availability_zone,block,optional"`
	CloudPlatform         *rac.ResourceAttributeConfig `river:"cloud.platform,block,optional"`
	CloudProvider         *rac.ResourceAttributeConfig `river:"cloud.provider,block,optional"`
	CloudRegion           *rac.ResourceAttributeConfig `river:"cloud.region,block,optional"`
	HostID                *rac.ResourceAttributeConfig `river:"host.id,block,optional"`
	HostImageID           *rac.ResourceAttributeConfig `river:"host.image.id,block,optional"`
	HostName              *rac.ResourceAttributeConfig `river:"host.name,block,optional"`
	HostType              *rac.ResourceAttributeConfig `river:"host.type,block,optional"`
}

func (r *ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"cloud.account.id":        r.CloudAccountID.Convert(),
		"cloud.availability_zone": r.CloudAvailabilityZone.Convert(),
		"cloud.platform":          r.CloudPlatform.Convert(),
		"cloud.provider":          r.CloudProvider.Convert(),
		"cloud.region":            r.CloudRegion.Convert(),
		"host.id":                 r.HostID.Convert(),
		"host.image.id":           r.HostImageID.Convert(),
		"host.name":               r.HostName.Convert(),
		"host.type":               r.HostType.Convert(),
	}
}
