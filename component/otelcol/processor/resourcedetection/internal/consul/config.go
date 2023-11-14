package consul

import (
	rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"
	"github.com/grafana/river/rivertypes"
	"go.opentelemetry.io/collector/config/configopaque"
)

// The struct requires no user-specified fields by default as consul agent's default
// configuration will be provided to the API client.
// See `consul.go#NewDetector` for more information.
type Config struct {
	// Address is the address of the Consul server
	Address string `river:"address,attr,optional"`

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string `river:"datacenter,attr,optional"`

	// Token is used to provide a per-request ACL token which overrides the
	// agent's default (empty) token. Token is only required if
	// [Consul's ACL System](https://www.consul.io/docs/security/acl/acl-system)
	// is enabled.
	Token rivertypes.Secret `river:"token,attr,optional"`

	// TokenFile is not necessary in River because users can use the local.file
	// Flow component instead.
	//
	// TokenFile string `river:"token_file"`

	// Namespace is the name of the namespace to send along for the request
	// when no other Namespace is present in the QueryOptions
	Namespace string `river:"namespace,attr,optional"`

	// Allowlist of [Consul Metadata](https://www.consul.io/docs/agent/options#node_meta)
	// keys to use as resource attributes.
	MetaLabels []string `river:"meta,attr,optional"`

	// ResourceAttributes configuration for Consul detector
	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

func (args *Config) Convert() map[string]interface{} {
	//TODO(ptodev): Change the OTel Collector's "meta" param to be a slice instead of a map.
	var metaLabels map[string]string
	if args.MetaLabels != nil {
		metaLabels = make(map[string]string, len(args.MetaLabels))
		for _, label := range args.MetaLabels {
			metaLabels[label] = ""
		}
	}

	return map[string]interface{}{
		"address":             args.Address,
		"datacenter":          args.Datacenter,
		"token":               configopaque.String(args.Token),
		"namespace":           args.Namespace,
		"meta":                metaLabels,
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/consul resource attributes.
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
