package importsource

import (
	"context"
	"fmt"
	"reflect"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/local/file"
	"github.com/grafana/river/vm"
)

// ImportFile imports a module from a file via the local.file component.
type ImportFile struct {
	fileComponent *file.Component
	arguments     component.Arguments
	managedOpts   component.Options
	eval          *vm.Evaluator
}

var _ ImportSource = (*ImportFile)(nil)

func NewImportFile(managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) *ImportFile {
	opts := managedOpts
	opts.OnStateChange = func(e component.Exports) {
		onContentChange(e.(file.Exports).Content.Value)
	}
	return &ImportFile{
		managedOpts: opts,
		eval:        eval,
	}
}

type importFileConfigBlock struct {
	LocalFileArguments file.Arguments `river:",squash"`
}

// SetToDefault implements river.Defaulter.
func (a *importFileConfigBlock) SetToDefault() {
	a.LocalFileArguments = file.DefaultArguments
}

func (im *ImportFile) Evaluate(scope *vm.Scope) error {
	var arguments importFileConfigBlock
	if err := im.eval.Evaluate(scope, &arguments); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}
	if im.fileComponent == nil {
		var err error
		im.fileComponent, err = file.New(im.managedOpts, arguments.LocalFileArguments)
		if err != nil {
			return fmt.Errorf("creating file component: %w", err)
		}
		im.arguments = arguments
	}

	if reflect.DeepEqual(im.arguments, arguments) {
		return nil
	}

	// Update the existing managed component
	if err := im.fileComponent.Update(arguments); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}
	im.arguments = arguments
	return nil
}

func (im *ImportFile) Run(ctx context.Context) error {
	return im.fileComponent.Run(ctx)
}

// CurrentHealth returns the health of the file component.
func (im *ImportFile) CurrentHealth() component.Health {
	return im.fileComponent.CurrentHealth()
}
