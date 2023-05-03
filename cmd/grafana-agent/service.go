//go:build !windows

package main

// IsWindowsService returns whether the current process is running as a Windows
// Service. On non-Windows platforms, this always returns false.
func IsWindowsService() bool {
	return false
}

// RunService runs the current process as a Windows service. On non-Windows platforms,
// this is always a no-op.
func RunService() error {
	return nil
}
