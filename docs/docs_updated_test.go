package docs

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"os"
	"strings"
	"testing"

	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
)

func allComponentsThat(f func(registration component.Registration) bool) []string {
	var result []string
	for _, name := range component.AllNames() {
		c, ok := component.Get(name)
		if !ok {
			continue
		}

		if f(c) {
			result = append(result, name)
		}
	}
	return result
}

func allComponentsThatOutput(dataType component.DataType) []string {
	return allComponentsThat(func(reg component.Registration) bool {
		return slices.Contains(reg.Metadata.Outputs, dataType)
	})
}

func allComponentsThatAccept(dataType component.DataType) []string {
	return allComponentsThat(func(reg component.Registration) bool {
		return slices.Contains(reg.Metadata.Accepts, dataType)
	})
}

func generateReferencesSection(t *testing.T, componentName string) string {
	c, ok := component.Get(componentName)
	require.True(t, ok, "expected component %q to exist", componentName)

	if c.Metadata.Empty() {
		return ""
	}

	heading := "\n## Compatible components\n\n"

	acceptingSection := acceptingComponentsSection(componentName, c)

	outputSection := outputComponentsSection(componentName, c)

	if acceptingSection == "" && outputSection == "" {
		return ""
	}

	note := "\nNote that connecting some components may not be feasible or components may require further " +
		"configuration to make the connection work correctly. " +
		"Please refer to the linked documentation for more details.\n"

	return heading + acceptingSection + outputSection + note
}

func outputComponentsSection(name string, c component.Registration) string {
	section := ""
	for _, outputDataType := range c.Metadata.Outputs {
		if list := listOfComponentsAccepting(outputDataType); list != "" {
			section += fmt.Sprintf("- %s:\n", outputDataType)
			section += list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can output data to the following components:\n\n", name) + section
	}
	return section
}

func listOfComponentsAccepting(dataType component.DataType) string {
	str := ""
	for _, linkedName := range allComponentsThatAccept(dataType) {
		str += fmt.Sprintf("  - [`%s`]()\n", linkedName)
	}
	return str
}

func acceptingComponentsSection(componentName string, c component.Registration) string {
	section := ""
	for _, acceptedDataType := range c.Metadata.Accepts {
		if list := listOfComponentsOutputting(acceptedDataType); list != "" {
			section += fmt.Sprintf("- %s:\n", acceptedDataType)
			section += list
		}
	}
	if section != "" {
		section = fmt.Sprintf("`%s` can accept data from the following components:\n\n", componentName) + section
	}
	return section
}

func listOfComponentsOutputting(dataType component.DataType) string {
	str := ""
	for _, linkedName := range allComponentsThatOutput(dataType) {
		str += fmt.Sprintf("  - [`%s`]()\n", linkedName)
	}
	return str
}

func TestGenerateReferencesSection(t *testing.T) {
	for _, name := range component.AllNames() {
		t.Run(name, func(t *testing.T) {
			generated := generateReferencesSection(t, name)

			filePath := fmt.Sprintf("sources/flow/reference/components/%s.md", name)
			contents, err := os.ReadFile(filePath)
			require.NoError(t, err, "failed to read %q", filePath)
			require.Contains(
				t,
				string(contents),
				strings.TrimSpace(generated),
				"expected documentation at %q to contain generated references section",
				filePath,
			)
		})
	}
}
