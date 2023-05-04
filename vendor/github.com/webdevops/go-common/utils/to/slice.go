package to

type SliceInterface interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string
}

// Slice returns a slice with values from a slice with pointer values
func Slice[N SliceInterface](val []*N) []N {
	ret := make([]N, len(val))
	for rowNum, rowVal := range val {
		if rowVal != nil {
			ret[rowNum] = *rowVal
		} else {
			var rowVal N
			ret[rowNum] = rowVal
		}
	}
	return ret
}

// Slice returns a slice with pointer values from a slice with values
func SlicePtr[N SliceInterface](val []N) []*N {
	ret := make([]*N, len(val))
	for rowNum, rowVal := range val {
		val := rowVal
		ret[rowNum] = &val
	}
	return ret
}
