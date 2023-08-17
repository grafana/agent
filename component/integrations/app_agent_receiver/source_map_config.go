package app_agent_receiver

import (
	"time"

	internal "github.com/grafana/agent/pkg/integrations/v2/app_agent_receiver"
)

type sourceMapConfig struct {
	download            bool                    `river:"download,bool"`
	downloadFromOrigins []string                `river:"download_from_origins,string,optional"`
	downloadTimeout     time.Duration           `river:"download_timeout,attr,optional"`
	fileSystem          []sourceMapFileLocation `river:"file_system,block,optional"`
}

type sourceMapFileLocation struct {
	path               string `yaml:"path,string"`
	minifiedPathPrefix string `yaml:"minifiedPathPrefix,string,optional"`
}

func (config *sourceMapConfig) toInternal() internal.SourceMapConfig {
	return internal.SourceMapConfig{
		Download:            config.download,
		DownloadFromOrigins: config.downloadFromOrigins,
		DownloadTimeout:     config.downloadTimeout,
		FileSystem:          sourceMapFileLocationsToInternals(config.fileSystem),
	}
}

func sourceMapFileLocationsToInternals(locations []sourceMapFileLocation) []internal.SourceMapFileLocation {
	var internals []internal.SourceMapFileLocation
	for _, l := range locations {
		internals = append(internals, internal.SourceMapFileLocation{
			Path:               l.path,
			MinifiedPathPrefix: l.minifiedPathPrefix,
		})
	}
	return internals
}
