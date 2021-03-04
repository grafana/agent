// +build !windows

package main

func IsWindowService() bool {
	return false
}

func RunService() {
}
