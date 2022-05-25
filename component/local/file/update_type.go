package file

import (
	"encoding"
	"fmt"
)

// UpdateType is used to specify how changes to the file should be detected.
type UpdateType int

const (
	// UpdateTypeInvalid indicates an invalid UpdateType.
	UpdateTypeInvalid UpdateType = iota
	// UpdateTypeWatch uses filesystem events to wait for changes to the file.
	UpdateTypeWatch
	// UpdateTypePoll will re-read the file on an interval to detect changes.
	UpdateTypePoll

	// UpdateTypeDefault holds the default UpdateType.
	UpdateTypeDefault = UpdateTypeWatch
)

var (
	_ encoding.TextMarshaler   = UpdateType(0)
	_ encoding.TextUnmarshaler = (*UpdateType)(nil)
)

// String returns the string representation of the UpdateType.
func (ut UpdateType) String() string {
	switch ut {
	case UpdateTypeWatch:
		return "watch"
	case UpdateTypePoll:
		return "poll"
	default:
		return fmt.Sprintf("UpdateType(%d)", ut)
	}
}

// MarshalText implements encoding.TextMarshaler.
func (ut UpdateType) MarshalText() (text []byte, err error) {
	return []byte(ut.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ut *UpdateType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "":
		*ut = UpdateTypeDefault
	case "watch":
		*ut = UpdateTypeWatch
	case "poll":
		*ut = UpdateTypePoll
	default:
		return fmt.Errorf("unrecognized update type %q", string(text))
	}
	return nil
}
