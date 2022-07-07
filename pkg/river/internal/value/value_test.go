package value_test

import (
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/river/internal/value"
)

func BenchmarkValue(b *testing.B) {
	b.Run("Null", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.Null
		}
	})

	b.Run("Uint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.Uint(1234)
		}
	})

	b.Run("Int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.Int(1234)
		}
	})

	b.Run("Float", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.Float(1234.5678)
		}
	})

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.String("foobar")
		}
	})

	b.Run("Bool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = value.Bool(true)
		}
	})

	b.Run("Object (empty)", func(b *testing.B) {
		vals := getObjectMap(b, 0)
		for i := 0; i < b.N; i++ {
			_ = value.Object(vals)
		}
	})

	b.Run("Object (10 elements)", func(b *testing.B) {
		vals := getObjectMap(b, 10)
		for i := 0; i < b.N; i++ {
			_ = value.Object(vals)
		}
	})

	b.Run("Object (100 elements)", func(b *testing.B) {
		vals := getObjectMap(b, 100)
		for i := 0; i < b.N; i++ {
			_ = value.Object(vals)
		}
	})

	b.Run("Array (empty)", func(b *testing.B) {
		vals := getArrayVals(b, 0)
		for i := 0; i < b.N; i++ {
			_ = value.Array(vals...)
		}
	})

	b.Run("Array (10 elements)", func(b *testing.B) {
		vals := getArrayVals(b, 10)
		for i := 0; i < b.N; i++ {
			_ = value.Array(vals...)
		}
	})

	b.Run("Array (100 elements)", func(b *testing.B) {
		vals := getArrayVals(b, 100)
		for i := 0; i < b.N; i++ {
			_ = value.Array(vals...)
		}
	})

	b.Run("Func", func(b *testing.B) {
		f := func() int { return 15 }
		for i := 0; i < b.N; i++ {
			_ = value.Func(f)
		}
	})

	b.Run("Capsule", func(b *testing.B) {
		ch := make(chan int)
		for i := 0; i < b.N; i++ {
			_ = value.Capsule(ch)
		}
	})

	b.Run("Value.Type", func(b *testing.B) {
		val := value.Int(1234)
		for i := 0; i < b.N; i++ {
			_ = val.Type()
		}
	})

	b.Run("Value.Bool", func(b *testing.B) {
		val := value.Bool(true)
		for i := 0; i < b.N; i++ {
			_ = val.Bool()
		}
	})

	b.Run("Value.Int", func(b *testing.B) {
		tt := []struct {
			name string
			val  value.Value
		}{
			{"From Int", value.Int(1234)},
			{"From Uint", value.Uint(1234)},
			{"From Float", value.Float(1234.5678)},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = tc.val.Int()
				}
			})
		}
	})

	b.Run("Value.Uint", func(b *testing.B) {
		tt := []struct {
			name string
			val  value.Value
		}{
			{"From Int", value.Int(1234)},
			{"From Uint", value.Uint(1234)},
			{"From Float", value.Float(1234.5678)},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = tc.val.Uint()
				}
			})
		}
	})

	b.Run("Value.Float", func(b *testing.B) {
		tt := []struct {
			name string
			val  value.Value
		}{
			{"From Int", value.Int(1234)},
			{"From Uint", value.Uint(1234)},
			{"From Float", value.Float(1234.5678)},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = tc.val.Float()
				}
			})
		}
	})

	b.Run("Value.Len", func(b *testing.B) {
		tt := []struct {
			name string
			val  value.Value
		}{
			{"From Array", value.Array(getArrayVals(b, 10)...)},
			{"From Object", value.Object(getObjectMap(b, 10))},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = tc.val.Len()
				}
			})
		}
	})

	b.Run("Value.Index", func(b *testing.B) {
		obj := value.Array(getArrayVals(b, 10)...)
		for i := 0; i < b.N; i++ {
			_ = obj.Index(5)
		}
	})

	b.Run("Value.Keys", func(b *testing.B) {
		type Person struct {
			Name string `rvr:"name,key"`
		}

		var (
			structVal = Person{Name: "John"}
			mapVal    = map[string]interface{}{"name": "John"}
		)

		tt := []struct {
			name string
			val  value.Value
		}{
			{"From struct", value.Encode(structVal)},
			{"From map", value.Encode(mapVal)},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = tc.val.Keys()
				}
			})
		}
	})

	b.Run("Value.Key", func(b *testing.B) {
		type Person struct {
			Name string `rvr:"name,key"`
		}

		var (
			structVal = Person{Name: "John"}
			mapVal    = map[string]interface{}{"name": "John"}
		)

		tt := []struct {
			name string
			val  value.Value
		}{
			{"From struct", value.Encode(structVal)},
			{"From map", value.Encode(mapVal)},
		}

		for _, tc := range tt {
			b.Run(tc.name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, _ = tc.val.Key("name")
				}
			})
		}
	})
}

func getObjectMap(b *testing.B, elems int) map[string]value.Value {
	b.StopTimer()
	defer b.StartTimer()

	res := make(map[string]value.Value)
	for i := 0; i < elems; i++ {
		res[fmt.Sprintf("field_%d", i)] = value.Null
	}
	return res
}

func getArrayVals(b *testing.B, elems int) []value.Value {
	b.StopTimer()
	defer b.StartTimer()

	res := make([]value.Value, elems)
	for i := 0; i < elems; i++ {
		res[i] = value.Null
	}
	return res
}
