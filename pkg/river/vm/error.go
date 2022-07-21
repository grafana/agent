package vm

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/internal/value"
	"github.com/grafana/agent/pkg/river/printer"
	"github.com/grafana/agent/pkg/river/token/builder"
)

type ValueError struct {
	// Node where the error originated.
	Node ast.Node

	// Error message (e.g., "foobar should be number, got string")
	Message string

	// Value is the formated value which caused the error. Value may contain
	// formatting characters such as tabs and newline depending on the type of
	// value.
	Value string

	// Literal indicates that the offending value is the Node itself, and not a
	// value found inside the evaluation of Node.
	//
	// Literal can be used to determine whether the Value field needs to be
	// printed. When Literal is true, it indicates that the offending value can
	// be seen by printing the line number of the Node.
	Literal bool
}

// Error returns the short-form error messsage of ve.
func (ve ValueError) Error() string {
	if ve.Node != nil {
		return fmt.Sprintf("%s: %s", ast.StartPos(ve.Node).Position(), ve.Message)
	}
	return ve.Message
}

// convertValueError tries to convert err into a ValueError. err must be an
// error from the river/internal/value package, otherwise err will be returned
// unmodified.
func convertValueError(err error, assoc map[value.Value]ast.Node) error {
	var (
		node    ast.Node
		expr    strings.Builder
		message string
		cause   value.Value

		// Until we find a node, we're not a literal error.
		literal = false
	)

	isValueError := value.WalkError(err, func(err error) {
		var val value.Value

		switch ne := err.(type) {
		case value.Error:
			message = ne.Error()
			val = ne.Value
		case value.TypeError:
			message = fmt.Sprintf("should be %s, got %s", ne.Expected, ne.Value.Type())
			val = ne.Value
		case value.MissingKeyError:
			message = fmt.Sprintf("does not have field named %q", ne.Missing)
			val = ne.Value
		case value.ElementError:
			fmt.Fprintf(&expr, "[%d]", ne.Index)
			val = ne.Value
		case value.FieldError:
			fmt.Fprintf(&expr, ".%s", ne.Field)
			val = ne.Value
		}

		cause = val

		if foundNode, ok := assoc[val]; ok {
			// If we just found a direct node, we can reset the expression buffer so
			// we don't unnecessarily print element and field accesses for we can see
			// directly in the file.
			if literal {
				expr.Reset()
			}

			node = foundNode
			literal = true
		} else {
			literal = false
		}
	})
	if !isValueError {
		return err
	}

	if node != nil {
		var nodeText strings.Builder
		if err := printer.Fprint(&nodeText, node); err != nil {
			// TODO(rfratto): is it OK for this to panic?
			panic(err)
		}

		// Merge the node text with the expression together (which will be relative
		// accesses to the expression).
		message = fmt.Sprintf("%s%s %s", nodeText.String(), expr.String(), message)
	} else {
		message = fmt.Sprintf("%s %s", expr.String(), message)
	}

	// Render the underlying problematic value as a string.
	var valueText string
	if cause != value.Null {
		be := builder.NewExpr()
		be.SetValue(cause)
		valueText = string(be.Bytes())
	}

	return ValueError{
		Node:    node,
		Message: message,
		Value:   valueText,
		Literal: literal,
	}
}
