package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/river"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

const (
	RFC3339Nano = "RFC3339Nano"
)

// DockerConfig is an empty struct that is used to enable a pre-defined
// pipeline for decoding entries that are using the Docker logs format.
type DockerConfig struct{}

// CRIConfig is an empty struct that is used to enable a pre-defined pipeline
// for decoding entries that are using the CRI logging format.
type CRIConfig struct {
	MaxPartialLines            int    `river:"max_partial_lines,attr,optional"`
	MaxPartialLineSize         uint64 `river:"max_partial_line_size,attr,optional"`
	MaxPartialLineSizeTruncate bool   `river:"max_partial_line_size_truncate,attr,optional"`
}

var (
	_ river.Defaulter = (*CRIConfig)(nil)
	_ river.Validator = (*CRIConfig)(nil)
)

// DefaultCRIConfig contains the default CRIConfig values.
var DefaultCRIConfig = CRIConfig{
	MaxPartialLines:            100,
	MaxPartialLineSize:         0,
	MaxPartialLineSizeTruncate: false,
}

// SetToDefault implements river.Defaulter.
func (args *CRIConfig) SetToDefault() {
	*args = DefaultCRIConfig
}

// Validate implements river.Validator.
func (args *CRIConfig) Validate() error {
	if args.MaxPartialLines <= 0 {
		return fmt.Errorf("max_partial_lines must be greater than 0")
	}

	return nil
}

// NewDocker creates a predefined pipeline for parsing entries in the Docker
// json log format.
func NewDocker(logger log.Logger, registerer prometheus.Registerer) (Stage, error) {
	stages := []StageConfig{
		{
			JSONConfig: &JSONConfig{
				Expressions: map[string]string{
					"output":    "log",
					"stream":    "stream",
					"timestamp": "time",
				},
			},
		},
		{
			LabelsConfig: &LabelsConfig{
				Values: map[string]*string{"stream": nil},
			},
		},
		{
			TimestampConfig: &TimestampConfig{
				Source: "timestamp",
				Format: RFC3339Nano,
			},
		},
		{
			OutputConfig: &OutputConfig{
				"output",
			},
		},
	}
	return NewPipeline(logger, stages, nil, registerer)
}

type cri struct {
	// bounded buffer for CRI-O Partial logs lines (identified with tag `P` till we reach first `F`)
	partialLines map[model.Fingerprint]Entry
	cfg          CRIConfig
	base         *Pipeline
}

var _ Stage = (*cri)(nil)

// Name implement the Stage interface.
func (c *cri) Name() string {
	return StageTypeCRI
}

// implements Stage interface
func (c *cri) Run(entry chan Entry) chan Entry {
	entry = c.base.Run(entry)

	in := RunWithSkipOrSendMany(entry, func(e Entry) ([]Entry, bool) {
		fingerprint := e.Labels.Fingerprint()

		// We received partial-line (tag: "P")
		if e.Extracted["flags"] == "P" {
			if len(c.partialLines) >= c.cfg.MaxPartialLines {
				// Merge existing partialLines
				entries := make([]Entry, 0, len(c.partialLines))
				for _, v := range c.partialLines {
					entries = append(entries, v)
				}

				level.Warn(c.base.logger).Log("msg", "cri stage: partial lines upperbound exceeded. merging it to single line", "threshold", c.cfg.MaxPartialLines)

				c.partialLines = make(map[model.Fingerprint]Entry, c.cfg.MaxPartialLines)
				c.ensureTruncateIfRequired(&e)
				c.partialLines[fingerprint] = e

				return entries, false
			}

			prev, ok := c.partialLines[fingerprint]
			if ok {
				var builder strings.Builder
				builder.WriteString(prev.Line)
				builder.WriteString(e.Line)
				e.Line = builder.String()
			}
			c.ensureTruncateIfRequired(&e)
			c.partialLines[fingerprint] = e

			return []Entry{e}, true // it's a partial-line so skip it.
		}

		// Now we got full-line (tag: "F").
		// 1. If any old partialLines matches with this full-line stream, merge it
		// 2. Else just return the full line.
		prev, ok := c.partialLines[fingerprint]
		if ok {
			var builder strings.Builder
			builder.WriteString(prev.Line)
			builder.WriteString(e.Line)
			e.Line = builder.String()
			c.ensureTruncateIfRequired(&e)
			delete(c.partialLines, fingerprint)
		}
		return []Entry{e}, false
	})

	return in
}

func (c *cri) ensureTruncateIfRequired(e *Entry) {
	if c.cfg.MaxPartialLineSizeTruncate && len(e.Line) > int(c.cfg.MaxPartialLineSize) {
		e.Line = e.Line[:c.cfg.MaxPartialLineSize]
	}
}

// NewCRI creates a predefined pipeline for parsing entries in the CRI log
// format.
func NewCRI(logger log.Logger, config CRIConfig, registerer prometheus.Registerer) (Stage, error) {
	base := []StageConfig{
		{
			RegexConfig: &RegexConfig{
				Expression: "^(?s)(?P<time>\\S+?) (?P<stream>stdout|stderr) (?P<flags>\\S+?) (?P<content>.*)$",
			},
		},
		{
			LabelsConfig: &LabelsConfig{
				Values: map[string]*string{"stream": nil},
			},
		},
		{
			TimestampConfig: &TimestampConfig{
				Source: "time",
				Format: RFC3339Nano,
			},
		},
		{
			OutputConfig: &OutputConfig{
				"content",
			},
		},
		{
			OutputConfig: &OutputConfig{
				"tags",
			},
		},
	}

	p, err := NewPipeline(logger, base, nil, registerer)
	if err != nil {
		return nil, err
	}

	c := cri{
		cfg:  config,
		base: p,
	}
	c.partialLines = make(map[model.Fingerprint]Entry, c.cfg.MaxPartialLines)
	return &c, nil
}
