package config

// LoaderConfigOptions is used to provide a set of options when a new config is loaded.
type LoaderConfigOptions struct {
	Scope interface{}
}

func DefaultLoaderConfigOptions() LoaderConfigOptions {
	return LoaderConfigOptions{}
}
