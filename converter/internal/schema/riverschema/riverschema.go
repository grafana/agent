// Package riverschema describes types to be able to write type-safe River in
// Go.
//
// Types in this package map to River types. Functions called New<Type> create
// a new value of that type, while Expr<Type> functions define an expression
// which is expected to resolve to that type at runtime.
//
// Some types only have a Expr<Type> function when defining New<Type> is not
// possible.
//
// All types implement [builder].Tokenizer, allowing them to marshal to River
// correctly.
package riverschema

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/grafana/agent/pkg/river/token"
	"github.com/grafana/agent/pkg/river/token/builder"
)

// Type represents any of the types in this package.
type Type interface {
	builder.Tokenizer
	riverType()
}

// Collection of types which map to River types.
type (
	// UnsignedNumber represents a River number backed by an unsigned Go integer.
	UnsignedNumber struct {
		value uint64
		expr  string
	}

	// SignedNumber represents a River number backed by a signed Go integer.
	SignedNumber struct {
		value int64
		expr  string
	}

	// FloatNumber represents a River number backed by a Go floating point value.
	FloatNumber struct {
		value float64
		expr  string
	}

	// String represents a River string backed by a Go string.
	String struct {
		value string
		expr  string
	}

	// Bool represents a River boolean backed by a Go boolean.
	Bool struct {
		value bool
		expr  string
	}

	// Array represents a River array backed by a Go slice.
	Array[T Type] struct {
		elems []T
		expr  string
	}

	// Object represents a River object backed by a Go map.
	Object[T Type] struct {
		kvps map[string]T
		expr string
	}

	// Function represents a River function mapped from an expression.
	Function struct{ expr string }

	// Capsule represents a River capsule mapped from an expression.
	Capsule struct{ expr string }
)

// NewUnsignedNumber creates a new unsigned number.
func NewUnsignedNumber(value uint64) *UnsignedNumber { return &UnsignedNumber{value: value} }

// ExprUnsignedNumber creates a new unsigned number represented by an
// expression.
func ExprUnsignedNumber(expr string) *UnsignedNumber { return &UnsignedNumber{expr: expr} }

// NewSignedNumber creates a new signed number.
func NewSignedNumber(value int64) *SignedNumber { return &SignedNumber{value: value} }

// ExprSignedNumber creates a new signed number represented by an expression.
func ExprSignedNumber(expr string) *SignedNumber { return &SignedNumber{expr: expr} }

// NewFloatNumber creates a new floating point number.
func NewFloatNumber(value float64) *FloatNumber { return &FloatNumber{value: value} }

// ExprFloatNumber creates a new floating point number represented by an
// expression.
func ExprFloatNumber(expr string) *FloatNumber { return &FloatNumber{expr: expr} }

// NewString creates a new string.
func NewString(value string) *String { return &String{value: value} }

// ExprString creates a new string represented by an expression.
func ExprString(expr string) *String { return &String{expr: expr} }

// NewBool creates a new boolean.
func NewBool(value bool) *Bool { return &Bool{value: value} }

// ExprBool creates a new boolean represented by an expression.
func ExprBool(expr string) *Bool { return &Bool{expr: expr} }

// NewArray creates a new boolean.
func NewArray[T Type](value []T) *Array[T] { return &Array[T]{elems: value} }

// ExprArray creates a new boolean represented by an expression.
func ExprArray[T Type](expr string) *Array[T] { return &Array[T]{expr: expr} }

// NewObject creates a new boolean.
func NewObject[T Type](value map[string]T) *Object[T] { return &Object[T]{kvps: value} }

// ExprObject creates a new boolean represented by an expression.
func ExprObject[T Type](expr string) *Object[T] { return &Object[T]{expr: expr} }

// ExprFunction creates a new function represented by an expression.
func ExprFunction(expr string) *Function { return &Function{expr: expr} }

// ExprCapsule creates a new function represented by an expression.
func ExprCapsule(expr string) *Capsule { return &Capsule{expr: expr} }

//
// Type implementations
//

var (
	_ Type = (*UnsignedNumber)(nil)
	_ Type = (*SignedNumber)(nil)
	_ Type = (*FloatNumber)(nil)
	_ Type = (*String)(nil)
	_ Type = (*Bool)(nil)
	_ Type = (*Array[Type])(nil)
	_ Type = (*Object[Type])(nil)
	_ Type = (*Function)(nil)
	_ Type = (*Capsule)(nil)
)

func (value *UnsignedNumber) riverType() {}
func (value *SignedNumber) riverType()   {}
func (value *FloatNumber) riverType()    {}
func (value *String) riverType()         {}
func (value *Bool) riverType()           {}
func (value *Array[T]) riverType()       {}
func (value *Object[T]) riverType()      {}
func (value *Function) riverType()       {}
func (value *Capsule) riverType()        {}

//
// NOTE(rfratto): A bug in River forces the function implementations below to
// be implemented as non-pointer receivers. Once that bug is fixed, we should
// change these back to pointer receivers for consistency with the rest of the
// methods.
//

// RiverTokenize tokenizes the UnsignedNumber.
func (value UnsignedNumber) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{{
		Tok: token.NUMBER,
		Lit: strconv.FormatUint(value.value, 10),
	}}
}

// RiverTokenize tokenizes the SignedNumber.
func (value SignedNumber) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{{
		Tok: token.NUMBER,
		Lit: strconv.FormatInt(value.value, 10),
	}}
}

// RiverTokenize tokenizes the FloatNumber.
func (value FloatNumber) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{{
		Tok: token.NUMBER,
		Lit: strconv.FormatFloat(value.value, 'f', -1, 64),
	}}
}

// RiverTokenize tokenizes the String.
func (value String) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{
		{Tok: token.STRING, Lit: `"` + value.value + `"`},
	}
}

// RiverTokenize tokenizes the Bool.
func (value Bool) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{
		{Tok: token.BOOL, Lit: fmt.Sprintf("%t", value.value)},
	}
}

// RiverTokenize tokenizes the Array.
func (value Array[T]) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}

	var toks []builder.Token

	toks = append(toks, builder.Token{Tok: token.LBRACK})

	for i := 0; i < len(value.elems); i++ {
		toks = append(toks, value.elems[i].RiverTokenize()...)

		if i+1 < len(value.elems) {
			toks = append(toks, builder.Token{Tok: token.COMMA})
		}
	}

	toks = append(toks, builder.Token{Tok: token.RBRACK})
	return toks
}

// RiverTokenize tokenizes the Object.
func (value Object[T]) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}

	var keys []string
	for key := range value.kvps {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var toks []builder.Token

	toks = append(toks, builder.Token{Tok: token.LCURLY})
	if len(keys) > 0 {
		toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
	}

	for _, key := range keys {
		toks = append(toks,
			builder.Token{Tok: token.STRING, Lit: `"` + key + `"`},
			builder.Token{Tok: token.ASSIGN},
		)
		toks = append(toks, value.kvps[key].RiverTokenize()...)
		toks = append(toks, builder.Token{Tok: token.COMMA})
		toks = append(toks, builder.Token{Tok: token.LITERAL, Lit: "\n"})
	}

	toks = append(toks, builder.Token{Tok: token.RCURLY})
	return toks
}

// RiverTokenize tokenizes the Function.
func (value Function) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{{Tok: token.NULL}}
}

// RiverTokenize tokenizes the Capsule.
func (value Capsule) RiverTokenize() []builder.Token {
	if value.expr != "" {
		return []builder.Token{{
			Tok: token.LITERAL,
			Lit: value.expr,
		}}
	}
	return []builder.Token{{Tok: token.NULL}}
}
