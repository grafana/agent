package to

import (
	"strings"
)

// String returns a string from a string pointer
func String(val *string) string {
	if val != nil {
		return *val
	}
	return ""
}

// StringPtr returns pointer from string
func StringPtr(val string) *string {
	return &val
}

// StringLower returns lowercased string from a string pointer
func StringLower(val *string) string {
	if val != nil {
		return strings.ToLower(*val)
	}
	return ""
}

// StringMap returns string map from a map string pointer values
func StringMap(val map[string]*string) (ret map[string]string) {
	ret = make(map[string]string, len(val))
	for rowNum, rowVal := range val {
		if rowVal != nil {
			ret[rowNum] = *rowVal
		} else {
			ret[rowNum] = ""
		}
	}
	return
}

// StringMapPtr returns pointer to a map with string pointers from a string map
func StringMapPtr(val map[string]string) *map[string]*string {
	ret := make(map[string]*string, len(val))
	for rowKey, rowVal := range val {
		ret[rowKey] = StringPtr(rowVal)
	}
	return &ret
}
