// Package converter exposes utilities to convert config files from other
// programs to Grafana Agent Flow configurations.
package converter

import (
	"bytes"
	"fmt"

	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/printer"
)

// Input represents the type of config file being fed into the converter.
type Input string

const (
	// InputPrometheus indicates that the input file is a prometheus.yaml file.
	InputPrometheus Input = "prometheus"
)

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
		return prettyPrint(prometheusconvert.Convert(in))
	}

	var diags diag.Diagnostics
	diags.Add(diag.SeverityLevelError, fmt.Sprintf("unrecognized kind %q", kind))
	return nil, diags
}

// prettyPrint attempts to pretty print the input slice. If prettyPrint fails,
// the input slice is returned unmodified.
func prettyPrint(in []byte, diags diag.Diagnostics) ([]byte, diag.Diagnostics) {
	// Return early if there was no file.
	if len(in) == 0 {
		return in, diags
	}

	f, err := parser.ParseFile("", in)
	if err != nil {
		diags.Add(diag.SeverityLevelWarn, err.Error())
		return in, diags
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, f); err != nil {
		diags.Add(diag.SeverityLevelWarn, err.Error())
		return in, diags
	}

	// Add a trailing newline at the end of the file, which is omitted by Fprint.
	_, _ = buf.Write([]byte{'\n'})
	return buf.Bytes(), nil
}
