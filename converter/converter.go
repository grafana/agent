// Package converter exposes utilities to convert config files from other
// programs to Grafana Agent Flow configurations.
package converter

import (
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/converter/internal/promtailconvert"
)

// Input represents the type of config file being fed into the converter.
type Input string

const (
	// InputPrometheus indicates that the input file is a prometheus YAML file.
	InputPrometheus Input = "prometheus"
	// InputPromtail indicates that the input file is a promtail YAML file.
	InputPromtail Input = "promtail"
)

var SupportedFormats = []string{
	string(InputPrometheus),
	string(InputPromtail),
}

// Convert generates a Grafana Agent Flow config given an input configuration
// file.
//
// Conversions are made as literally as possible, so the resulting config files
// may be unoptimized (i.e., lacking component reuse). A converted config file
// should just be the starting point rather than the final destination.
//
// Note that not all functionality defined in the input configuration may have
// an equivalent in Grafana Agent Flow. If the conversion could not complete
// because of mismatched functionality, an error is returned with no resulting
// config. If the conversion completed successfully but generated warnings, an
// error is returned alongside the resulting config.
func Convert(in []byte, kind Input) ([]byte, diag.Diagnostics) {
	switch kind {
	case InputPrometheus:
		return prometheusconvert.Convert(in)
	case InputPromtail:
		return promtailconvert.Convert(in)
	}

	var diags diag.Diagnostics
	diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("unrecognized kind %q given to the config converter", kind))
	return nil, diags
}
