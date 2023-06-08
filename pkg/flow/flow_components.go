package flow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/river/encoding/riverjson"
)

// ComponentID is a globally unique identifier for a component.
type ComponentID struct {
	ModuleID string // Unique ID of the module that the component is running in.
	LocalID  string // Local ID of the component, unique to the module it is running in.
}

// String returns the "<ModuleID>/<LocalID>" string representation of the id.
func (id ComponentID) String() string {
	if id.ModuleID == "" {
		return id.LocalID
	}
	return id.ModuleID + "/" + id.LocalID
}

type ComponentDetail struct {
	// Component is the instance of the component. Component will be nil if the
	// component exists in the graph but has not successfully evaluated yet.
	Component component.Component

	ID    ComponentID // ID of the component.
	Label string      // Component label. Not set for singleton components.

	// References and ReferencedBy are the list of IDs in the same module that
	// this component depends on, or is depended on by, respectively.
	References, ReferencedBy []string

	Registration component.Registration // Component registration.
	Health       component.Health       // Current component health.

	Arguments component.Arguments // Current arguments value of the component.
	Exports   component.Exports   // Current exports value of the component.
	DebugInfo interface{}         // Current debug info of the component.
}

// MarshalJSON returns a JSON representation of cd. The format of the
// representation is not stable and is subject to change.
func (cd *ComponentDetail) MarshalJSON() ([]byte, error) {
	type (
		componentHealthJSON struct {
			State       string    `json:"state"`
			Message     string    `json:"message"`
			UpdatedTime time.Time `json:"updatedTime"`
		}

		componentDetailJSON struct {
			Name         string               `json:"name,omitempty"`
			Type         string               `json:"type,omitempty"`
			ID           string               `json:"id,omitempty"`
			Label        string               `json:"label,omitempty"`
			References   []string             `json:"referencesTo"`
			ReferencedBy []string             `json:"referencedBy"`
			Health       *componentHealthJSON `json:"health"`
			Original     string               `json:"original"`
			Arguments    json.RawMessage      `json:"arguments,omitempty"`
			Exports      json.RawMessage      `json:"exports,omitempty"`
			DebugInfo    json.RawMessage      `json:"debugInfo,omitempty"`
		}
	)

	var (
		references   = cd.References
		referencedBy = cd.ReferencedBy

		arguments, exports, debugInfo json.RawMessage
		err                           error
	)

	if references == nil {
		references = []string{}
	}
	if referencedBy == nil {
		referencedBy = []string{}
	}

	arguments, err = riverjson.MarshalBody(cd.Arguments)
	if err != nil {
		return nil, err
	}
	exports, err = riverjson.MarshalBody(cd.Exports)
	if err != nil {
		return nil, err
	}
	debugInfo, err = riverjson.MarshalBody(cd.DebugInfo)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&componentDetailJSON{
		Name:         cd.Registration.Name,
		Type:         "block",
		ID:           cd.ID.LocalID, // TODO(rfratto): support getting component from module.
		Label:        cd.Label,
		References:   references,
		ReferencedBy: referencedBy,
		Health: &componentHealthJSON{
			State:       cd.Health.Health.String(),
			Message:     cd.Health.Message,
			UpdatedTime: cd.Health.UpdateTime,
		},
		Arguments: arguments,
		Exports:   exports,
		DebugInfo: debugInfo,
	})
}

// ComponentDetailOptions is used by [Flow.ListComponents] and
// [Flow.GetComponent] to determine how much information
type ComponentDetailOptions struct {
	GetHealth    bool // When true, sets the Health field of returned components.
	GetArguments bool // When true, sets the Arguments field of returned components.
	GetExports   bool // When true, sets the Exports field of returned components.
	GetDebugInfo bool // When true, sets the DebugInfo field of returned components.
}

// GetComponent returns the detail on an individual component by its global ID.
// The provided opts configures how much detail to return; see
// [ComponentDetailOptions] for more information.
//
// GetComponent returns an error if a component is not found.
//
// BUG(rfratto): The ModuleID in the component ID is ignored.
func (f *Flow) GetComponent(id ComponentID, opts ComponentDetailOptions) (*ComponentDetail, error) {
	f.loadMut.RLock()
	defer f.loadMut.RUnlock()

	// TODO(rfratto): navigate running modules and return a component within a
	// module.

	graph := f.loader.OriginalGraph()

	node := graph.GetByID(id.LocalID)
	if node == nil {
		return nil, fmt.Errorf("component %q does not exist", id)
	}

	cn, ok := node.(*controller.ComponentNode)
	if !ok {
		return nil, fmt.Errorf("%q is not a component", id)
	}

	return f.getComponentDetail(cn, graph, opts), nil
}

// ListComponents returns the list of active components. The provided opts
// configures what components and to what detail they are returned; see
// [ComponentDetailOptions] for more information.
func (f *Flow) ListComponents(opts ComponentDetailOptions) []*ComponentDetail {
	f.loadMut.RLock()
	defer f.loadMut.RUnlock()

	// TODO(rfratto): support returning components inside modules.

	var (
		components = f.loader.Components()
		graph      = f.loader.OriginalGraph()
	)

	detail := make([]*ComponentDetail, len(components))
	for i, component := range components {
		detail[i] = f.getComponentDetail(component, graph, opts)
	}
	return detail
}

func (f *Flow) getComponentDetail(cn *controller.ComponentNode, graph *dag.Graph, opts ComponentDetailOptions) *ComponentDetail {
	var references, referencedBy []string

	// Skip over any edge which isn't between two component nodes. This is a
	// temporary workaround needed until there's athe concept of configuration
	// blocks in the API.
	//
	// Without this change, the graph fails to render when a configuration
	// block is referenced in the graph.
	//
	// TODO(rfratto): add support for config block nodes in the API and UI.
	for _, dep := range graph.Dependencies(cn) {
		if _, ok := dep.(*controller.ComponentNode); ok {
			references = append(references, dep.NodeID())
		}
	}
	for _, dep := range graph.Dependants(cn) {
		if _, ok := dep.(*controller.ComponentNode); ok {
			referencedBy = append(referencedBy, dep.NodeID())
		}
	}

	// Fields which are optional to set.
	var (
		health    component.Health
		arguments component.Arguments
		exports   component.Exports
		debugInfo interface{}
	)

	if opts.GetHealth {
		health = cn.CurrentHealth()
	}
	if opts.GetArguments {
		arguments = cn.Arguments()
	}
	if opts.GetExports {
		exports = cn.Exports()
	}
	if opts.GetDebugInfo {
		debugInfo = cn.DebugInfo()
	}

	return &ComponentDetail{
		Component: cn.Component(),

		ID: ComponentID{
			ModuleID: f.opts.ControllerID,
			LocalID:  cn.NodeID(),
		},
		Label: cn.Label(),

		References:   references,
		ReferencedBy: referencedBy,

		Registration: cn.Registration(),
		Health:       health,

		Arguments: arguments,
		Exports:   exports,
		DebugInfo: debugInfo,
	}
}
