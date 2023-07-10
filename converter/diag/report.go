package diag

import (
	"io"
)

const Text = ".txt"

// generateTextReport generates a text report for the diagnostics.
func generateTextReport(ds Diagnostics, writer io.Writer) error {
	content := ds.Error()

	_, err := writer.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}
