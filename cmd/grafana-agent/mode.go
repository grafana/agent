package main

import (
	"fmt"
	"log"
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
		// Fall back to old EXPERIMENTAL_ENABLE_FLOW flag.
		// TODO: remove support for EXPERIMENTAL_ENABLE_FLOW in v0.32.
		if isFlowEnabled() {
			log.Println("warning: setting EXPERIMENTAL_ENABLE_FLOW is deprecated and will be removed in v0.32, set AGENT_MODE to flow instead")
			return runModeFlow, nil
		}
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

func isFlowEnabled() bool {
	key, found := os.LookupEnv("EXPERIMENTAL_ENABLE_FLOW")
	if !found {
		return false
	}
	return key == "true" || key == "1"
}
