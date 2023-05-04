package uptime

import "time"

// Get uptime duration
func Get() (time.Duration, error) {
	return get()
}
