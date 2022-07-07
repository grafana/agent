package value

import (
	"math"
	"reflect"
	"strconv"
)

var (
	nativeIntBits  = reflect.TypeOf(int(0)).Bits()
	nativeUintBits = reflect.TypeOf(uint(0)).Bits()
)

// numberKind categorizes a type of Go number.
type numberKind uint8

const (
	numberKindInt numberKind = iota
	numberKindUint
	numberKindFloat
)

func makeNumberKind(k reflect.Kind) numberKind {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return numberKindInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return numberKindUint
	case reflect.Float32, reflect.Float64:
		return numberKindFloat
	default:
		panic("river/value: makeNumberKind called with unsupported Kind value")
	}
}

// numberValue is a generic representation of Go numbers. It is intended to be
// created on the fly for numerical operations when the real number type is not
// known.
type numberValue struct {
	// Value holds the raw data for the number. Note that for numberKindFloat,
	// value is the raw bits of the float64 and must be converted back to a
	// float64 before it can be used.
	value uint64

	bits uint8      // 8, 16, 32, 64
	k    numberKind // int, uint, float
}

func newNumberValue(v reflect.Value) numberValue {
	var (
		val  uint64
		bits int
		nk   numberKind
	)

	switch v.Kind() {
	case reflect.Int:
		val, bits, nk = uint64(v.Int()), nativeIntBits, numberKindInt
	case reflect.Int8:
		val, bits, nk = uint64(v.Int()), 8, numberKindInt
	case reflect.Int16:
		val, bits, nk = uint64(v.Int()), 16, numberKindInt
	case reflect.Int32:
		val, bits, nk = uint64(v.Int()), 32, numberKindInt
	case reflect.Int64:
		val, bits, nk = uint64(v.Int()), 64, numberKindInt
	case reflect.Uint:
		val, bits, nk = v.Uint(), nativeUintBits, numberKindUint
	case reflect.Uint8:
		val, bits, nk = v.Uint(), 8, numberKindUint
	case reflect.Uint16:
		val, bits, nk = v.Uint(), 16, numberKindUint
	case reflect.Uint32:
		val, bits, nk = v.Uint(), 32, numberKindUint
	case reflect.Uint64:
		val, bits, nk = v.Uint(), 64, numberKindUint
	case reflect.Float32:
		val, bits, nk = math.Float64bits(v.Float()), 32, numberKindFloat
	case reflect.Float64:
		val, bits, nk = math.Float64bits(v.Float()), 64, numberKindFloat
	default:
		panic("river/value: unrecognized Go number type " + v.Kind().String())
	}

	return numberValue{val, uint8(bits), nk}
}

func (nv numberValue) Int() int64 {
	if nv.k == numberKindFloat {
		return int64(math.Float64frombits(nv.value))
	}
	return int64(nv.value)
}

func (nv numberValue) Uint() uint64 {
	if nv.k == numberKindFloat {
		return uint64(math.Float64frombits(nv.value))
	}
	return nv.value
}

func (nv numberValue) Float() float64 {
	if nv.k == numberKindFloat {
		return math.Float64frombits(nv.value)
	}
	return float64(nv.value)
}

// ToString converts the number to a string.
func (nv numberValue) ToString() string {
	switch nv.k {
	case numberKindUint:
		return strconv.FormatUint(nv.value, 10)
	case numberKindInt:
		return strconv.FormatInt(int64(nv.value), 10)
	case numberKindFloat:
		return strconv.FormatFloat(math.Float64frombits(nv.value), 'f', -1, 64)
	}
	panic("river/value: unreachable")
}
