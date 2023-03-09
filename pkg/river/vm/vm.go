// Package vm provides a River expression evaluator.
package vm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/diag"
	"github.com/grafana/agent/pkg/river/internal/reflectutil"
	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/grafana/agent/pkg/river/internal/stdlib"
	"github.com/grafana/agent/pkg/river/internal/value"
)

// Evaluator evaluates River AST nodes into Go values. Each Evaluator is bound
// to a single AST node. To evaluate the node, call Evaluate.
type Evaluator struct {
	// node for the AST.
	//
	// Each Evaluator is bound to a single node to allow for future performance
	// optimizations, allowing for precomputing and storing the result of
	// anything that is constant.
	node ast.Node
}

// New creates a new Evaluator for the given AST node. The given node must be
// either an *ast.File, *ast.BlockStmt, ast.Body, or assignable to an ast.Expr.
func New(node ast.Node) *Evaluator {
	return &Evaluator{node: node}
}

// Evaluate evaluates the Evaluator's node into a River value and decodes that
// value into the Go value v.
//
// Each call to Evaluate may provide a different scope with new values for
// available variables. If a variable used by the Evaluator's node isn't
// defined in scope or any of the parent scopes, Evaluate will return an error.
func (vm *Evaluator) Evaluate(scope *Scope, v interface{}) (err error) {
	// Track a map that allows us to associate values with ast.Nodes so we can
	// return decorated error messages.
	assoc := make(map[value.Value]ast.Node)

	defer func() {
		if err != nil {
			// Decorate the error on return.
			err = makeDiagnostic(err, assoc)
		}
	}()

	switch node := vm.node.(type) {
	case *ast.BlockStmt, ast.Body:
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("river/vm: expected pointer, got %s", rv.Kind()))
		}
		return vm.evaluateBlockOrBody(scope, assoc, node, rv)
	case *ast.File:
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Pointer {
			panic(fmt.Sprintf("river/vm: expected pointer, got %s", rv.Kind()))
		}
		return vm.evaluateBlockOrBody(scope, assoc, node.Body, rv)
	default:
		//TODO: Why isn't "ast.Expr" another "case" in the switch statement?
		expr, ok := node.(ast.Expr)
		if !ok {
			panic(fmt.Sprintf("river/vm: unexpected value type %T", node))
		}
		val, err := vm.evaluateExpr(scope, assoc, expr)
		if err != nil {
			return err
		}
		return value.Decode(val, v)
	}
}

func (vm *Evaluator) evaluateBlockOrBody(scope *Scope, assoc map[value.Value]ast.Node, node ast.Node, rv reflect.Value) error {
	// TODO(paulin): the errors returned by this function are missing context to
	// be able to print line numbers. We need to return decorated error types.

	// Before decoding the block, we need to temporarily take the address of rv
	// to handle the case of it implementing the unmarshaler interface.
	if rv.CanAddr() {
		rv = rv.Addr()
	}

	if ru, ok := rv.Interface().(value.Unmarshaler); ok {
		return ru.UnmarshalRiver(func(v interface{}) error {
			rv := reflect.ValueOf(v)
			if rv.Kind() != reflect.Pointer {
				panic(fmt.Sprintf("river/vm: expected pointer, got %s", rv.Kind()))
			}
			return vm.evaluateBlockOrBody(scope, assoc, node, rv.Elem())
		})
	}

	// Fully deference rv and allocate pointers as necessary.
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}

	// TODO(rfratto): potentially loosen this restriction and allow decoding into
	// an interface{} or map[string]interface{}.
	if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("river/vm: can only evaluate blocks into structs, got %s", rv.Kind()))
	}

	ti := getCachedTagInfo(rv.Type())

	var stmts ast.Body
	switch node := node.(type) {
	case *ast.BlockStmt:
		// Decode the block label first.
		if err := vm.evaluateBlockLabel(node, ti.Tags, rv); err != nil {
			return err
		}
		stmts = node.Body
	case ast.Body:
		stmts = node
	default:
		panic(fmt.Sprintf("river/vm: unrecognized node type %T", node))
	}

	sd := structDecoder{
		VM:      vm,
		Scope:   scope,
		Assoc:   assoc,
		TagInfo: ti,
	}
	//TODO(ptodev): if Decode() returns an error which is not a diagnostic, we should assume that
	// it was because the error applies to the whole block and we should repackage this error here
	// as a diagnostic with the start and end position of the block.
	return sd.Decode(stmts, rv)
}

