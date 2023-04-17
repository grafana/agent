package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/alecthomas/units"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Configuration errors.
var (
	ErrDropStageEmptyConfig   = errors.New("drop stage config must contain at least one of `source`, `expression`, `older_than` or `longer_than`")
	ErrDropStageInvalidConfig = errors.New("drop stage config error, `value` and `expression` cannot both be defined at the same time")
	ErrDropStageInvalidRegex  = errors.New("drop stage regex compilation error")
)

var defaultDropReason = "drop_stage"

var emptyDuration time.Duration

// DropConfig contains the configuration for a dropStage
type DropConfig struct {
	DropReason string           `river:"drop_counter_reason,attr,optional"`
	Source     string           `river:"source,attr,optional"`
	Value      string           `river:"value,attr,optional"`
	Expression string           `river:"expression,attr,optional"`
	OlderThan  time.Duration    `river:"older_than,attr,optional"`
	LongerThan units.Base2Bytes `river:"longer_than,attr,optional"`
	regex      *regexp.Regexp
}

// validateDropConfig validates the DropConfig for the dropStage
func validateDropConfig(cfg *DropConfig) error {
	if cfg.Source == "" && cfg.Expression == "" && cfg.OlderThan == emptyDuration && cfg.LongerThan == 0 {
		return ErrDropStageEmptyConfig
	}
	if cfg.DropReason == "" {
		cfg.DropReason = defaultDropReason
	}
	if cfg.Value != "" && cfg.Expression != "" {
		return ErrDropStageInvalidConfig
	}
	if cfg.Expression != "" {
		expr, err := regexp.Compile(cfg.Expression)
		if err != nil {
			return fmt.Errorf("%v: %w", ErrDropStageInvalidRegex, err)
		}
		cfg.regex = expr
	}
	return nil
}

// newDropStage creates a DropStage from config
func newDropStage(logger log.Logger, config DropConfig, registerer prometheus.Registerer) (Stage, error) {
	err := validateDropConfig(&config)
	if err != nil {
		return nil, err
	}

	return &dropStage{
		logger:    log.With(logger, "component", "stage", "type", "drop"),
		cfg:       &config,
		dropCount: getDropCountMetric(registerer),
	}, nil
}

// dropStage applies Label matchers to determine if the include stages should be run
type dropStage struct {
	logger    log.Logger
	cfg       *DropConfig
	dropCount *prometheus.CounterVec
}

func (m *dropStage) Run(in chan Entry) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		for e := range in {
			if !m.shouldDrop(e) {
				out <- e
				continue
			}
			m.dropCount.WithLabelValues(m.cfg.DropReason).Inc()
		}
	}()
	return out
}

func (m *dropStage) shouldDrop(e Entry) bool {
	// There are many options for dropping a log and if multiple are defined it's treated like an AND condition
	// where all drop conditions must be met to drop the log.
	// Therefore if at any point there is a condition which does not match we can return.
	// The order is what I roughly think would be fastest check to slowest check to try to quit early whenever possible

	if m.cfg.LongerThan != 0 {
		if len(e.Line) > int(m.cfg.LongerThan) {
			// Too long, drop
			level.Debug(m.logger).Log("msg", fmt.Sprintf("line met drop criteria for length %v > %v bytes", len(e.Line), int(m.cfg.LongerThan)))
		} else {
			level.Debug(m.logger).Log("msg", fmt.Sprintf("line will not be dropped, it did not meet criteria for drop length %v is not greater than %v", len(e.Line), int(m.cfg.LongerThan)))
			return false
		}
	}

	if m.cfg.OlderThan != emptyDuration {
		ct := time.Now()
		if e.Timestamp.Before(ct.Add(-m.cfg.OlderThan)) {
			// Too old, drop
			level.Debug(m.logger).Log("msg", fmt.Sprintf("line met drop criteria for age; current time=%v, drop before=%v, log timestamp=%v", ct, ct.Add(-m.cfg.OlderThan), e.Timestamp))
		} else {
			level.Debug(m.logger).Log("msg", fmt.Sprintf("line will not be dropped, it did not meet drop criteria for age; current time=%v, drop before=%v, log timestamp=%v", ct, ct.Add(-m.cfg.OlderThan), e.Timestamp))
			return false
		}
	}

	if m.cfg.Source != "" && m.cfg.Expression == "" {
		if v, ok := e.Extracted[m.cfg.Source]; ok {
			if m.cfg.Value == "" {
				// Found in map, no value set meaning drop if found in map
				level.Debug(m.logger).Log("msg", "line met drop criteria for finding source key in extracted map")
			} else {
				if m.cfg.Value == v {
					// Found in map with value set for drop
					level.Debug(m.logger).Log("msg", "line met drop criteria for finding source key in extracted map with value matching desired drop value")
				} else {
					// Value doesn't match, don't drop
					level.Debug(m.logger).Log("msg", fmt.Sprintf("line will not be dropped, source key was found in extracted map but value '%v' did not match desired value '%v'", v, m.cfg.Value))
					return false
				}
			}
		} else {
			// Not found in extracted map, don't drop
			level.Debug(m.logger).Log("msg", "line will not be dropped, the provided source was not found in the extracted map")
			return false
		}
	}

	if m.cfg.Expression != "" {
		if m.cfg.Source != "" {
			if v, ok := e.Extracted[m.cfg.Source]; ok {
				s, err := getString(v)
				if err != nil {
					level.Debug(m.logger).Log("msg", "Failed to convert extracted map value to string, cannot test regex line will not be dropped.", "err", err, "type", reflect.TypeOf(v))
					return false
				}
				match := m.cfg.regex.FindStringSubmatch(s)
				if match == nil {
					// Not a match to the regex, don't drop
					level.Debug(m.logger).Log("msg", fmt.Sprintf("line will not be dropped, the provided regular expression did not match the value found in the extracted map for source key: %v", m.cfg.Source))
					return false
				}
				// regex match, will be dropped
				level.Debug(m.logger).Log("msg", "line met drop criteria, regex matched the value in the extracted map source key")
			} else {
				// Not found in extracted map, don't drop
				level.Debug(m.logger).Log("msg", "line will not be dropped, the provided source was not found in the extracted map")
				return false
			}
		} else {
			match := m.cfg.regex.FindStringSubmatch(e.Line)
			if match == nil {
				// Not a match to the regex, don't drop
				level.Debug(m.logger).Log("msg", "line will not be dropped, the provided regular expression did not match the log line")
				return false
			}
			level.Debug(m.logger).Log("msg", "line met drop criteria, the provided regular expression matched the log line")
		}
	}

	// Everything matched, drop the line
	level.Debug(m.logger).Log("msg", "all criteria met, line will be dropped")
	return true
}

// Name implements Stage
func (m *dropStage) Name() string {
	return StageTypeDrop
}
