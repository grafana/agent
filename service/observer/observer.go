// Package observer implements the Observer service for Flow.
package observer

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/encoding/riverjson"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the Observer service.
const ServiceName = "observer"

type Options struct {
	refreshFrequency time.Duration
	remoteEndpoint   string
}

type Observer struct {
	opts Options
}

var _ service.Service = (*Observer)(nil)

// New returns a new, unstarted instance of the HTTP service.
func New(opts Options) *Observer {
	return &Observer{
		opts: opts,
	}
}

// Data implements service.Service.
func (*Observer) Data() any {
	return nil
}

// Definition implements service.Service.
func (*Observer) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: nil, // observer does not accept configuration
		DependsOn:  nil, // observer has no dependencies.
	}
}

// Run implements service.Service.
func (*Observer) Run(ctx context.Context, host service.Host) error {
	components, err := host.ListComponents("", component.InfoOptions{
		GetHealth:    true,
		GetArguments: true,
		GetExports:   true,
		GetDebugInfo: true,
	})
	if err != nil {
		//TODO: Log a warning and continue?
	}

	getAgentState(components)

	<-ctx.Done()
	return nil
}

func getAgentState(components []*component.Info) []Component {
	res := []Component{}

	parentId := uint(0)
	for _, cInfo := range components {
		// ComponentDetail should start at index 0 for each component
		idCounter := parentId + 1

		componentDetail := []riverjson.ComponentDetail{}

		// Add the arguments
		componentDetail = append(componentDetail, getTopLevelComponentDetail("arguments", parentId, &idCounter))
		componentDetail = append(componentDetail, riverjson.GetComponentDetail(cInfo.Arguments, parentId, &idCounter)...)

		// Add the exports
		componentDetail = append(componentDetail, getTopLevelComponentDetail("exports", parentId, &idCounter))
		componentDetail = append(componentDetail, riverjson.GetComponentDetail(cInfo.Exports, parentId, &idCounter)...)

		// Add the debug info
		componentDetail = append(componentDetail, getTopLevelComponentDetail("debug_info", parentId, &idCounter))
		componentDetail = append(componentDetail, riverjson.GetComponentDetail(cInfo.DebugInfo, parentId, &idCounter)...)

		componentState := Component{
			ID:       cInfo.ID.LocalID,
			ModuleID: cInfo.ID.ModuleID,
			Health: Health{
				Health:     cInfo.Health.Health.String(),
				Message:    cInfo.Health.Message,
				UpdateTime: cInfo.Health.UpdateTime,
			},
			ComponentDetail: componentDetail,
		}

		res = append(res, componentState)
	}

	return res
}

func getTopLevelComponentDetail(componentName string, parentId uint, idCounter *uint) riverjson.ComponentDetail {
	res := riverjson.ComponentDetail{
		ID:         *idCounter,
		ParentID:   parentId,
		Name:       componentName,
		Label:      "",
		RiverType:  "",
		RiverValue: []byte{},
	}
	*idCounter += 1
	return res
}

// Update implements service.Service.
func (*Observer) Update(newConfig any) error {
	return fmt.Errorf("Observer service does not support configuration")
}

//-----------------------------------------
// TODO: Remove these structs later
//-----------------------------------------

// Metadata
type AgentState struct {
	ID     string
	Labels map[string]string
}

// RG type
type Component struct {
	ID              string                      `parquet:"id"`
	ModuleID        string                      `parquet:"module_id"`
	Health          Health                      `parquet:"health"`
	ComponentDetail []riverjson.ComponentDetail `parquet:"component_detail"`
}

type Health struct {
	Health     string    `parquet:"state"`
	Message    string    `parquet:"message"`
	UpdateTime time.Time `parquet:"update_time"`
}
