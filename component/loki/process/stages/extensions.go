package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

const (
	RFC3339Nano         = "RFC3339Nano"
	MaxPartialLinesSize = 100 // MaxPartialLinesSize is the max buffer size to hold partial lines when parsing the CRI stage format.lines.
)

// DockerConfig is an empty struct that is used to enable a pre-defined
// pipeline for decoding entries that are using the Docker logs format.
type DockerConfig struct{}

// CRIConfig is an empty struct that is used to enable a pre-defined pipeline
// for decoding entries that are using the CRI logging format.
type CRIConfig struct{}

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
	partialLines    map[model.Fingerprint]Entry
	maxPartialLines int
	base            *Pipeline
}

// Name implement the Stage interface.
func (c *cri) Name() string {
	return "cri"
}

// Run implements Stage interface
func (c *cri) Run(entry chan Entry) chan Entry {
	entry = c.base.Run(entry)

	in := RunWithSkipOrSendMany(entry, func(e Entry) ([]Entry, bool) {
		fingerprint := e.Labels.Fingerprint()

		// We received partial-line (tag: "P")
		if e.Extracted["flags"] == "P" {
			if len(c.partialLines) > c.maxPartialLines {
				// Merge existing partialLines
				entries := make([]Entry, 0, len(c.partialLines))
				for _, v := range c.partialLines {
					entries = append(entries, v)
				}

				level.Warn(c.base.logger).Log("msg", "cri stage: partial lines upperbound exceeded. merging it to single line", "threshold", MaxPartialLinesSize)

				c.partialLines = make(map[model.Fingerprint]Entry)
				c.partialLines[fingerprint] = e

				return entries, false
			}

			prev, ok := c.partialLines[fingerprint]
			if ok {
				e.Line = strings.Join([]string{prev.Line, e.Line}, "")
			}
			c.partialLines[fingerprint] = e

			return []Entry{e}, true // it's a partial-line so skip it.
		}

		// Now we got full-line (tag: "F").
		// 1. If any old partialLines matches with this full-line stream, merge it
		// 2. Else just return the full line.
		prev, ok := c.partialLines[fingerprint]
		if ok {
			e.Line = strings.Join([]string{prev.Line, e.Line}, "")
			delete(c.partialLines, fingerprint)
		}
		return []Entry{e}, false
	})

	return in
}

// NewCRI creates a predefined pipeline for parsing entries in the CRI log
// format.
func NewCRI(logger log.Logger, registerer prometheus.Registerer) (Stage, error) {
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
		maxPartialLines: MaxPartialLinesSize,
		base:            p,
	}
	c.partialLines = make(map[model.Fingerprint]Entry)
	return &c, nil
}
