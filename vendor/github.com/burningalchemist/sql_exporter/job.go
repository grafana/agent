package sql_exporter

import (
	"fmt"

	"github.com/burningalchemist/sql_exporter/config"
	"github.com/burningalchemist/sql_exporter/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// Job is a collection of targets with the same collectors applied.
type Job interface {
	Targets() []Target
}

// job implements Job. It wraps the corresponding JobConfig and a set of Targets.
type job struct {
	config     *config.JobConfig
	targets    []Target
	logContext string
}

// NewJob returns a new Job with the given configuration.
func NewJob(jc *config.JobConfig, gc *config.GlobalConfig) (Job, errors.WithContext) {
	j := job{
		config:     jc,
		targets:    make([]Target, 0, 10),
		logContext: fmt.Sprintf("job=%q", jc.Name),
	}

	for _, sc := range jc.StaticConfigs {
		for tname, dsn := range sc.Targets {
			constLabels := prometheus.Labels{
				"job":      jc.Name,
				"instance": tname,
			}
			for name, value := range sc.Labels {
				// Shouldn't happen as there are sanity checks in config, but check nonetheless.
				if _, found := constLabels[name]; found {
					return nil, errors.Errorf(j.logContext, "duplicate label %q", name)
				}
				constLabels[name] = value
			}
			t, err := NewTarget(j.logContext, tname, string(dsn), jc.Collectors(), constLabels, gc)
			if err != nil {
				return nil, err
			}
			j.targets = append(j.targets, t)
		}
	}

	return &j, nil
}

func (j *job) Targets() []Target {
	return j.targets
}