func (vm *Evaluator) evaluateBlockLabel(node *ast.BlockStmt, tfs []rivertags.Field, rv reflect.Value) error {
	var (
		labelField rivertags.Field
		foundField bool
	)
	for _, tf := range tfs {
		if tf.Flags&rivertags.FlagLabel != 0 {
			labelField = tf
			foundField = true
			break
		}
	}

	// Check for user errors first.
	//
	// We return parser.Error here to restrict the position of the error to just
	// the name. We might be able to clean this up in the future by extending
	// ValueError to have an explicit position.
	switch {
	case node.Label == "" && foundField: // No user label, but struct expects one
		return diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			StartPos: node.NamePos.Position(),
			EndPos:   node.LCurlyPos.Position(),
			Message:  fmt.Sprintf("block %q requires non-empty label", strings.Join(node.Name, ".")),
		}
	case node.Label != "" && !foundField: // User label, but struct doesn't expect one
		return diag.Diagnostic{
			Severity: diag.SeverityLevelError,
			StartPos: node.NamePos.Position(),
			EndPos:   node.LCurlyPos.Position(),
			Message:  fmt.Sprintf("block %q does not support specifying labels", strings.Join(node.Name, ".")),
		}
	}

	if node.Label == "" {
		// no-op: no labels to set.
		return nil
	}

	var (
		field     = reflectutil.GetOrAlloc(rv, labelField)
		fieldType = field.Type()
	)
	if !reflect.TypeOf(node.Label).AssignableTo(fieldType) {
		// The Label struct field needs to be a string.
		panic(fmt.Sprintf("river/vm: cannot assign block label to non-string type %s", fieldType))
	}
	field.Set(reflect.ValueOf(node.Label))
	return nil
}

