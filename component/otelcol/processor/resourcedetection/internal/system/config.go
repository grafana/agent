package system

import (
	"fmt"

	rac "github.com/grafana/agent/component/otelcol/processor/resourcedetection/internal/resource_attribute_config"
)

// Config defines user-specified configurations unique to the system detector
type Config struct {
	// The HostnameSources is a priority list of sources from which hostname will be fetched.
	// In case of the error in fetching hostname from source,
	// the next source from the list will be considered.
	HostnameSources []string `river:"hostname_sources,attr,optional"`

	ResourceAttributes ResourceAttributesConfig `river:"resource_attributes,block"`
}

var DefaultConfig = Config{
	HostnameSources: []string{"dns", "os"},
}

func (c *Config) SetToDefault() {
	*c = DefaultConfig
}

// Validate config
func (cfg *Config) Validate() error {
	for _, hostnameSource := range cfg.HostnameSources {
		switch hostnameSource {
		case "os", "dns", "cname", "lookup":
			// Valid option - nothing to do
		default:
			return fmt.Errorf("invalid hostname source: %s", hostnameSource)
		}
	}
	return nil
}

func (args *Config) Convert() map[string]interface{} {
	return map[string]interface{}{
		"hostname_sources":    args.HostnameSources,
		"resource_attributes": args.ResourceAttributes.Convert(),
	}
}

// ResourceAttributesConfig provides config for resourcedetectionprocessor/system resource attributes.
type ResourceAttributesConfig struct {
	HostArch           *rac.ResourceAttributeConfig `river:"host.arch,block,optional"`
	HostCPUCacheL2Size *rac.ResourceAttributeConfig `river:"host.cpu.cache.l2.size,block,optional"`
	HostCPUFamily      *rac.ResourceAttributeConfig `river:"host.cpu.family,block,optional"`
	HostCPUModelID     *rac.ResourceAttributeConfig `river:"host.cpu.model.id,block,optional"`
	HostCPUModelName   *rac.ResourceAttributeConfig `river:"host.cpu.model.name,block,optional"`
	HostCPUStepping    *rac.ResourceAttributeConfig `river:"host.cpu.stepping,block,optional"`
	HostCPUVendorID    *rac.ResourceAttributeConfig `river:"host.cpu.vendor.id,block,optional"`
	HostID             *rac.ResourceAttributeConfig `river:"host.id,block,optional"`
	HostName           *rac.ResourceAttributeConfig `river:"host.name,block,optional"`
	OsDescription      *rac.ResourceAttributeConfig `river:"os.description,block,optional"`
	OsType             *rac.ResourceAttributeConfig `river:"os.type,block,optional"`
}

func (r *ResourceAttributesConfig) Convert() map[string]interface{} {
	return map[string]interface{}{
		"host.arch":              r.HostArch.Convert(),
		"host.cpu.cache.l2.size": r.HostCPUCacheL2Size.Convert(),
		"host.cpu.family":        r.HostCPUFamily.Convert(),
		"host.cpu.model.id":      r.HostCPUModelID.Convert(),
		"host.cpu.model.name":    r.HostCPUModelName.Convert(),
		"host.cpu.stepping":      r.HostCPUStepping.Convert(),
		"host.cpu.vendor.id":     r.HostCPUVendorID.Convert(),
		"host.id":                r.HostID.Convert(),
		"host.name":              r.HostName.Convert(),
		"os.description":         r.OsDescription.Convert(),
		"os.type":                r.OsType.Convert(),
	}
}
