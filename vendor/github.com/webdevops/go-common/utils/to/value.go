package to

type PointerInterface interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string | []int | []int8 | []int16 | []int32 | []int64 | []float32 | []float64 | []string
}

// Value return the value from a pointer
func Value[N PointerInterface](val *N) N {
	if val != nil {
		return *val
	}
	var def N
	return def
}

// ValuePtr return the pointer from value
func ValuePtr[N PointerInterface](val N) *N {
	return &val
}
