package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	type Struct struct {
		IgnoreMe bool

		ReqAttr  string     `river:"req_attr,attr"`
		OptAttr  string     `river:"opt_attr,attr,optional"`
		ReqBlock struct{}   `river:"req_block,block"`
		OptBlock struct{}   `river:"opt_block,block,optional"`
		ReqEnum  []struct{} `river:"req_enum,enum"`
		OptEnum  []struct{} `river:"opt_enum,enum,optional"`
		Label    string     `river:",label"`
	}

	fs := Get(reflect.TypeOf(Struct{}))
	tt := Struct{}

	expect := []Field{
		{[]string{"req_attr"}, []int{1}, FlagAttr, reflect.TypeOf(tt.ReqAttr)},
		{[]string{"opt_attr"}, []int{2}, FlagAttr | FlagOptional, reflect.TypeOf(tt.OptAttr)},
		{[]string{"req_block"}, []int{3}, FlagBlock, reflect.TypeOf(tt.ReqBlock)},
		{[]string{"opt_block"}, []int{4}, FlagBlock | FlagOptional, reflect.TypeOf(tt.OptBlock)},
		{[]string{"req_enum"}, []int{5}, FlagEnum, reflect.TypeOf(tt.ReqEnum)},
		{[]string{"opt_enum"}, []int{6}, FlagEnum | FlagOptional, reflect.TypeOf(tt.OptEnum)},
		{[]string{""}, []int{7}, FlagLabel, reflect.TypeOf(tt.Label)},
	}

	for i, f := range expect {
		found := fs[i]
		equal(t, f, found)
	}
}

func TestEmbedded(t *testing.T) {
	type InnerStruct struct {
		InnerField1 string `river:"inner_field_1,attr"`
		InnerField2 string `river:"inner_field_2,attr"`
	}

	type Struct struct {
		Field1 string `river:"parent_field_1,attr"`
		InnerStruct
		Field2 string `river:"parent_field_2,attr"`
	}
	require.PanicsWithValue(t, "river: anonymous fields not supported schema.Struct.InnerStruct", func() { Get(reflect.TypeOf(Struct{})) })
}

func TestSquash(t *testing.T) {
	type InnerStruct struct {
		InnerField1 string `river:"inner_field_1,attr"`
		InnerField2 string `river:"inner_field_2,attr"`
	}

	type Struct struct {
		Field1 string      `river:"parent_field_1,attr"`
		Inner  InnerStruct `river:",squash"`
		Field2 string      `river:"parent_field_2,attr"`
	}

	type StructWithPointer struct {
		Field1 string       `river:"parent_field_1,attr"`
		Inner  *InnerStruct `river:",squash"`
		Field2 string       `river:"parent_field_2,attr"`
	}

	st := Struct{
		Inner: InnerStruct{},
	}

	expect := []Field{
		{
			Name:  []string{"parent_field_1"},
			Index: []int{0},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Field1),
		},
		{
			Name:  []string{"inner_field_1"},
			Index: []int{1, 0},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Inner.InnerField1),
		},
		{
			Name:  []string{"inner_field_2"},
			Index: []int{1, 1},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Inner.InnerField2),
		},
		{
			Name:  []string{"parent_field_2"},
			Index: []int{2},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Field2),
		},
	}

	structActual := Get(reflect.TypeOf(st))
	for i, f := range expect {
		found := structActual[i]
		equal(t, f, found)
	}

	structPointerActual := Get(reflect.TypeOf(StructWithPointer{}))
	for i, f := range expect {
		found := structPointerActual[i]
		equal(t, f, found)
	}
}

func TestDeepSquash(t *testing.T) {
	type Inner2Struct struct {
		InnerField1 string `river:"inner_field_1,attr"`
		InnerField2 string `river:"inner_field_2,attr"`
	}

	type InnerStruct struct {
		Inner2Struct Inner2Struct `river:",squash"`
	}

	type Struct struct {
		Inner InnerStruct `river:",squash"`
	}

	st := Struct{}
	expect := []Field{
		{
			Name:  []string{"inner_field_1"},
			Index: []int{0, 0, 0},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Inner.Inner2Struct.InnerField1),
		},
		{
			Name:  []string{"inner_field_2"},
			Index: []int{0, 0, 1},
			Flags: FlagAttr,
			Type:  reflect.TypeOf(st.Inner.Inner2Struct.InnerField2),
		},
	}

	structActual := Get(reflect.TypeOf(Struct{}))
	for i, f := range expect {
		found := structActual[i]
		equal(t, f, found)
	}
}

func Test_Get_Panics(t *testing.T) {
	expectPanic := func(t *testing.T, expect string, v interface{}) {
		t.Helper()
		require.PanicsWithValue(t, expect, func() {
			_ = Get(reflect.TypeOf(v))
		})
	}

	t.Run("Tagged fields must be exported", func(t *testing.T) {
		type Struct struct {
			attr string `river:"field,attr"` // nolint:unused //nolint:rivertags
		}
		expect := `river: river tag found on unexported field at schema.Struct.attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Options are required", func(t *testing.T) {
		type Struct struct {
			Attr string `river:"field"` //nolint:rivertags
		}
		expect := `river: field schema.Struct.Attr tag is missing options`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Field names must be unique", func(t *testing.T) {
		type Struct struct {
			Attr  string `river:"field1,attr"`
			Block string `river:"field1,block,optional"` //nolint:rivertags
		}
		expect := `river: field name field1 already used by schema.Struct.Attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Name is required for non-label field", func(t *testing.T) {
		type Struct struct {
			Attr string `river:",attr"` //nolint:rivertags
		}
		expect := `river: non-empty field name required at schema.Struct.Attr`
		expectPanic(t, expect, Struct{})
	})

	t.Run("Only one label field may exist", func(t *testing.T) {
		type Struct struct {
			Label1 string `river:",label"`
			Label2 string `river:",label"`
		}
		expect := `river: label field already used by schema.Struct.Label2`
		expectPanic(t, expect, Struct{})
	})
}

func equal(t *testing.T, a Field, b Field) {
	require.True(t, a.Name[0] == b.Name[0])
	require.True(t, a.Index[0] == b.Index[0])
	require.True(t, a.Flags == b.Flags)
	if a.Type == nil && b.Type != nil {
		require.True(t, false)
	}
	require.True(t, a.Type.Name() == b.Type.Name())
}
