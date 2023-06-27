// Package diag exposes error types used throughout converter and a method to
// pretty-print them to the screen.
package diag

import (
	"fmt"
	"strings"
)

// Severity denotes the severity level of a diagnostic. The zero value of
// severity is invalid.
type Severity int

var _ fmt.Stringer = (*Severity)(nil)

func (s Severity) String() string {
	switch s {
	case SeverityLevelError:
		return "Error"
	case SeverityLevelWarn:
		return "Warning"
	case SeverityLevelInfo:
		return "Info"
	default:
		return "Unknown"
	}
}

// implement fmt.Stringer

// Supported severity levels.
const (
	SeverityLevelInfo Severity = iota + 1
	SeverityLevelWarn
	SeverityLevelError
)

// Diagnostic is an individual diagnostic message. Diagnostic messages can have
// different levels of severities.
type Diagnostic struct {
	// Severity holds the severity level of this Diagnostic.
	Severity Severity

	Message string
}

var _ fmt.Stringer = (*Diagnostic)(nil)

func (d Diagnostic) String() string {
	return fmt.Sprintf("(%s) %s", d.Severity.String(), d.Message)
}

// Error implements error.
func (d Diagnostic) Error() string {
	return d.String()
}

// Diagnostics is a collection of diagnostic messages.
type Diagnostics []Diagnostic

// Add adds an individual Diagnostic to the diagnostics list.
func (ds *Diagnostics) Add(severity Severity, message string) {
	*ds = append(*ds, Diagnostic{
		Severity: severity,
		Message:  message,
	})
}

// Error implements error.
func (ds Diagnostics) Error() string {
	var sb strings.Builder
	for ix, diag := range ds {
		fmt.Fprint(&sb, diag.Error())
		if ix+1 < len(ds) {
			fmt.Fprintln(&sb)
		}
	}

	return sb.String()
}
