package importsource

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/river/vm"
)

type SourceType int

// TODO: add other import sources

const (
	File SourceType = iota
)

const (
	BlockImportFile = "import.file"
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
	case File:
		return NewImportFile(managedOpts, eval, onContentChange)
	}
	// This is a programming error, not a config error so this is ok to panic.
	panic(fmt.Errorf("unsupported source type: %v", sourceType))
}

func GetSourceType(fullName string) SourceType {
	switch fullName {
	case BlockImportFile:
		return File
	}
	// This is a programming error, not a config error so this is ok to panic.
	panic(fmt.Errorf("name does not map to a know source type: %v", fullName))
}
