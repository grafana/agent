package count

// MetricInfo for a data type
type MetricInfo struct {
	Name        string            `river:"name,attr"`
	Description string            `river:"description,attr,optional"`
	Conditions  []string          `river:"conditions,attr,optional"`
	Attributes  []AttributeConfig `river:"attributes,attr,optional"`
}

type AttributeConfig struct {
	Key          string `river:"key,attr"`
	DefaultValue string `river:"default_value,attr"`
}
