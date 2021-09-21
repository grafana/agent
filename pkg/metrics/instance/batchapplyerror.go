package instance

import "fmt"

// BatchApplyError contains the failed configs with error, successful configurations, and then if there is an error
// that is not tied to a specific config then NonConfigError contains it
type BatchApplyError struct {
	Failed []BatchFailure
	// NonConfigError is used when the error is not with a specific configuration, but likely the grouping
	// or applying the group.
	NonConfigError error
}

// BatchFailure for a given config has the associated error
type BatchFailure struct {
	Err    error
	Config Config
}

// CreateBatchApplyErrorOrNil will create an error if failed has > 0 elements on nonConfigError is not nil, else will
// return nil
func CreateBatchApplyErrorOrNil(failed []BatchFailure, nonConfigError error) error {
	if len(failed) == 0 && nonConfigError == nil {
		return nil
	}
	return &BatchApplyError{
		Failed:         failed,
		NonConfigError: nonConfigError,
	}
}

// FindSuccessfulConfigs will take the list of all configs and compare them against the failed
// return the successful based on name
func FindSuccessfulConfigs(e *BatchApplyError, allConfigs []Config) []Config {
	if e == nil || len(e.Failed) == 0 {
		return allConfigs
	}
	succeeded := make([]Config, 0)
	// Need to get the list of successful configs that were applied
	for _, c := range allConfigs {
		var didFail = false
		for _, failed := range e.Failed {
			if failed.Config.Name == c.Name {
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
	if len(e.Failed) == 0 && e.NonConfigError == nil {
		return ""
	}
	return fmt.Sprintf("%d configs failed to apply. Non config error %s ", len(e.Failed), e.NonConfigError)
}
