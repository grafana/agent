// Package config generates Grafana Agent configuration based on Kubernetes
// resources.
package config

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/assets"
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
	// IntegrationsType genertes a configuration for integrations.
	IntegrationsType
)

// String returns the string form of Type.
func (t Type) String() string {
	switch t {
	case MetricsType:
		return "metrics"
	case LogsType:
		return "logs"
	case IntegrationsType:
		return "integrations"
	default:
		return fmt.Sprintf("unknown (%d)", int(t))
	}
}

//go:embed templates/*
var templates embed.FS

// TODO(rfratto): the "Optional" field of secrets is currently ignored.

// BuildConfig builds an Agent configuration file.
func BuildConfig(h grafana.Hierarchy, ty Type) (string, error) {
	vm, err := createVM(h.Secrets)
	if err != nil {
		return "", err
	}

	bb, err := jsonnetMarshal(h)
	if err != nil {
		return "", err
	}

	vm.TLACode("ctx", string(bb))

	switch ty {
	case MetricsType:
		return vm.EvaluateFile("./agent-metrics.libsonnet")
	case LogsType:
		return vm.EvaluateFile("./agent-logs.libsonnet")
	case IntegrationsType:
		return vm.EvaluateFile("./agent-integrations.libsonnet")
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
		Name:   "unmarshalRawJSON",
		Params: ast.Identifiers{"json"},
		Func: func(i []interface{}) (interface{}, error) {
			text, ok := i[0].(string)
			if !ok {
				return nil, jsonnet.RuntimeError{Msg: "argument must be a string"}
			}

			// The json has been encoded as base64, decode it.
			bb, err := base64.StdEncoding.DecodeString(text)
			if err != nil {
				return nil, jsonnet.RuntimeError{Msg: "failed to decode raw JSON: " + err.Error()}
			}

			var v interface{}
			if err := json.Unmarshal(bb, &v); err != nil {
				return nil, jsonnet.RuntimeError{Msg: err.Error()}
			}
			return v, nil
		},
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
