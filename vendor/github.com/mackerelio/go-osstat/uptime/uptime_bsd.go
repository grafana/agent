//go:build darwin || freebsd || netbsd || openbsd
// +build darwin freebsd netbsd openbsd

package uptime

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

func get() (time.Duration, error) {
	out, err := unix.SysctlRaw("kern.boottime")
	if err != nil {
		return 0, err
	}
	var timeval syscall.Timeval
	if len(out) != int(unsafe.Sizeof(timeval)) {
		return 0, fmt.Errorf("unexpected output of sysctl kern.boottime: %v (len: %d)", out, len(out))
	}
	timeval = *(*syscall.Timeval)(unsafe.Pointer(&out[0]))
	sec, nsec := timeval.Unix()
	return time.Since(time.Unix(sec, nsec)), nil
}
