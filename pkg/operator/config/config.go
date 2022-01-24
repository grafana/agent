// Package config generates Grafana Agent configuration based on Kubernetes
// resources.
package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/fatih/structs"
	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v3"
)

// Type is the type of Agent deployment that a config is being generated
// for.
type Type int

const (
	// MetricsType generates a configuration for metrics.
	MetricsType Type = iota + 1
	// LogsType generates a configuration for logs.
	LogsType
)

// String returns the string form of Type.
func (t Type) String() string {
	switch t {
	case MetricsType:
		return "metrics"
	case LogsType:
		return "logs"
	default:
		return fmt.Sprintf("unknown (%d)", int(t))
	}
}

//go:embed templates/*
var templates embed.FS

// Deployment is a set of resources used for one deployment of the Agent.
type Deployment struct {
	// Agent is the root resource that the deployment represents.
	Agent *grafana.GrafanaAgent
	// Metrics is the set of metrics instances discovered from the root Agent resource.
	Metrics []MetricsInstance
	// Logs is the set of logging instances discovered from the root Agent
	// resource.
	Logs []LogInstance
	// Secrets that can be referenced in the deployment.
	Secrets assets.SecretStore
}

// DeepCopy creates a deep copy of d.
func (d *Deployment) DeepCopy() *Deployment {
	p := make([]MetricsInstance, 0, len(d.Metrics))
	for _, i := range d.Metrics {
		var (
			inst   = i.Instance.DeepCopy()
			sMons  = make([]*prom.ServiceMonitor, 0, len(i.ServiceMonitors))
			pMons  = make([]*prom.PodMonitor, 0, len(i.PodMonitors))
			probes = make([]*prom.Probe, 0, len(i.Probes))
		)

		for _, sMon := range i.ServiceMonitors {
			sMons = append(sMons, sMon.DeepCopy())
		}
		for _, pMon := range i.PodMonitors {
			pMons = append(pMons, pMon.DeepCopy())
		}
		for _, probe := range i.Probes {
			probes = append(probes, probe.DeepCopy())
		}

		p = append(p, MetricsInstance{
			Instance:        inst,
			ServiceMonitors: sMons,
			PodMonitors:     pMons,
			Probes:          probes,
		})
	}

	l := make([]LogInstance, 0, len(d.Logs))
	for _, i := range d.Logs {
		var (
			inst  = i.Instance.DeepCopy()
			pLogs = make([]*grafana.PodLogs, 0, len(i.PodLogs))
		)
		for _, pLog := range i.PodLogs {
			pLogs = append(pLogs, pLog.DeepCopy())
		}
		l = append(l, LogInstance{
			Instance: inst,
			PodLogs:  pLogs,
		})
	}

	return &Deployment{
		Agent:   d.Agent.DeepCopy(),
		Metrics: p,
		Logs:    l,
	}
}

// TODO(rfratto): the "Optional" field of secrets is currently ignored.

// BuildConfig builds an Agent configuration file.
func (d *Deployment) BuildConfig(secrets assets.SecretStore, ty Type) (string, error) {
	vm, err := createVM(secrets)
	if err != nil {
		return "", err
	}

	bb, err := jsonnetMarshal(d)
	if err != nil {
		return "", err
	}

	vm.TLACode("ctx", string(bb))

	switch ty {
	case MetricsType:
		return vm.EvaluateFile("./agent-metrics.libsonnet")
	case LogsType:
		return vm.EvaluateFile("./agent-logs.libsonnet")
	default:
		panic(fmt.Sprintf("unexpected config type %v", ty))
	}
}

func createVM(secrets assets.SecretStore) (*jsonnet.VM, error) {
	vm := jsonnet.MakeVM()
	vm.StringOutput = true

	templatesContents, err := fs.Sub(templates, "templates")
	if err != nil {
		return nil, err
	}

	vm.Importer(NewFSImporter(templatesContents, []string{"./"}))

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "marshalYAML",
		Params: ast.Identifiers{"object"},
		Func: func(i []interface{}) (interface{}, error) {
			bb, err := yaml.Marshal(i[0])
			if err != nil {
				return nil, jsonnet.RuntimeError{Msg: err.Error()}
			}
			return string(bb), nil
		},
	})
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "unmarshalYAML",
		Params: ast.Identifiers{"text"},
		Func:   unmarshalYAML,
	})
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "intoStages",
		Params: ast.Identifiers{"text"},
		Func:   intoStages,
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "trimOptional",
		Params: ast.Identifiers{"value"},
		Func: func(i []interface{}) (interface{}, error) {
			m := i[0].(map[string]interface{})
			trimMap(m)
			return m, nil
		},
	})
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "secretLookup",
		Params: ast.Identifiers{"key"},
		Func: func(i []interface{}) (interface{}, error) {
			if i[0] == nil {
				return nil, nil
			}

			k := assets.Key(i[0].(string))
			val, ok := secrets[k]
			if !ok {
				return nil, jsonnet.RuntimeError{Msg: fmt.Sprintf("key not provided: %s", k)}
			}
			return val, nil
		},
	})
	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "secretPath",
		Params: ast.Identifiers{"key"},
		Func: func(i []interface{}) (interface{}, error) {
			if i[0] == nil {
				return nil, nil
			}

			key := SanitizeLabelName(i[0].(string))
			return path.Join("/var/lib/grafana-agent/secrets", key), nil
		},
	})

	vm.NativeFunction(&jsonnet.NativeFunction{
		Name:   "sanitize",
		Params: ast.Identifiers{"text"},
		Func: func(i []interface{}) (interface{}, error) {
			if len(i) != 1 {
				return nil, jsonnet.RuntimeError{Msg: "inappropriate number of arguments"}
			}
			s, ok := i[0].(string)
			if !ok {
				return nil, jsonnet.RuntimeError{Msg: "text must be a string"}
			}
			return SanitizeLabelName(s), nil
		},
	})

	return vm, nil
}

// jsonnetMarshal marshals a value for passing to Jsonnet. This marshals to a
// JSON representation of the Go value, ignoring all json struct tags. Fields
// must be access as they would from Go, with the exception of embedded fields,
// which must be accessed through the embedded type name (a.Embedded.Field).
func jsonnetMarshal(v interface{}) ([]byte, error) {
	if structs.IsStruct(v) {
		return json.Marshal(structs.Map(v))
	}
	return json.Marshal(v)
}

// MetricsInstance is an instance with a set of associated service monitors,
// pod monitors, and probes, which compose the final configuration of the
// generated Metrics instance.
type MetricsInstance struct {
	Instance        *grafana.MetricsInstance
	ServiceMonitors []*prom.ServiceMonitor
	PodMonitors     []*prom.PodMonitor
	Probes          []*prom.Probe
}

// LogInstance is an instance with a set of associated PodLogs.
type LogInstance struct {
	Instance *grafana.LogsInstance
	PodLogs  []*grafana.PodLogs
}
