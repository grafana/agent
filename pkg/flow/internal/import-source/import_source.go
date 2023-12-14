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
	GIT
)

type ImportSource interface {
	Evaluate(scope *vm.Scope) error
	Run(ctx context.Context) error
	Component() component.Component
	CurrentHealth() component.Health
	DebugInfo() interface{}
	Arguments() component.Arguments
}

func CreateImportSource(sourceType SourceType, managedOpts component.Options, eval *vm.Evaluator) ImportSource {
	switch sourceType {
	case FILE:
		return NewImportFile(managedOpts, eval)
		// add other cases
	}
	// This is a programming error, not a config error so this is ok to panic.
	panic(fmt.Errorf("unsupported source type: %v", sourceType))
}
