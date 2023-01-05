package exportedcomments

import "go/types"

// sortInterface is the [go/types] representation of an interface identical to
// sort.Interface.
var sortInterface = types.NewInterfaceType([]*types.Func{
	// Len() int
	newBasicFuncType(
		"Len",
		nil,
		[]types.BasicKind{types.Int},
		false,
	),

	// Less(i, j int) bool
	newBasicFuncType(
		"Less",
		[]types.BasicKind{types.Int, types.Int},
		[]types.BasicKind{types.Bool},
		false,
	),

	// Swap(i, j int)
	newBasicFuncType(
		"Swap",
		[]types.BasicKind{types.Int, types.Int},
		nil,
		false,
	),
}, nil)

func newBasicFuncType(name string, params []types.BasicKind, returns []types.BasicKind, variadic bool) *types.Func {
	var convParams, convReturns []*types.Var
	for _, param := range params {
		convParams = append(convParams, types.NewVar(0, nil, "", types.Typ[param]))
	}
	for _, returnType := range returns {
		convReturns = append(convReturns, types.NewVar(0, nil, "", types.Typ[returnType]))
	}

	var (
		tupParams  = types.NewTuple(convParams...)
		tupReturns = types.NewTuple(convReturns...)
		sig        = types.NewSignatureType(nil, nil, nil, tupParams, tupReturns, variadic)
	)
	return types.NewFunc(0, nil, name, sig)
}

func newTypeVar(name string, typ types.Type) *types.Var {
	return types.NewVar(0, nil, name, typ)
}