// prepareDecodeValue prepares v for decoding. Pointers will be fully
// deferenced until finding a non-pointer value. nil pointers will be
// allocated.
func prepareDecodeValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// TODO: Should diag.Diagnostic be returned as a pointer? Or should we make a Diagnostic interface?
// TODO: For nil diagnostics, should we make a diagn.Null type, like the "var Null = Value{}" in value.go?
func (vm *Evaluator) evaluateExpr(scope *Scope, assoc map[value.Value]ast.Node, expr ast.Expr) (v value.Value, err *diag.Diagnostic) {
	defer func() {
		if v != value.Null {
			assoc[v] = expr
		}
	}()

	switch expr := expr.(type) {
	case *ast.LiteralExpr:
		val, err := valueFromLiteral(expr.Value, expr.Kind)
		if err == nil {
			return val, nil
		} else {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  err.Error(),
			}
		}

	case *ast.BinaryExpr:
		lhs, err := vm.evaluateExpr(scope, assoc, expr.Left)
		if err != nil {
			return value.Null, err
		}
		rhs, err := vm.evaluateExpr(scope, assoc, expr.Right)
		if err != nil {
			return value.Null, err
		}
		val, plainErr := evalBinop(lhs, expr.Kind, rhs)
		if plainErr == nil {
			return val, nil
		} else {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  plainErr.Error(),
			}
		}

	case *ast.ArrayExpr:
		vals := make([]value.Value, len(expr.Elements))
		for i, element := range expr.Elements {
			val, err := vm.evaluateExpr(scope, assoc, element)
			if err != nil {
				return value.Null, err
			}
			vals[i] = val
		}
		return value.Array(vals...), nil

	case *ast.ObjectExpr:
		fields := make(map[string]value.Value, len(expr.Fields))
		for _, field := range expr.Fields {
			val, err := vm.evaluateExpr(scope, assoc, field.Value)
			if err != nil {
				return value.Null, err
			}
			fields[field.Name.Name] = val
		}
		return value.Object(fields), nil

	case *ast.IdentifierExpr:
		val, found := scope.Lookup(expr.Ident.Name)
		if !found {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  fmt.Sprintf("identifier %q does not exist", expr.Ident.Name),
			}
		}
		return value.Encode(val), nil

	case *ast.AccessExpr:
		val, err := vm.evaluateExpr(scope, assoc, expr.Value)
		if err != nil {
			return value.Null, err
		}

		switch val.Type() {
		case value.TypeObject:
			res, ok := val.Key(expr.Name.Name)
			if !ok {
				return value.Null, &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(expr.Name).Position(),
					EndPos:   ast.EndPos(expr.Name).Position(),
					Message:  fmt.Sprintf("field %q does not exist", expr.Name.Name),
				}
			}
			return res, nil
		default:
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr.Name).Position(),
				EndPos:   ast.EndPos(expr.Name).Position(),
				Message:  fmt.Sprintf("cannot access field %q on value of type %s", expr.Name.Name, val.Type()),
			}
		}

	case *ast.IndexExpr:
		val, err := vm.evaluateExpr(scope, assoc, expr.Value)
		if err != nil {
			return value.Null, err
		}
		idx, err := vm.evaluateExpr(scope, assoc, expr.Index)
		if err != nil {
			return value.Null, err
		}

		switch val.Type() {
		case value.TypeArray:
			// Arrays are indexed with a number.
			if idx.Type() != value.TypeNumber {
				return value.Null, &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(expr).Position(),
					EndPos:   ast.EndPos(expr).Position(),
					Message:  fmt.Sprintf("Expected value of type 'number', got value of type '%s'", idx.Type()),
				}
			}
			intIndex := int(idx.Int())

			if intIndex < 0 || intIndex >= val.Len() {
				return value.Null, &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(expr).Position(),
					EndPos:   ast.EndPos(expr).Position(),
					Message:  fmt.Sprintf("index %d is out of range of array with length %d", intIndex, val.Len()),
				}
			}
			return val.Index(intIndex), nil

		case value.TypeObject:
			// Objects are indexed with a string.
			if idx.Type() != value.TypeString {
				return value.Null, &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(expr).Position(),
					EndPos:   ast.EndPos(expr).Position(),
					//TODO: Should we use ' inside error messages? It's not common in the codebase.
					Message: fmt.Sprintf("Expected value of type 'string', got value of type '%s'", idx.Type()),
				}
			}

			field, ok := val.Key(idx.Text())
			if !ok {
				return value.Null, &diag.Diagnostic{
					Severity: diag.SeverityLevelError,
					StartPos: ast.StartPos(expr.Index).Position(),
					EndPos:   ast.EndPos(expr.Index).Position(),
					Message:  fmt.Sprintf("field %q does not exist", idx.Text()),
				}
			}
			return field, nil

		default:
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  fmt.Sprintf("expected object or array, got %s", val.Type()),
			}
		}

	case *ast.ParenExpr:
		return vm.evaluateExpr(scope, assoc, expr.Inner)

	case *ast.UnaryExpr:
		val, err := vm.evaluateExpr(scope, assoc, expr.Value)
		if err != nil {
			return value.Null, err
		}
		val, plainErr := evalUnaryOp(expr.Kind, val)
		if plainErr == nil {
			return val, nil
		} else {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  plainErr.Error(),
			}
		}

	case *ast.CallExpr:
		funcVal, err := vm.evaluateExpr(scope, assoc, expr.Value)
		if err != nil {
			return funcVal, err
		}
		if funcVal.Type() != value.TypeFunction {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  fmt.Sprintf("Expected value of type 'function', got value of type '%s'", funcVal.Type()),
			}
		}

		args := make([]value.Value, len(expr.Args))
		for i := 0; i < len(expr.Args); i++ {
			args[i], err = vm.evaluateExpr(scope, assoc, expr.Args[i])
			if err != nil {
				return value.Null, err
			}
		}
		val, plainErr := funcVal.Call(args...)
		if plainErr == nil {
			return val, nil
		} else {
			return value.Null, &diag.Diagnostic{
				Severity: diag.SeverityLevelError,
				StartPos: ast.StartPos(expr).Position(),
				EndPos:   ast.EndPos(expr).Position(),
				Message:  plainErr.Error(),
			}
		}

	default:
		panic(fmt.Sprintf("river/vm: unexpected ast.Expr type %T", expr))
	}
}

// A Scope exposes a set of variables available to use during evaluation.
type Scope struct {
	// Parent optionally points to a parent Scope containing more variable.
	// Variables defined in children scopes take precedence over variables of the
	// same name found in parent scopes.
	Parent *Scope

	// Variables holds the list of available variable names that can be used when
	// evaluating a node.
	//
	// Values in the Variables map should considered immutable after passed to
	// Evaluate; maps and slices will be copied by reference for performance
	// optimizations.
	Variables map[string]interface{}
}

// Lookup looks up a named identifier from the scope, all of the scope's
// parents, and the stdlib.
func (s *Scope) Lookup(name string) (interface{}, bool) {
	// Traverse the scope first, then fall back to stdlib.
	for s != nil {
		if val, ok := s.Variables[name]; ok {
			return val, true
		}
		s = s.Parent
	}
	if ident, ok := stdlib.Identifiers[name]; ok {
		return ident, true
	}
	return nil, false
}
