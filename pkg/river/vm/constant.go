package vm

import (
	"fmt"
	"strconv"

	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/grafana/agent/pkg/river/token"
)

func valueFromLiteral(lit string, tok token.Token) (value.Value, error) {
	// NOTE(rfratto): this function should never return an error, since the
	// parser only produces valid tokens; it can only fail if a user hand-builds
	// an AST with invalid literals.

	switch tok {
	case token.NULL:
		return value.Null, nil

	case token.NUMBER:
		v, err := strconv.ParseInt(lit, 0, 64)
		if err != nil {
			return value.Null, err
		}
		return value.Int(v), nil

	case token.FLOAT:
		v, err := strconv.ParseFloat(lit, 64)
		if err != nil {
			return value.Null, err
		}
		return value.Float(v), nil

	case token.STRING:
		v, err := strconv.Unquote(lit)
		if err != nil {
			return value.Null, err
		}
		return value.String(v), nil

	case token.BOOL:
		switch lit {
		case "true":
			return value.Bool(true), nil
		case "false":
			return value.Bool(false), nil
		default:
			return value.Null, fmt.Errorf("invalid boolean literal %q", lit)
		}
	default:
		panic(fmt.Sprintf("%v is not a valid token", tok))
	}
}
