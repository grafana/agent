//go:build linux
// +build linux

package perf

import (
	"io/ioutil"
	"strconv"
)

const (
	PMUEventBaseDir = "/sys/bus/event_source/devices"
)

// AvailablePMUs returns a mapping of available PMUs from
// /sys/bus/event_sources/devices to the PMU event type (number).
func AvailablePMUs() (map[string]int, error) {
	pmus := make(map[string]int)
	pmuTypes, err := ioutil.ReadDir(PMUEventBaseDir)
	if err != nil {
		return nil, err
	}
	for _, pmuFileInfo := range pmuTypes {
		pmu := pmuFileInfo.Name()
		pmuEventStr, err := fileToStrings(PMUEventBaseDir + "/" + pmu + "/type")
		if err != nil {
			return nil, err
		}
		pmuEvent, err := strconv.Atoi(pmuEventStr[0])
		if err != nil {
			return nil, err
		}
		pmus[pmu] = pmuEvent
	}
	return pmus, nil
}
