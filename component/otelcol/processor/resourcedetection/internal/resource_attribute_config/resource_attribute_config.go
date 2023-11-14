package resource_attribute_config

// Configures whether a resource attribute should be enabled or not.
type ResourceAttributeConfig struct {
	// "enabled" as a mandatory parameter, because if this block is present in the config,
	// it makes sense for the user to explicitly say whether they want to enable the attribute.
	//
	// Unlike the Collector, the Agent does not try to set a default value for "enabled" because:
	// * Different resource attributes have different default values.
	//   It is time consuming to try to keep default values in sync with the Collector.
	// * If we set a default value in the Agent, Collector will think that the user set it explicitly.
	//   This is due to an "enabledSetByUser" parameter which Collector uses to print warnings such as:
	//     * "[WARNING] Please set `enabled` field explicitly for `default.metric`: This metric will be disabled by default soon."
	//     * "[WARNING] `default.metric.to_be_removed` should not be enabled: This metric is deprecated and will be removed soon."
	//   Users who did not explicitly enable such resource attributes may be confused by these warnings.
	//   Therefore it's easier to not set a default value in the Agent.
	Enabled bool `river:"enabled,attr"`
}

func (r *ResourceAttributeConfig) Convert() map[string]interface{} {
	if r == nil {
		return nil
	}

	return map[string]interface{}{
		"enabled": r.Enabled,
	}
}
