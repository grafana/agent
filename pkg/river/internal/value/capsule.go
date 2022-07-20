package value

import (
	"fmt"
)

// Capsule is a marker interface for Go values which forces a type to be
// represented as a River capsule. This is useful for types whose underlying
// value is not a capsule, such as:
//
//   // Secret is a secret value. It would normally be a River string since the
//   // underlying Go type is string, but it's a capsule since it implements
//   // the Capsule interface.
//   type Secret string
//
//   func (s Secret) RiverCapsule() {}
//
// Extension interfaces are used to describe additional behaviors for Capsules.
// ConvertibleCapsule allows defining custom conversion rules to convert
// between other Go values.
type Capsule interface {
	RiverCapsule()
}

// ErrNoConversion is returned by implementations of ConvertibleCapsule to
// denote that a custom conversion from or to a specific type is unavailable.
var ErrNoConversion = fmt.Errorf("no custom capsule conversion available")

// ConvertibleCapsule is a Capsule which supports custom conversion rules
// between Go types which are not the same.
//
// ConvertibleCapsule's methods are used even if the other type in the
// conversion is not a capsule; such as converting a bool into a capsule type
// or vice-versa.
type ConvertibleCapsule interface {
	Capsule

	// ConvertFrom should modify the ConvertibleCapsule value based on the value
	// of src.
	//
	// ConvertFrom should return ErrNoConversion if no conversion is available
	// from src.
	ConvertFrom(src interface{}) error

	// ConvertInto should convert its value and store it into dst. dst will be a
	// pointer to a value which ConvertInto is expected to update.
	//
	// ConvertInto should return ErrNoConversion if no conversion into dst is
	// available.
	ConvertInto(dst interface{}) error
}
