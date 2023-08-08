package docs

import (
	"flag"
	"strings"
	"testing"

	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
	"github.com/grafana/agent/docs/generator"
	"github.com/stretchr/testify/require"
)

// Run the below generate command to automatically update the Markdown docs with generated content
//go:generate go test -run TestCompatibleComponentsSectionUpdated -fix-tests

var fixTestsFlag = flag.Bool("fix-tests", false, "update the test files with the current generated content")

func TestCompatibleComponentsSectionUpdated(t *testing.T) {
	for _, name := range component.AllNames() {
		t.Run(name, func(t *testing.T) {
			generated, err := generator.GenerateCompatibleComponentsSection(name)
			require.NoError(t, err, "failed to generate references section for %q", name)

			if generated == "" {
				t.Skipf("no compatible components section defined for %q", name)
			}

			if *fixTestsFlag {
				err = generator.WriteCompatibleComponentsSection(name)
				require.NoError(t, err, "failed to write generated references section for %q", name)
				t.Log("updated the docs with generated content")
			}

			actual, err := generator.ReadCompatibleComponentsSection(name)
			require.NoError(t, err, "failed to read generated components section for %q", name)
			require.Contains(
				t,
				actual,
				strings.TrimSpace(generated),
				"expected documentation for %q to contain generated references section",
				name,
			)
		})
	}
}
