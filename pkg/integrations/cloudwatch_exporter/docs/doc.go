package main

import (
	_ "embed"
	"fmt"
	yaceSvc "github.com/nerdswords/yet-another-cloudwatch-exporter/pkg/services"
	"strings"
)

//go:embed template.md
var docTemplate string

const servicesListReplacement = "{{SERVICE_LIST}}"

// main is used for programmatically generating a documentation section containing all AWS services supported in cloudwatch
// exporter discovery jobs.
func main() {
	var sb strings.Builder
	for _, supportedSvc := range yaceSvc.SupportedServices {
		sb.WriteString(
			fmt.Sprintf("- Namespace: \"%s\" or Alias: \"%s\"\n", supportedSvc.Namespace, supportedSvc.Alias),
		)
	}
	doc := strings.Replace(docTemplate, servicesListReplacement, sb.String(), 1)
	fmt.Println(doc)
}
