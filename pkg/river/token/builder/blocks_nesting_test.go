package builder_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/parser"
	"github.com/grafana/agent/pkg/river/token/builder"
	"github.com/grafana/agent/pkg/river/vm"
	"github.com/stretchr/testify/require"
)

const (
	defaultNumber      = 123
	otherDefaultNumber = 321
)

/*
This test intends to cover various edge cases when we nest blocks and use defaults.
In order to reduce verbosity of the test case and structures names, we use the following naming convention in this file:
	* Out 		- outer block
	* In 		- inner block/attribute
	* NoDef 	- has no defaults defined
	* WithDef 	- has defaults defined via Defaulter interface
	* MatchDef 	- has defaults defined using Defaulter interface that match (the inner defaults match the outer defaults)
	* ZeroDef 	- had defaults defined that are zero values (empty struct initialization)
	* DiffDef 	- has some defaults values defined that are different between the inner and outer types
	* Blk 		- block
	* Attr 		- attribute
	* Opt 		- optional
	* Str,Ptr,Slice,SlicePtr - struct, pointer to struct, slice of structs, slice of pointers to structs
*/

// testCase defines the test case that will:
// 1. encode the `in` to river string
// 2. compare the above value against the `river` string
// 3. decode the `river` string above to a new struct
// 4. compare the above struct against the `in` struct - checking the invariant of the encoding/decoding
type testCase struct {
	name  string
	in    interface{}
	river string
}

// testFactory is a convenience interface for creating test cases, so that we can define test cases near the
// structures that are involved in each test (as methods).
type testFactory interface{ testCases() []testCase }

// all test factories should be added here, so we can run all their tests
var testFactories = []testFactory{
	&OutZeroDefInStrBlkOptWithDef{},
	&OutMatchDefInStrBlkOptMatchDef{},
	&OutNoDefInStrBlkOptWithDef{},
	&OutDiffDefInStrBlkOptDiffDef{},

	&OutZeroDefInPtrBlkOptWithDef{},
	&OutMatchDefInPtrBlkOptMatchDef{},
	&OutNoDefInPtrBlkOptWithDef{},
	&OutDiffDefInPtrBlkOptDiffDef{},

	&OutZeroDefInPtrBlkWithDef{},
	&OutMatchDefInPtrBlkMatchDef{},
	&OutNoDefInPtrBlkWithDef{},
	&OutDiffDefInPtrBlkDiffDef{},
}

// ========== tests with inner struct ==========

// OutZeroDefInStrBlkOptWithDef - outer with zero value default, inner struct block, optional with a default value
type OutZeroDefInStrBlkOptWithDef struct {
	Inner AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutZeroDefInStrBlkOptWithDef) SetToDefault() {
	*o = OutZeroDefInStrBlkOptWithDef{Inner: AttrWithDefault{}}
}

