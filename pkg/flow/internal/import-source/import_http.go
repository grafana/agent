package importsource

import (
	"context"
	"fmt"
	"reflect"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/remote/http"
	remote_http "github.com/grafana/agent/component/remote/http"
	"github.com/grafana/river/vm"
)

type ImportHTTP struct {
	managedRemoteHTTP *remote_http.Component
	arguments         component.Arguments
	managedOpts       component.Options
	eval              *vm.Evaluator
}

var _ ImportSource = (*ImportHTTP)(nil)

func NewImportHTTP(managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) *ImportHTTP {
	opts := managedOpts
	opts.OnStateChange = func(e component.Exports) {
		onContentChange(e.(http.Exports).Content.Value)
	}
	return &ImportHTTP{
		managedOpts: opts,
		eval:        eval,
	}
}

type ImportHTTPConfigBlock struct {
	RemoteHTTPArguments remote_http.Arguments `river:",squash"`
}

// SetToDefault implements river.Defaulter.
func (a *ImportHTTPConfigBlock) SetToDefault() {
	a.RemoteHTTPArguments.SetToDefault()
}

func (im *ImportHTTP) Evaluate(scope *vm.Scope) error {
	var arguments ImportHTTPConfigBlock
	if err := im.eval.Evaluate(scope, &arguments); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}
	if im.managedRemoteHTTP == nil {
		var err error
		im.managedRemoteHTTP, err = remote_http.New(im.managedOpts, arguments.RemoteHTTPArguments)
		if err != nil {
			return fmt.Errorf("creating http component: %w", err)
		}
		im.arguments = arguments
	}

	if reflect.DeepEqual(im.arguments, arguments) {
		return nil
	}

	// Update the existing managed component
	if err := im.managedRemoteHTTP.Update(arguments); err != nil {
		return fmt.Errorf("updating component: %w", err)
	}
	return nil
}

func (im *ImportHTTP) Run(ctx context.Context) error {
	return im.managedRemoteHTTP.Run(ctx)
}

func (im *ImportHTTP) Arguments() component.Arguments {
	return im.arguments
}

func (im *ImportHTTP) Component() component.Component {
	return im.managedRemoteHTTP
}

func (im *ImportHTTP) CurrentHealth() component.Health {
	return im.managedRemoteHTTP.CurrentHealth()
}

// DebugInfo() is not implemented by the http component.
func (im *ImportHTTP) DebugInfo() interface{} {
	return nil
}
