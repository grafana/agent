package diag

import (
	"html/template"
	"os"
	"strings"
)

const Text = ".txt"
const HTML = ".html"

// generateTextReport generates a text report for the diagnostics.
func generateTextReport(ds Diagnostics, filename string) error {
	content := ds.Error()

	err := writeToFile(content, filename)
	if err != nil {
		return err
	}

	return nil
}

// generateHTMLReport generates an HTML report for the diagnostics.
func generateHTMLReport(ds Diagnostics, filename string) error {
	// Define the HTML template
	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Diagnostics</title>
		<style>
			.critical { color: red; }
			.error { color: orange; }
			.warning { color: yellow; }
			.info { color: green; }
		</style>
	</head>
	<body>
		<h1>Diagnostics</h1>
		<ul>
			{{range .}}
			<li><strong><span class="{{getClass .Severity}}">{{.Severity}}</span></strong>: {{.Message}}</li>
			{{end}}
		</ul>
	</body>
	</html>
	`

	// Define a function to get the class based on severity
	funcMap := template.FuncMap{
		"getClass": func(severity Severity) string {
			switch severity {
			case SeverityLevelCritical:
				return "critical"
			case SeverityLevelError:
				return "error"
			case SeverityLevelWarn:
				return "warning"
			case SeverityLevelInfo:
				return "info"
			default:
				return ""
			}
		},
	}

	// Parse the template with the function map
	t, err := template.New("ds").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	// Generate the HTML content
	var sb strings.Builder
	err = t.Execute(&sb, ds)
	if err != nil {
		return err
	}

	err = writeToFile(sb.String(), filename)
	if err != nil {
		return err
	}

	return nil
}

// writeToFile writes the content to a file.
func writeToFile(content, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
