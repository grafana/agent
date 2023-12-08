//go:build !windows

package docs

import (
	"flag"
	"strings"
	"testing"

	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
	"github.com/grafana/agent/component/metadata"
	"github.com/grafana/agent/docs/generator"
	"github.com/stretchr/testify/require"
)

// Run the below generate command to automatically update the Markdown docs with generated content
//go:generate go test -fix-tests -v

var fixTestsFlag = flag.Bool("fix-tests", false, "update the test files with the current generated content")

func TestLinksToTypesSectionsUpdated(t *testing.T) {
	for _, name := range component.AllNames() {
		t.Run(name, func(t *testing.T) {
			runForGenerator(t, generator.NewLinksToTypesGenerator(name))
		})
	}
}

func TestCompatibleComponentsPageUpdated(t *testing.T) {
	path := "sources/flow/reference/compatibility/_index.md"
	for _, typ := range metadata.AllTypes {
		t.Run(typ.Name, func(t *testing.T) {
			t.Run("exporters", func(t *testing.T) {
				runForGenerator(t, generator.NewExportersListGenerator(typ, path))
			})
			t.Run("consumers", func(t *testing.T) {
				runForGenerator(t, generator.NewConsumersListGenerator(typ, path))
			})
		})
	}
}

func runForGenerator(t *testing.T, g generator.DocsGenerator) {
	if *fixTestsFlag {
		err := g.Write()
		require.NoError(t, err, "failed to write generated content for: %q", g.Name())
		t.Log("updated the docs with generated content", g.Name())
		return
	}

	generated, err := g.Generate()
	require.NoError(t, err, "failed to generate: %q", g.Name())

	if strings.TrimSpace(generated) == "" {
		actual, err := g.Read()
		require.Error(t, err, "expected error when reading existing generated docs for %q", g.Name())
		require.Contains(t, err.Error(), "markers not found", "expected error to be about missing markers")
		require.Empty(t, actual, "expected empty actual content for %q", g.Name())
		return
	}

	actual, err := g.Read()
	require.NoError(t, err, "failed to read existing generated docs for %q, try running 'go generate ./docs'", g.Name())
	require.Contains(
		t,
		actual,
		strings.TrimSpace(generated),
		"outdated docs detected when running %q, try updating with 'go generate ./docs'",
		g.Name(),
	)
}
