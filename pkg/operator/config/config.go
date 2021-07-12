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
	"gopkg.in/yaml.v2"
)

//go:embed templates/*
var templates embed.FS

// Deployment is a set of resources used for one deployment of the Agent.
type Deployment struct {
	// Agent is the root resource that the deployment represents.
	Agent *grafana.GrafanaAgent
	// Prometheis is the set of prometheus instances discovered from the root Agent resource.
	Prometheis []PrometheusInstance
}

// DeepCopy creates a deep copy of d.
func (d *Deployment) DeepCopy() *Deployment {
	p := make([]PrometheusInstance, 0, len(d.Prometheis))
	for _, i := range d.Prometheis {
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

		p = append(p, PrometheusInstance{
			Instance:        inst,
			ServiceMonitors: sMons,
			PodMonitors:     pMons,
			Probes:          probes,
		})
	}

	return &Deployment{
		Agent:      d.Agent.DeepCopy(),
		Prometheis: p,
	}
}

// TODO(rfratto): the "Optional" field of secrets is currently ignored.

// BuildConfig builds an Agent configuration file.
func (d *Deployment) BuildConfig(secrets assets.SecretStore) (string, error) {
	vm, err := createVM(secrets)
	if err != nil {
		return "", err
	}

	bb, err := jsonnetMarshal(d)
	if err != nil {
		return "", err
	}

	vm.TLACode("ctx", string(bb))
	return vm.EvaluateFile("./agent.libsonnet")
}

func createVM(secrets assets.SecretStore) (*jsonnet.VM, error) {
	vm := jsonnet.MakeVM()
	vm.StringOutput = true

	templatesContents, err := fs.Sub(templates, "templates")
	if err != nil {
		return nil, err
	}

	vm.Importer(NewFSImporter(templatesContents))

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

// PrometheusInstance is an instance with a set of associated service monitors,
// pod monitors, and probes, which compose the final configuration of the
// generated Prometheus instance.
type PrometheusInstance struct {
	Instance        *grafana.PrometheusInstance
	ServiceMonitors []*prom.ServiceMonitor
	PodMonitors     []*prom.PodMonitor
	Probes          []*prom.Probe
}
