package to

// Bool returns bool from a bool pointer
func Bool(val *bool) bool {
	if val != nil {
		return *val
	}
	return false
}

// BoolPtr returns bool pointer from a bool
func BoolPtr(val bool) *bool {
	return &val
}

// BoolString returns string (True/False) from a bool
func BoolString(val bool) string {
	if val {
		return "true"
	}
	return "false"
}
