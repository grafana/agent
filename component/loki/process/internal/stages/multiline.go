package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/common/model"
)

// Configuration errors.
var (
	ErrMultilineStageEmptyConfig        = errors.New("multiline stage config must define `firstline` regular expression")
	ErrMultilineStageInvalidRegex       = errors.New("multiline stage first line regex compilation error")
	ErrMultilineStageInvalidMaxWaitTime = errors.New("multiline stage `max_wait_time` parse error")
)

// MultilineConfig contains the configuration for a Multiline stage.
type MultilineConfig struct {
	Expression  string        `river:"firstline,attr"`
	MaxLines    uint64        `river:"max_lines,attr,optional"`
	MaxWaitTime time.Duration `river:"max_wait_time,attr,optional"`
	regex       *regexp.Regexp
}

// DefaultMultilineConfig applies the default values on
var DefaultMultilineConfig = MultilineConfig{
	MaxLines:    128,
	MaxWaitTime: 3 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (args *MultilineConfig) SetToDefault() {
	*args = DefaultMultilineConfig
}

// Validate implements river.Validator.
func (args *MultilineConfig) Validate() error {
	if args.MaxWaitTime <= 0 {
		return fmt.Errorf("max_wait_time must be greater than 0")
	}

	return nil
}

func validateMultilineConfig(cfg *MultilineConfig) error {
	if cfg.Expression == "" {
		return ErrMultilineStageEmptyConfig
	}

	expr, err := regexp.Compile(cfg.Expression)
	if err != nil {
		return fmt.Errorf("%v: %w", ErrMultilineStageInvalidRegex, err)
	}
	cfg.regex = expr

	return nil
}

// multilineStage matches lines to determine whether the following lines belong to a block and should be collapsed
type multilineStage struct {
	logger log.Logger
	cfg    MultilineConfig
}

// multilineState captures the internal state of a running multiline stage.
type multilineState struct {
	buffer         *bytes.Buffer // The lines of the current multiline block.
	startLineEntry Entry         // The entry of the start line of a multiline block.
	currentLines   uint64        // The number of lines of the current multiline block.
}

// newMulitlineStage creates a MulitlineStage from config
func newMultilineStage(logger log.Logger, config MultilineConfig) (Stage, error) {
	err := validateMultilineConfig(&config)
	if err != nil {
		return nil, err
	}

	return &multilineStage{
		logger: log.With(logger, "component", "stage", "type", "multiline"),
		cfg:    config,
	}, nil
}

func (m *multilineStage) Run(in chan Entry) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)

		streams := make(map[model.Fingerprint](chan Entry))
		wg := new(sync.WaitGroup)

		for e := range in {
			key := e.Labels.FastFingerprint()
			s, ok := streams[key]
			if !ok {
				// Pass through entries until we hit first start line.
				if !m.cfg.regex.MatchString(e.Line) {
					level.Debug(m.logger).Log("msg", "pass through entry", "stream", key)
					out <- e
					continue
				}

				level.Debug(m.logger).Log("msg", "creating new stream", "stream", key)
				s = make(chan Entry)
				streams[key] = s

				wg.Add(1)
				go m.runMultiline(s, out, wg)
			}
			level.Debug(m.logger).Log("msg", "pass entry", "stream", key, "line", e.Line)
			s <- e
		}

		// Close all streams and wait for them to finish being processed.
		for _, s := range streams {
			close(s)
		}
		wg.Wait()
	}()
	return out
}

func (m *multilineStage) runMultiline(in chan Entry, out chan Entry, wg *sync.WaitGroup) {
	defer wg.Done()

	state := &multilineState{
		buffer:       new(bytes.Buffer),
		currentLines: 0,
	}

	for {
		select {
		case <-time.After(m.cfg.MaxWaitTime):
			level.Debug(m.logger).Log("msg", fmt.Sprintf("flush multiline block due to %v timeout", m.cfg.MaxWaitTime), "block", state.buffer.String())
			m.flush(out, state)
		case e, ok := <-in:
			level.Debug(m.logger).Log("msg", "processing line", "line", e.Line, "stream", e.Labels.FastFingerprint())

			if !ok {
				level.Debug(m.logger).Log("msg", "flush multiline block because inbound closed", "block", state.buffer.String(), "stream", e.Labels.FastFingerprint())
				m.flush(out, state)
				return
			}

			isFirstLine := m.cfg.regex.MatchString(e.Line)
			if isFirstLine {
				level.Debug(m.logger).Log("msg", "flush multiline block because new start line", "block", state.buffer.String(), "stream", e.Labels.FastFingerprint())
				m.flush(out, state)

				// The start line entry is used to set timestamp and labels in the flush method.
				// The timestamps for following lines are ignored for now.
				state.startLineEntry = e
			}

			// Append block line
			if state.buffer.Len() > 0 {
				state.buffer.WriteRune('\n')
			}
			state.buffer.WriteString(e.Line)
			state.currentLines++

			if state.currentLines == m.cfg.MaxLines {
				m.flush(out, state)
			}
		}
	}
}

func (m *multilineStage) flush(out chan Entry, s *multilineState) {
	if s.buffer.Len() == 0 {
		level.Debug(m.logger).Log("msg", "nothing to flush", "buffer_len", s.buffer.Len())
		return
	}
	// copy extracted data.
	extracted := make(map[string]interface{}, len(s.startLineEntry.Extracted))
	for k, v := range s.startLineEntry.Extracted {
		extracted[k] = v
	}
	collapsed := Entry{
		Extracted: extracted,
		Entry: loki.Entry{
			Labels: s.startLineEntry.Entry.Labels.Clone(),
			Entry: logproto.Entry{
				Timestamp: s.startLineEntry.Entry.Entry.Timestamp,
				Line:      s.buffer.String(),
			},
		},
	}
	s.buffer.Reset()
	s.currentLines = 0

	out <- collapsed
}

// Name implements Stage
func (m *multilineStage) Name() string {
	return StageTypeMultiline
}
