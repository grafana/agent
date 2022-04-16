package component

import (
	"encoding"
	"fmt"
	"time"
)

// HealthComponent is an optional extension interface which Components that
// export health information may implement.
type HealthComponent interface {
	Component

	// CurrentHealth returns the current Health status for the component.
	// CurrentHealth may be overridden by Flow if there is a higher-level issue
	// with the component.
	CurrentHealth() Health
}

// StatusComponent is an optional extension interface that Components that
// export debug status information may implement.
type StatusComponent interface {
	Component

	// CurrentStatus returns the current status of the component. May return nil
	// if there is no status to report.
	CurrentStatus() any
}

type Health struct {
	Health     HealthType `hcl:"state,attr"`
	Message    string     `hcl:"message,optional"`
	UpdateTime time.Time  `hcl:"update_time,optional"`
}

// HealthType is the specific type of health for a component.
type HealthType uint8

var _ encoding.TextMarshaler = HealthType(0)
var _ encoding.TextUnmarshaler = (*HealthType)(nil)

const (
	HealthTypeUnkown HealthType = iota
	HealthTypeRunning
	HealthTypeHealthy
	HealthTypeUnhealthy
	HealthTypeExited
)

// String returns the string representation of ht.
func (ht HealthType) String() string {
	switch ht {
	case HealthTypeRunning:
		return "running"
	case HealthTypeHealthy:
		return "health"
	case HealthTypeUnhealthy:
		return "unhealthy"
	case HealthTypeExited:
		return "exited"
	default:
		return "unknown"
	}
}

// MarshalText implements encoding.TextMarshaler.
func (ht HealthType) MarshalText() (text []byte, err error) {
	return []byte(ht.String()), nil
}

// UnmarshalText implments encoding.TextUnmarshaler.
func (ht *HealthType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "running":
		*ht = HealthTypeRunning
	case "healthy":
		*ht = HealthTypeHealthy
	case "unhealthy":
		*ht = HealthTypeUnhealthy
	case "unknown":
		*ht = HealthTypeUnkown
	case "exited":
		*ht = HealthTypeExited
	default:
		return fmt.Errorf("invalid health type %q", string(text))
	}
	return nil
}
