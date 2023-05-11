package flow

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/parser"
)

// Source holds the contents of a parsed Flow source.
type Source struct {
	sourceMap map[string][]byte
	hash      [sha256.Size]byte // Hash of all files in sourceMap sorted by name.

	// components holds the list of raw River AST blocks describing components.
	// The Flow controller can interpret them.
	components   []*ast.BlockStmt
	configBlocks []*ast.BlockStmt
}

// ParseSource parses the River contents specified by bb into a Source. name
// should be the name of the source used for reporting errors.
//
// bb must not be modified after passing to ParseSource.
func ParseSource(name string, bb []byte) (*Source, error) {
	node, err := parser.ParseFile(name, bb)
	if err != nil {
		return nil, err
	}

	// Look for predefined non-components blocks (i.e., logging), and store
	// everything else into a list of components.
	//
	// TODO(rfratto): should this code be brought into a helper somewhere? Maybe
	// in ast?
	var (
		components []*ast.BlockStmt
		configs    []*ast.BlockStmt
	)

	for _, stmt := range node.Body {
		switch stmt := stmt.(type) {
		case *ast.AttributeStmt:
			return nil, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt.Name).Position(),
				EndPos:   ast.EndPos(stmt.Name).Position(),
				Message:  "unrecognized attribute " + stmt.Name.Name,
			}

		case *ast.BlockStmt:
			fullName := strings.Join(stmt.Name, ".")
			switch fullName {
			case "logging":
				configs = append(configs, stmt)
			case "tracing":
				configs = append(configs, stmt)
			case "argument":
				configs = append(configs, stmt)
			case "export":
				configs = append(configs, stmt)
			default:
				components = append(components, stmt)
			}

		default:
			return nil, diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(stmt).Position(),
				EndPos:   ast.EndPos(stmt).Position(),
				Message:  fmt.Sprintf("unsupported statement type %T", stmt),
			}
		}
	}

	return &Source{
		components:   components,
		configBlocks: configs,
		sourceMap:    map[string][]byte{name: bb},
		hash:         sha256.Sum256(bb),
	}, nil
}

// ParseSources parses the map of sources and combines them into a single
// Source. sources must not be modified after calling ParseSources.
func ParseSources(sources map[string][]byte) (*Source, error) {
	var (
		sortedSources = sourcesMapToSlice(sources)  // Sorted slice so ParseSources always does the same thing.
		mergedSource  = &Source{sourceMap: sources} // Combined source from all the input content.
		hash          = sha256.New()                // Combined hash of all the sources.
	)

	for _, namedSource := range sortedSources {
		hash.Write(namedSource.Content)

		sourceFragment, err := ParseSource(namedSource.Name, namedSource.Content)
		if err != nil {
			return nil, err
		}

		mergedSource.components = append(mergedSource.components, sourceFragment.components...)
		mergedSource.configBlocks = append(mergedSource.configBlocks, sourceFragment.configBlocks...)
	}

	mergedSource.hash = [32]byte(hash.Sum(nil))
	return mergedSource, nil
}

type namedSource struct {
	Name    string
	Content []byte
}

// sourcesMapToSlice returns a sorted slice of sources from an input map.
func sourcesMapToSlice(in map[string][]byte) []namedSource {
	out := make([]namedSource, 0, len(in))

	for name, bb := range in {
		out = append(out, namedSource{
			Name:    name,
			Content: bb,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

// RawConfigs returns the raw source content used to create Source. Do not
// modify the returned map.
func (s *Source) RawConfigs() map[string][]byte {
	if s == nil {
		return nil
	}
	return s.sourceMap
}

// SHA256 returns the sha256 checksum of the source. Do not modify the returned
// byte array.
func (s *Source) SHA256() [sha256.Size]byte {
	if s == nil {
		return [sha256.Size]byte{}
	}
	return s.hash
}
