package to

type NumberInterface interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

// Number returns number from an number pointer (generic)
func Number[N NumberInterface](val *N) N {
	if val != nil {
		return *val
	}
	return 0
}

// NumberPtr returns number pointer from number (generic)
func NumberPtr[N NumberInterface](val N) *N {
	return &val
}

// Int returns int from an int pointer
func Int(val *int) int {
	return Number(val)
}

// IntPtr returns int pointer from int value
func IntPtr(val int) *int {
	return &val
}

// Int32 returns int32 from an int32 pointer
func Int32(val *int32) int32 {
	return Number(val)
}

// Int32Ptr returns int32 pointer from int32
func Int32Ptr(val int32) *int32 {
	return &val
}

// Int64 returns int64 from an int64 pointer
func Int64(val *int64) int64 {
	return Number(val)
}

// Int64Ptr returns int64 pointer from int64 value
func Int64Ptr(val int64) *int64 {
	return &val
}

// Float32 returns float32 from a float32 pointer
func Float32(val *float32) float32 {
	return Number(val)
}

// Float32Ptr returns float32 ptr from float32 value
func Float32Ptr(val float32) *float32 {
	return &val
}

// Float64 returns float64 from a float64 pointer
func Float64(val *float64) float64 {
	return Number(val)
}

// Float64Ptr returns float64 pointer from float64 value
func Float64Ptr(val float64) *float64 {
	return &val
}
