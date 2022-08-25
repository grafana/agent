//go:build !windows
// +build !windows

package initiate

//We must keep this package clear of importing any large packages so that we can initialise the windows service ASAP

// Channel to inform server of service stop request
var ServiceExit = make(chan bool)

// IsWindowsService returns whether the current process is running as a Windows
// Service. On non-Windows platforms, this always returns false.
func IsWindowsService() bool {
	return false
}
