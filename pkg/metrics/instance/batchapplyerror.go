package instance

import "fmt"

// BatchApplyError contains the failed configs with error, successful configurations
type BatchApplyError struct {
	Failed []BatchFailure
}

// BatchFailure for a given config has the associated error
type BatchFailure struct {
	Err        error
	ConfigName string
}

// FindSuccessfulConfigs will take the list of all configs and compare them against the failed
// return the successful based on name
func (e *BatchApplyError) FindSuccessfulConfigs(allConfigs []Config) []Config {
	if e == nil || len(e.Failed) == 0 {
		return allConfigs
	}
	var succeeded []Config
	// Need to get the list of successful configs that were applied
	for _, c := range allConfigs {
		var didFail = false
		for _, failed := range e.Failed {
			if failed.ConfigName == c.Name {
				didFail = true
				break
			}

		}
		if didFail {
			continue
		}
		succeeded = append(succeeded, c)
	}

	return succeeded
}

func (e BatchApplyError) Error() string {
	if len(e.Failed) == 0 {
		return ""
	}
	return fmt.Sprintf("%d configs failed to apply", len(e.Failed))
}
