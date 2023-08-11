package bijection

import (
	"fmt"
	"reflect"
)

type StructBijection[A interface{}, B interface{}] struct {
	aToBMap map[Names]unsafeConversionFn
	bToAMap map[Names]unsafeConversionFn
}

// We need to use unsafe conversion function because we cannot store types
// with different generic parameters in the same map.
type unsafeConversionFn func(interface{}, interface{}) error

type Names struct {
	A string
	B string
}

// BindField needs to be a function, because methods cannot have extra generic parameters
// other than the ones defined on the structure.
func BindField[A any, B any, FieldA any, FieldB any](si *StructBijection[A, B], names Names, fBijection Bijection[FieldA, FieldB]) {
	si.ensureInitialised()
	var (
		exampleA      A
		exampleB      B
		exampleFieldA FieldA
		exampleFieldB FieldB
	)
	fieldOfA, ok := reflect.TypeOf(exampleA).FieldByName(names.A)
	if !ok {
		panic(fmt.Sprintf("cannot find field %s in type %T", names.A, exampleA))
	}
	if fieldOfA.Type != reflect.TypeOf(exampleFieldA) {
		panic(fmt.Sprintf("field %s in type %T is not of specified type %T", names.A, exampleA, exampleFieldA))
	}

	fieldOfB, ok := reflect.TypeOf(exampleB).FieldByName(names.B)
	if !ok {
		panic(fmt.Sprintf("cannot find field %s in type %T", names.B, exampleB))
	}
	if fieldOfB.Type != reflect.TypeOf(exampleFieldB) {
		panic(fmt.Sprintf("field %s in type %T is not of specified type %T", names.B, exampleB, exampleFieldB))
	}

	si.aToBMap[names] = func(a interface{}, b interface{}) error {
		typedA, ok := a.(*FieldA)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the A type of bijection %+v", a, fBijection)
		}
		typedB, ok := b.(*FieldB)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the B type of bijection %+v", b, fBijection)
		}
		return fBijection.ConvertAToB(typedA, typedB)
	}

	si.bToAMap[names] = func(b interface{}, a interface{}) error {
		typedA, ok := a.(*FieldA)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the A type of bijection %+v", a, fBijection)
		}
		typedB, ok := b.(*FieldB)
		if !ok {
			return fmt.Errorf("type of %+v is %[1]T, which does not match the B type of bijection %+v", b, fBijection)
		}
		return fBijection.ConvertBToA(typedB, typedA)
	}
}

func (si *StructBijection[A, B]) ensureInitialised() {
	if si.aToBMap == nil {
		si.aToBMap = make(map[Names]unsafeConversionFn)
	}
	if si.bToAMap == nil {
		si.bToAMap = make(map[Names]unsafeConversionFn)
	}
}

func (si *StructBijection[A, B]) ConvertAToB(from *A, to *B) error {
	for names, fn := range si.aToBMap {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(names.A)
		toProp := reflect.ValueOf(to).Elem().FieldByName(names.B)
		err := fn(fromProp.Addr().Interface(), toProp.Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (si *StructBijection[A, B]) ConvertBToA(from *B, to *A) error {
	for names, fn := range si.bToAMap {
		fromProp := reflect.ValueOf(from).Elem().FieldByName(names.B)
		toProp := reflect.ValueOf(to).Elem().FieldByName(names.A)
		err := fn(fromProp.Addr().Interface(), toProp.Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}
