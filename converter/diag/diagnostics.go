// Package diag exposes error types used throughout converter and a method to
// pretty-print them to the screen.
package diag

// Severity denotes the severity level of a diagnostic. The zero value of
// severity is invalid.
type Severity int

// Supported severity levels.
const (
	SeverityLevelWarn Severity = iota + 1
	SeverityLevelError
)

// Diagnostic is an individual diagnostic message. Diagnostic messages can have
// different levels of severities.
type Diagnostic struct {
	// Severity holds the severity level of this Diagnostic.
	Severity Severity

	Message string
}

// Error implements error.
func (d Diagnostic) Error() string {
	return d.Message
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
