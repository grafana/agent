package main

import (
	"fmt"
	"os"
)

type runMode int8

const (
	runModeInvalid runMode = iota
	runModeStatic
	runModeFlow
)

func getRunMode() (runMode, error) {
	key, found := os.LookupEnv("AGENT_MODE")
	if !found {
		return runModeStatic, nil
	}

	switch key {
	case "flow":
		return runModeFlow, nil
	case "static", "":
		return runModeStatic, nil
	default:
		return runModeInvalid, fmt.Errorf("unrecognized run mode %q", key)
	}
}