func (o *OutZeroDefInStrBlkOptWithDef) testCases() []testCase {
	return []testCase{
		{
			name:  "no value set",
			in:    OutZeroDefInStrBlkOptWithDef{},
			river: ``,
		},
		{
			name: "different value set",
			in: OutZeroDefInStrBlkOptWithDef{
				Inner: AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutZeroDefInStrBlkOptWithDef{
				Inner: AttrWithDefault{Number: defaultNumber},
			},
			// NOTE: this is correct, because inner block's defaults will be applied
			river: `
			inner { }
		`,
		},
	}
}

// OutMatchDefInStrBlkOptMatchDef - outer with matching default, inner struct block, optional with a matching default value
type OutMatchDefInStrBlkOptMatchDef struct {
	Inner AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutMatchDefInStrBlkOptMatchDef) SetToDefault() {
	*o = OutMatchDefInStrBlkOptMatchDef{Inner: AttrWithDefault{Number: defaultNumber}}
}

func (o *OutMatchDefInStrBlkOptMatchDef) testCases() []testCase {
	return []testCase{
		{
			name: "no value set",
			in:   OutMatchDefInStrBlkOptMatchDef{},
			river: `
		inner {
			number = 0
		}`,
		},
		{
			name: "different value set",
			in: OutMatchDefInStrBlkOptMatchDef{
				Inner: AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutMatchDefInStrBlkOptMatchDef{
				Inner: AttrWithDefault{Number: defaultNumber},
			},
			river: ``,
		},
	}
}

// OutNoDefInStrBlkOptWithDef - outer without default, inner struct block, optional with a default value
type OutNoDefInStrBlkOptWithDef struct {
	Inner AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutNoDefInStrBlkOptWithDef) testCases() []testCase {
	return []testCase{
		{
			// NOTE: even though the inner block has a default value, it will not be applied because the inner block
			// is a struct, so it will be initialized to zero value as we create a new outer block.
			// See the Defaulter interface docs for more details.
			name:  "no value set",
			in:    OutNoDefInStrBlkOptWithDef{},
			river: "",
		},
		{
			name: "different value set",
			in: OutNoDefInStrBlkOptWithDef{
				Inner: AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			// NOTE: this is correct, because inner block's defaults will be applied to empty body `{ }`
			name: "default value set",
			in: OutNoDefInStrBlkOptWithDef{
				Inner: AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
	}
}

// OutDiffDefInStrBlkOptDiffDef - outer with different default, inner struct block, optional with a different default value
type OutDiffDefInStrBlkOptDiffDef struct {
	Inner AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutDiffDefInStrBlkOptDiffDef) SetToDefault() {
	*o = OutDiffDefInStrBlkOptDiffDef{Inner: AttrWithDefault{Number: otherDefaultNumber}}
}

func (o *OutDiffDefInStrBlkOptDiffDef) testCases() []testCase {
	return []testCase{
		{
			name: "no value set",
			in:   OutDiffDefInStrBlkOptDiffDef{},
			river: `
		inner {
			number = 0
		}`,
		},
		{
			name: "different value set",
			in: OutDiffDefInStrBlkOptDiffDef{
				Inner: AttrWithDefault{Number: 42},
			},
			river: `
			inner {
				number = 42
			}
		`,
		},
		{
			// NOTE: again, when we provide empty body `{ }`, the inner block's defaults will be applied
			name: "inner default value set",
			in: OutDiffDefInStrBlkOptDiffDef{
				Inner: AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
		{
			// NOTE: when we don't provide anything, the outer block's defaults will be applied
			name: "outer default value set",
			in: OutDiffDefInStrBlkOptDiffDef{
				Inner: AttrWithDefault{Number: otherDefaultNumber},
			},
			river: ``,
		},
	}
}

// ========== tests with inner pointer to struct ==========

// OutZeroDefInPtrBlkOptWithDef - outer with zero value default, inner struct pointer, optional with a default value
type OutZeroDefInPtrBlkOptWithDef struct {
	Inner *AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutZeroDefInPtrBlkOptWithDef) SetToDefault() {
	*o = OutZeroDefInPtrBlkOptWithDef{Inner: &AttrWithDefault{}}
}

func (o *OutZeroDefInPtrBlkOptWithDef) testCases() []testCase {
	return []testCase{
		//TODO: invariant violated.
		// The test case is: outer block has zero value default and a pointer to inner block. The inner block has a
		// default value. So the outer block's default set the inner to nil.
		// Seems impossible to encode this case in River, because we would need to somehow explicitly set the inner block
		// to nil? How can we do that?

		//{
		//	name: "nil",
		//	in: OutZeroDefInPtrBlkOptWithDef{
		//		Inner: nil,
		//	},
		//	river: ``,
		//},
		{
			name: "zero value set",
			in: OutZeroDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{},
			},
			river: ``,
		},
		{
			name: "different value set",
			in: OutZeroDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutZeroDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `
			inner { }
		`,
		},
	}
}

// OutMatchDefInPtrBlkOptMatchDef - outer with matching default, inner struct pointer, optional with a matching default value
type OutMatchDefInPtrBlkOptMatchDef struct {
	Inner *AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutMatchDefInPtrBlkOptMatchDef) SetToDefault() {
	*o = OutMatchDefInPtrBlkOptMatchDef{Inner: &AttrWithDefault{Number: defaultNumber}}
}

func (o *OutMatchDefInPtrBlkOptMatchDef) testCases() []testCase {
	return []testCase{
		//TODO: invariant violated - not clear how to explicitly set the inner block to nil in River
		//{
		//	name:  "nil",
		//	in:    OutMatchDefInPtrBlkOptMatchDef{},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutMatchDefInPtrBlkOptMatchDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutMatchDefInPtrBlkOptMatchDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutMatchDefInPtrBlkOptMatchDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: ``,
		},
	}
}

// OutNoDefInPtrBlkOptWithDef - outer without default, inner struct pointer, optional with a default value
type OutNoDefInPtrBlkOptWithDef struct {
	Inner *AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutNoDefInPtrBlkOptWithDef) testCases() []testCase {
	return []testCase{
		{
			name:  "nil",
			in:    OutNoDefInPtrBlkOptWithDef{},
			river: "",
		},
		{
			name: "zero value set",
			in: OutNoDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutNoDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutNoDefInPtrBlkOptWithDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
	}
}

// OutDiffDefInPtrBlkOptDiffDef - outer with different default, inner struct pointer, optional with a different default value
type OutDiffDefInPtrBlkOptDiffDef struct {
	Inner *AttrWithDefault `river:"inner,block,optional"`
}

func (o *OutDiffDefInPtrBlkOptDiffDef) SetToDefault() {
	*o = OutDiffDefInPtrBlkOptDiffDef{Inner: &AttrWithDefault{Number: otherDefaultNumber}}
}

func (o *OutDiffDefInPtrBlkOptDiffDef) testCases() []testCase {
	return []testCase{
		//TODO: invariant violated - not clear how to explicitly set the inner block to nil in River
		//{
		//	name:  "nil",
		//	in:    OutDiffDefInPtrBlkOptDiffDef{},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutDiffDefInPtrBlkOptDiffDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutDiffDefInPtrBlkOptDiffDef{
				Inner: &AttrWithDefault{Number: 42},
			},
			river: `
			inner {
				number = 42
			}
		`,
		},
		{
			name: "inner default value set",
			in: OutDiffDefInPtrBlkOptDiffDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
		{
			name: "outer default value set",
			in: OutDiffDefInPtrBlkOptDiffDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: ``,
		},
	}
}

// ========== tests with inner pointer to struct, but not optional ==========

// OutZeroDefInPtrBlkWithDef - outer with zero value default, inner struct pointer, required with a default value
type OutZeroDefInPtrBlkWithDef struct {
	Inner *AttrWithDefault `river:"inner,block"`
}

func (o *OutZeroDefInPtrBlkWithDef) SetToDefault() {
	*o = OutZeroDefInPtrBlkWithDef{Inner: &AttrWithDefault{}}
}

func (o *OutZeroDefInPtrBlkWithDef) testCases() []testCase {
	return []testCase{
		//{
		//	//TODO: encoded river doesn't decode correctly due to error: missing required block "inner"
		//	// seems like we should have failed on encoding to river?
		//	name: "nil",
		//	in: OutZeroDefInPtrBlkWithDef{
		//		Inner: nil,
		//	},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutZeroDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutZeroDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutZeroDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `
			inner { }
		`,
		},
	}
}

// OutMatchDefInPtrBlkMatchDef - outer with matching default, inner struct pointer, required with a matching default value
type OutMatchDefInPtrBlkMatchDef struct {
	Inner *AttrWithDefault `river:"inner,block"`
}

func (o *OutMatchDefInPtrBlkMatchDef) SetToDefault() {
	*o = OutMatchDefInPtrBlkMatchDef{Inner: &AttrWithDefault{Number: defaultNumber}}
}

func (o *OutMatchDefInPtrBlkMatchDef) testCases() []testCase {
	return []testCase{
		////TODO: encoded river doesn't decode correctly due to error: missing required block "inner"
		//// seems like we should have failed on encoding to river?
		//{
		//	name:  "nil",
		//	in:    OutMatchDefInPtrBlkMatchDef{},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutMatchDefInPtrBlkMatchDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutMatchDefInPtrBlkMatchDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutMatchDefInPtrBlkMatchDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: "inner { }",
		},
	}
}

// OutNoDefInPtrBlkWithDef - outer without default, inner struct pointer, required with a default value
type OutNoDefInPtrBlkWithDef struct {
	Inner *AttrWithDefault `river:"inner,block"`
}

func (o *OutNoDefInPtrBlkWithDef) testCases() []testCase {
	return []testCase{
		////TODO: encoded river doesn't decode correctly due to error: missing required block "inner"
		//// seems like we should have failed on encoding to river?
		//{
		//	name:  "nil",
		//	in:    OutNoDefInPtrBlkWithDef{},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutNoDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutNoDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}
		`,
		},
		{
			name: "default value set",
			in: OutNoDefInPtrBlkWithDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
	}
}

// OutDiffDefInPtrBlkDiffDef - outer with different default, inner struct pointer, required with a different default value
type OutDiffDefInPtrBlkDiffDef struct {
	Inner *AttrWithDefault `river:"inner,block"`
}

func (o *OutDiffDefInPtrBlkDiffDef) SetToDefault() {
	*o = OutDiffDefInPtrBlkDiffDef{Inner: &AttrWithDefault{Number: otherDefaultNumber}}
}

func (o *OutDiffDefInPtrBlkDiffDef) testCases() []testCase {
	return []testCase{
		////TODO: encoded river doesn't decode correctly due to error: missing required block "inner"
		//// seems like we should have failed on encoding to river?
		//{
		//	name:  "nil",
		//	in:    OutDiffDefInPtrBlkDiffDef{},
		//	river: "",
		//},
		{
			name: "zero value set",
			in: OutDiffDefInPtrBlkDiffDef{
				Inner: &AttrWithDefault{},
			},
			river: `
			inner {
				number = 0
			}`,
		},
		{
			name: "different value set",
			in: OutDiffDefInPtrBlkDiffDef{
				Inner: &AttrWithDefault{Number: 42},
			},
			river: `
			inner {
				number = 42
			}
		`,
		},
		{
			name: "inner default value set",
			in: OutDiffDefInPtrBlkDiffDef{
				Inner: &AttrWithDefault{Number: defaultNumber},
			},
			river: `inner { }`,
		},
		{
			name: "outer default value set",
			in: OutDiffDefInPtrBlkDiffDef{
				Inner: &AttrWithDefault{Number: otherDefaultNumber},
			},
			river: `
			inner {
				number = 321
			}`,
		},
	}
}

type AttrWithDefault struct {
	Number int `river:"number,attr,optional"`
}

func (i *AttrWithDefault) SetToDefault() {
	*i = AttrWithDefault{Number: defaultNumber}
}

func TestBlockNesting(t *testing.T) {
	var testCases []testCase
	for _, f := range testFactories {
		testCases = append(testCases, f.testCases()...)
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%T/%s", tc.in, tc.name), func(t *testing.T) {
			f := builder.NewFile()
			f.Body().AppendFrom(tc.in)
			actualRiver := string(f.Bytes())
			fmt.Println("====== ACTUAL ======")
			fmt.Println(actualRiver)
			fmt.Println("====================")
			expected := format(t, tc.river)
			require.Equal(t, expected, actualRiver, "generated river didn't match expected")

			// Now decode the River produced above and make sure it's the same as the input.
			eval := vm.New(parseBlock(t, actualRiver))
			vPtr := reflect.New(reflect.TypeOf(tc.in)).Interface()
			require.NoError(t, eval.Evaluate(nil, vPtr), "river evaluation error")

			actualOut := reflect.ValueOf(vPtr).Elem().Interface()
			require.Equal(t, tc.in, actualOut, "Invariant violated: encoded and then decoded block didn't match the original value")
		})
	}
}

func parseBlock(t *testing.T, input string) *ast.BlockStmt {
	t.Helper()

	input = fmt.Sprintf("test { %s }", input)
	res, err := parser.ParseFile("", []byte(input))
	require.NoError(t, err)
	require.Len(t, res.Body, 1)

	stmt, ok := res.Body[0].(*ast.BlockStmt)
	require.True(t, ok, "Expected stmt to be a ast.BlockStmt, got %T", res.Body[0])
	return stmt
}
