package diag

import (
	"io"
)

const Text = ".txt"

const criticalErrorFooter = `

A configuration file was not generated due to critical issues. Review the diagnostics above.`

const errorFooter = `

A configuration file was not generated due to errors. Review the diagnostics above.

These may be bypassed which is not recommended for production use.`

const successFooter = `

A configuration file was generated successfully.`

// generateTextReport generates a text report for the diagnostics.
func generateTextReport(writer io.Writer, ds Diagnostics, bypassErrors bool) error {
	content := getContent(ds, bypassErrors)
	_, err := writer.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}

// getContent returns the formatted content for the report based on the diagnostics and bypassErrors.
func getContent(ds Diagnostics, bypassErrors bool) string {
	var content string
	switch {
	case ds.HasSeverityLevel(SeverityLevelCritical):
		content = criticalErrorFooter
		ds.RemoveDiagsBySeverity(SeverityLevelInfo)
		ds.RemoveDiagsBySeverity(SeverityLevelWarn)
		ds.RemoveDiagsBySeverity(SeverityLevelError)
	case ds.HasSeverityLevel(SeverityLevelError) && !bypassErrors:
		content = errorFooter
		ds.RemoveDiagsBySeverity(SeverityLevelInfo)
		ds.RemoveDiagsBySeverity(SeverityLevelWarn)
	default:
		content = successFooter
	}

	return ds.Error() + content
}
