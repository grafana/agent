package importsource

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/vm"
)

type SourceType int

const (
	File SourceType = iota
	String
)

const (
	BlockImportFile   = "import.file"
	BlockImportString = "import.string"
)

// ImportSource is a generic representation of a component that retrieves a module.
type ImportSource interface {
	// Evaluate updates the arguments for the managed component.
	Evaluate(scope *vm.Scope) error
	// Run the managed component.
	Run(ctx context.Context) error
}

// NewImportSource creates a new ImportSource depending on the type.
// onContentChange is used by the source when it receives new content.
func NewImportSource(sourceType SourceType, managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) ImportSource {
	switch sourceType {
	case File:
		return NewImportFile(managedOpts, eval, onContentChange)
	case String:
		return NewImportString(eval, onContentChange)
	}
	panic(fmt.Errorf("unsupported source type: %v", sourceType))
}

// GetSourceType returns a SourceType matching a source name.
func GetSourceType(fullName string) SourceType {
	switch fullName {
	case BlockImportFile:
		return File
	case BlockImportString:
		return String
	}
	panic(fmt.Errorf("name does not map to a know source type: %v", fullName))
}
