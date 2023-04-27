package vm

import (
	"fmt"
	"strconv"

	"github.com/grafana/agent/pkg/river/internal/value"
)

// stringify tries to turn v into a string.
func stringify(v value.Value) (string, error) {
	switch ty := v.Type(); ty {
	case value.TypeNull:
		return "null", nil

	case value.TypeNumber:
		switch n := v.Number(); n.Kind() {
		case value.NumberKindInt:
			return strconv.FormatInt(n.Int(), 10), nil
		case value.NumberKindUint:
			return strconv.FormatUint(n.Uint(), 10), nil
		case value.NumberKindFloat:
			return strconv.FormatFloat(n.Float(), 'f', -1, 64), nil
		default:
			panic("river/vm: unknown NumberKind value")
		}

	case value.TypeString:
		return v.Text(), nil

	case value.TypeBool:
		if v.Bool() {
			return "true", nil
		} else {
			return "false", nil
		}

	default:
		return "", value.Error{
			Value: v,
			Inner: fmt.Errorf("value of type %s cannot be converted into string", ty),
		}
	}
}
