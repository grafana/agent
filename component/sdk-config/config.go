package sdkconfig

// ServiceConfig defines the SDK configuration for a service.
type ServiceConfig struct {
	Name   string `river:"name,attr"`
	Config string `river:"config,attr"`
}

// Arguments defines the configuration parameters for this component.
type Arguments struct {
	// Configuration of the actual service
	Service []*ServiceConfig `river:"service,block"`
}
