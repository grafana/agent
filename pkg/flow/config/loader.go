package config

// LoaderConfigOptions is used to provide a set of options when a new config is loaded.
type LoaderConfigOptions struct {
	// AdditionalDeclareContents can be used to pass custom components definition to the loader.
	// This is needed when a custom component is instantiated within a custom component and the corresponding
	// declare of the nested custom component is defined in a parent.
	AdditionalDeclareContents map[string]string
}

func DefaultLoaderConfigOptions() LoaderConfigOptions {
	return LoaderConfigOptions{}
}
