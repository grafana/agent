// +build !windows

package main

func IsWindowsService() bool {
	return false
}

func RunService() {
}
