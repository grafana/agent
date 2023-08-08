package metadata

import (
	"fmt"
	"reflect"

	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/discovery"
)

type DataType string

var (
	// DataTypeTargets represents things that need to be scraped. These are used by multiple telemetry signals
	// scraping components and often require special labels, e.g. __path__ label is required for scraping
	// logs from files using loki.source.file component.
	DataTypeTargets = DataType("Targets")

	// DataTypeLokiLogs represent logs in Loki format
	DataTypeLokiLogs = DataType("Loki Logs")

	DataTypeOTELTelemetry     = DataType("OTEL Telemetry")
	DataTypePromMetrics       = DataType("Prometheus Metrics")
	DataTypePyroscopeProfiles = DataType("Pyroscope Profiles")
)

type Metadata struct {
	Accepts []DataType
	Outputs []DataType
}

func (m Metadata) Empty() bool {
	return len(m.Accepts) == 0 && len(m.Outputs) == 0
}

func ForComponent(name string) (Metadata, error) {
	reg, ok := component.Get(name)
	if !ok {
		return Metadata{}, fmt.Errorf("could not find component %q", name)
	}
	return inferMetadata(reg.Args, reg.Exports), nil
}

func inferMetadata(args component.Arguments, exports component.Exports) Metadata {
	var accepts []DataType
	var outputs []DataType

	if exports != nil {
		if hasFieldOfType(exports, reflect.TypeOf(loki.NewLogsReceiver())) {
			accepts = append(accepts, DataTypeLokiLogs)
		}
		if hasFieldOfType(exports, reflect.TypeOf([]discovery.Target{})) {
			outputs = append(outputs, DataTypeTargets)
		}
	}

	if args != nil {
		if hasFieldOfType(args, reflect.TypeOf([]discovery.Target{})) {
			accepts = append(accepts, DataTypeTargets)
		}
		// Components that have e.g. `ForwardsTo []loki.LogsReceiver` arguments, typically output logs
		if hasFieldOfType(args, reflect.TypeOf([]loki.LogsReceiver{})) {
			outputs = append(outputs, DataTypeLokiLogs)
		}
	}

	return Metadata{
		Accepts: accepts,
		Outputs: outputs,
	}
}

func hasFieldOfType(obj interface{}, fieldType reflect.Type) bool {
	objValue := reflect.ValueOf(obj)

	// If the object is a pointer, dereference it
	for objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	// If the object is not a struct or interface, return false
	if objValue.Kind() != reflect.Struct && objValue.Kind() != reflect.Interface {
		return false
	}

	for i := 0; i < objValue.NumField(); i++ {
		fv := objValue.Field(i)
		ft := fv.Type()

		// If the field type matches the given type, return true
		if ft == fieldType {
			return true
		}

		if fv.Kind() == reflect.Interface && fieldType.AssignableTo(ft) {
			return true
		}

		// If the field is a struct, recursively check its fields
		if fv.Kind() == reflect.Struct {
			if hasFieldOfType(fv.Interface(), fieldType) {
				return true
			}
		}
	}

	return false
}
