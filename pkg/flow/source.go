package flow

import (
	"crypto/sha256"
	"sort"

	"github.com/grafana/agent/pkg/config/encoder"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
)

// A Source holds the contents of a parsed Flow source
type Source struct {
	sourceMap map[string][]byte // Map that links parsed Flow source's name with its content.
	hash      [sha256.Size]byte // Hash of all files in sourceMap sorted by name.

	// Components holds the list of raw River AST blocks describing components.
	// The Flow controller can interpret them.
	components    []*ast.BlockStmt
	configBlocks  []*ast.BlockStmt
	declareBlocks []*ast.BlockStmt
}

// ParseSource parses the River file specified by bb into a File. name should be
// the name of the file used for reporting errors.
//
// bb must not be modified after passing to ParseSource.
func ParseSource(name string, bb []byte) (*Source, error) {
	bb, err := encoder.EnsureUTF8(bb, true)
	if err != nil {
		return nil, err
	}
	node, err := parser.ParseFile(name, bb)
	if err != nil {
		return nil, err
	}
	categorizedBlocks, err := controller.CategorizeStatements(node.Body)
	if err != nil {
		return nil, err
	}
	return &Source{
		components:    categorizedBlocks.Components,
		configBlocks:  categorizedBlocks.Configs,
		declareBlocks: categorizedBlocks.DeclareBlocks,
		sourceMap:     map[string][]byte{name: bb},
		hash:          sha256.Sum256(bb),
	}, nil
}

type namedSource struct {
	Name    string
	Content []byte
}

// ParseSources parses the map of sources and combines them into a single
// Source. sources must not be modified after calling ParseSources.
func ParseSources(sources map[string][]byte) (*Source, error) {
	var (
		mergedSource = &Source{sourceMap: sources} // Combined source from all the input content.
		hash         = sha256.New()                // Combined hash of all the sources.
	)

	// Sorted slice so ParseSources always does the same thing.
	sortedSources := make([]namedSource, 0, len(sources))
	for name, bb := range sources {
		sortedSources = append(sortedSources, namedSource{
			Name:    name,
			Content: bb,
		})
	}
	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].Name < sortedSources[j].Name
	})

	// Parse each .river source and compute new hash for the whole sourceMap
	for _, namedSource := range sortedSources {
		hash.Write(namedSource.Content)

		sourceFragment, err := ParseSource(namedSource.Name, namedSource.Content)
		if err != nil {
			return nil, err
		}

		mergedSource.components = append(mergedSource.components, sourceFragment.components...)
		mergedSource.configBlocks = append(mergedSource.configBlocks, sourceFragment.configBlocks...)
		mergedSource.declareBlocks = append(mergedSource.declareBlocks, sourceFragment.declareBlocks...)
	}

	mergedSource.hash = [32]byte(hash.Sum(nil))
	return mergedSource, nil
}

// RawConfigs returns the raw source content used to create Source.
// Do not modify the returned map.
func (s *Source) RawConfigs() map[string][]byte {
	if s == nil {
		return nil
	}
	return s.sourceMap
}

// SHA256 returns the sha256 checksum of the source.
// Do not modify the returned byte array.
func (s *Source) SHA256() [sha256.Size]byte {
	if s == nil {
		return [sha256.Size]byte{}
	}
	return s.hash
}
