//go:build windows
// +build windows

package uptime

import (
	"syscall"
	"time"
)

var getTickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

func get() (time.Duration, error) {
	ret, _, err := getTickCount.Call()
	if errno, ok := err.(syscall.Errno); !ok || errno != 0 {
		return time.Duration(0), err
	}
	return time.Duration(ret) * time.Millisecond, nil
}
