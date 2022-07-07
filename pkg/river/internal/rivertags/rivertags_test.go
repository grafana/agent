package rivertags_test

import (
	"reflect"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/rivertags"
	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	type Struct struct {
		IgnoreMe bool

		ReqAttr  string   `rvr:"req_attr,attr"`
		OptAttr  string   `rvr:"opt_attr,attr,optional"`
		ReqBlock struct{} `rvr:"req_block,block"`
		OptBlock struct{} `rvr:"opt_block,block,optional"`
		Label    string   `rvr:",label"`
	}

	fs := rivertags.Get(reflect.TypeOf(Struct{}))

	expect := []rivertags.Field{
		{"req_attr", []int{1}, rivertags.FlagAttr},
		{"opt_attr", []int{2}, rivertags.FlagAttr | rivertags.FlagOptional},
		{"req_block", []int{3}, rivertags.FlagBlock},
		{"opt_block", []int{4}, rivertags.FlagBlock | rivertags.FlagOptional},
		{"", []int{5}, rivertags.FlagLabel},
	}

	require.Equal(t, expect, fs)
}

func Test_Get_Embedded(t *testing.T) {
	type InnerStruct struct {
		InnerField1 string `rvr:"inner_field_1,attr"`
		InnerField2 string `rvr:"inner_field_2,attr"`
	}

	type Struct struct {
		Field1 string `rvr:"parent_field_1,attr"`
		InnerStruct
		Field2 string `rvr:"parent_field_2,attr"`
	}

	fs := rivertags.Get(reflect.TypeOf(Struct{}))

	expect := []rivertags.Field{
		{"parent_field_1", []int{0}, rivertags.FlagAttr},
		{"inner_field_1", []int{1, 0}, rivertags.FlagAttr},
		{"inner_field_2", []int{1, 1}, rivertags.FlagAttr},
		{"parent_field_2", []int{2}, rivertags.FlagAttr},
	}

	require.Equal(t, expect, fs)
}

func Test_Get_Panics(t *testing.T) {
	expectPanic := func(t *testing.T, expect string, v interface{}) {
		t.Helper()
		require.PanicsWithValue(t, expect, func() {
			_ = rivertags.Get(reflect.TypeOf(v))
		})
	}

	t.Run("Tagged fields must be exported", func(t *testing.T) {
		type Struct struct {
			attr string `rvr:"field,attr"` // nolint:unused
		}
		expect := `river: rvr tag found on unexported field at rivertags_test.Struct.attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Options are required", func(t *testing.T) {
		type Struct struct {
			Attr string `rvr:"field"`
		}
		expect := `river: field rivertags_test.Struct.Attr tag is missing options`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Field names must be unique", func(t *testing.T) {
		type Struct struct {
			Attr  string `rvr:"field1,attr"`
			Block string `rvr:"field1,block,optional"`
		}
		expect := `river: field name field1 already used by rivertags_test.Struct.Attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Name is required for non-label field", func(t *testing.T) {
		type Struct struct {
			Attr string `rvr:",attr"`
		}
		expect := `river: non-empty field name required at rivertags_test.Struct.Attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Only one label field may exist", func(t *testing.T) {
		type Struct struct {
			Label1 string `rvr:",label"`
			Label2 string `rvr:",label"`
		}
		expect := `river: label field already used by rivertags_test.Struct.Label2`
		expectPanic(t, expect, Struct{})
	})
}
