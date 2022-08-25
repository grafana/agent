//go:build !windows
// +build !windows

package initiate

// IsWindowsService returns whether the current process is running as a Windows
// Service. On non-Windows platforms, this always returns false.
func IsWindowsService() bool {
	return false
}
