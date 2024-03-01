package sdkconfig

// Config defines the configuration options for the host_info connector.
type Arguments struct {
	// Configuration of the actual service
	Config string `river:"config,attr"`
}
