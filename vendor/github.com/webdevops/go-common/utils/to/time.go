package to

import (
	"time"
)

// UnixTime converts time.Time to unix timestamp (float64)
func UnixTime(val time.Time) float64 {
	return float64(val.Unix())
}
