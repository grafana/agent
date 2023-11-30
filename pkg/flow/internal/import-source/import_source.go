package importsource

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/vm"
)

type SourceType int

const (
	FILE SourceType = iota
	HTTP
	GIT
)

const (
	BlockImportFile = "import.file"
	BlockImportHTTP = "import.http"
	BlockImportGit  = "import.git"
)

type ImportSource interface {
	Evaluate(scope *vm.Scope) error
	Run(ctx context.Context) error
	Component() component.Component
	CurrentHealth() component.Health
	DebugInfo() interface{}
	Arguments() component.Arguments
}

func NewImportSource(sourceType SourceType, managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) ImportSource {
	switch sourceType {
	case FILE:
		return NewImportFile(managedOpts, eval, onContentChange)
	case HTTP:
		return NewImportHTTP(managedOpts, eval, onContentChange)
	case GIT:
		return NewImportGit(managedOpts, eval, onContentChange)
	}
	// This is a programming error, not a config error so this is ok to panic.
	panic(fmt.Errorf("unsupported source type: %v", sourceType))
}

func GetSourceType(fullName string) SourceType {
	switch fullName {
	case BlockImportFile:
		return FILE
	case BlockImportGit:
		return GIT
	case BlockImportHTTP:
		return HTTP
	}
	// This is a programming error, not a config error so this is ok to panic.
	panic(fmt.Errorf("name does not map to a know source type: %v", fullName))
}
