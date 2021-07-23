// Package crow implements a correctness checker tool similar to Loki Canary.
// Inspired by Cortex test-exporter.
package crow

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	promapi "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

// Config for the Crow metrics checker.
type Config struct {
	PrometheusAddr string // Base URL of Prometheus server
	NumSamples     int    // Number of samples to generate
	UserID         string // User ID to use when querying.
	ExtraSelectors string // Extra selectors for queries, i.e., cluster="prod"

	// Querying Params

	QueryTimeout  time.Duration // Timeout for querying
	QueryDuration time.Duration // Time before and after sample to search
	QueryStep     time.Duration // Step between samples in search

	// Validation Params

	MaxValidations    int           // Maximum amount of times to search for a sample
	MaxTimestampDelta time.Duration // Maximum timestamp delta to use for validating.
	ValueEpsilon      float64       // Maximum epsilon to use for validating.
}

// RegisterFlags registers flags for the config to the given FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.PrometheusAddr, "prometheus-addr", DefaultConfig.PrometheusAddr, "Root URL of the Prometheus API to query against")
	f.IntVar(&c.NumSamples, "generate-samples", DefaultConfig.NumSamples, "Number of samples to generate when being scraped")
	f.StringVar(&c.UserID, "user-id", DefaultConfig.UserID, "UserID to attach to query. Useful for querying multi-tenated Cortex.")
	f.StringVar(&c.ExtraSelectors, "extra-selectors", DefaultConfig.ExtraSelectors, "Extra selectors to include in queries, useful for identifying different instances of this job.")

	f.DurationVar(&c.QueryTimeout, "query-timeout", DefaultConfig.QueryTimeout, "timeout for querying")
	f.DurationVar(&c.QueryDuration, "query-duration", DefaultConfig.QueryDuration, "time before and after sample to search")
	f.DurationVar(&c.QueryStep, "query-step", DefaultConfig.QueryStep, "step between samples when searching")

	f.IntVar(&c.MaxValidations, "max-validations", DefaultConfig.MaxValidations, "Maximum number of times to try validating a sample")
	f.DurationVar(&c.MaxTimestampDelta, "max-timestamp-delta", DefaultConfig.MaxTimestampDelta, "maximum difference from the stored timestamp from the validating sample to allow")
	f.Float64Var(&c.ValueEpsilon, "sample-epsilon", DefaultConfig.ValueEpsilon, "maximum difference from the stored value from the validating sample to allow")
}

// DefaultConfig holds defaults for Crow settings.
var DefaultConfig = Config{
	MaxValidations: 5,
	NumSamples:     10,

	QueryTimeout:  150 * time.Millisecond,
	QueryDuration: 2 * time.Second,
	QueryStep:     100 * time.Millisecond,

	// MaxTimestampDelta is set to 750ms to allow some buffer for a slow network
	// before the scrape goes through.
	MaxTimestampDelta: 750 * time.Millisecond,
	ValueEpsilon:      0.0001,
}

// Crow is a collectness checker that validates scraped metrics reach a
// Prometheus-compatible server with the same values and roughly the same
// timestamp.
//
// Crow exposes two sets of metrics:
//
// 1. Test metrics, where each scrape generates a validation job.
// 2. State metrics, exposing state of the Crow checker itself.
//
// These two metrics should be exposed via different endpoints, and only state
// metrics are safe to be manually collecetd from.
//
// Collecting from the set of test metrics generates a validation job, where
// Crow will query the Prometheus API to ensure the metrics that were scraped
// were written with (approximately) the same timestamp as the scrape time
// and with (approximately) the same floatnig point values exposed in the
// scrape.
//
// If a set of test metrics were not found and retries have been exhausted,
// or if the metrics were found but the values did not match, the error
// counter will increase.
type Crow struct {
	cfg Config
	m   *metrics

	promClient promapi.API

	wg   sync.WaitGroup
	quit chan struct{}

	pendingMtx sync.Mutex
	pending    []*sample
	sampleCh   chan []*sample
}

// New creates a new Crow.
func New(cfg Config) (*Crow, error) {
	c, err := newCrow(cfg)
	if err != nil {
		return nil, err
	}

	c.wg.Add(1)
	go c.runLoop()
	return c, nil
}

func newCrow(cfg Config) (*Crow, error) {
	cli, err := api.NewClient(api.Config{
		Address: cfg.PrometheusAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	c := &Crow{
		cfg:        cfg,
		m:          newMetrics(),
		promClient: promapi.NewAPI(cli),

		quit: make(chan struct{}),

		sampleCh: make(chan []*sample),
	}
	return c, nil
}

func (c *Crow) runLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.quit:
			return
		case samples := <-c.sampleCh:
			c.m.totalScrapes.Inc()
			c.m.totalSamples.Add(float64(len(samples)))

			c.appendSamples(samples)
		case <-ticker.C:
			c.checkPending()
		}
	}
}

// appendSamples queues samples to be checked.
func (c *Crow) appendSamples(samples []*sample) {
	c.pendingMtx.Lock()
	defer c.pendingMtx.Unlock()
	c.pending = append(c.pending, samples...)
	c.m.pendingSets.Set(float64(len(c.pending)))
}

// checkPending iterates over all pending samples. Samples that are ready
// are immediately validated. Samples are requeued if they're not ready or
// not found during validation.
func (c *Crow) checkPending() {
	c.pendingMtx.Lock()
	defer c.pendingMtx.Unlock()

	now := time.Now().UTC()

	requeued := []*sample{}
	for _, s := range c.pending {
		if !s.Ready(now) {
			requeued = append(requeued, s)
			continue
		}

		err := c.validate(s)
		if err == nil {
			c.m.totalResults.WithLabelValues("success").Inc()
			continue
		}

		s.ValidationAttempt++
		if s.ValidationAttempt < c.cfg.MaxValidations {
			requeued = append(requeued, s)
			continue
		}

		var vf errValidationFailed
		if errors.As(err, &vf) {
			switch {
			case vf.mismatch:
				c.m.totalResults.WithLabelValues("mismatch").Inc()
			case vf.missing:
				c.m.totalResults.WithLabelValues("missing").Inc()
			default:
				c.m.totalResults.WithLabelValues("unknown").Inc()
			}
		}
	}
	c.pending = requeued
	c.m.pendingSets.Set(float64(len(c.pending)))
}

type errValidationFailed struct {
	missing  bool
	mismatch bool
}

func (e errValidationFailed) Error() string {
	switch {
	case e.missing:
		return "validation failed: sample missing"
	case e.mismatch:
		return "validation failed: sample does not match"
	default:
		return "validation failed"
	}
}

// validate validates a sample. If the sample should be requeued (i.e.,
// couldn't be found), returns true.
func (c *Crow) validate(b *sample) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.QueryTimeout)
	defer cancel()

	labels := make([]string, 0, len(b.Labels))
	for k, v := range b.Labels {
		labels = append(labels, fmt.Sprintf(`%s="%s"`, k, v))
	}
	if c.cfg.ExtraSelectors != "" {
		labels = append(labels, c.cfg.ExtraSelectors)
	}

	query := fmt.Sprintf("%s{%s}", validationSampleName, strings.Join(labels, ","))
	val, _, err := c.promClient.QueryRange(ctx, query, promapi.Range{
		Start: b.ScrapeTime.UTC().Add(-c.cfg.QueryDuration),
		End:   b.ScrapeTime.UTC().Add(+c.cfg.QueryDuration),
		Step:  c.cfg.QueryStep,
	})

	if err != nil {
		fmt.Println(err)
	} else if m, ok := val.(model.Matrix); ok {
		return c.validateInMatrix(b, m)
	}

	return errValidationFailed{missing: true}
}

func (c *Crow) validateInMatrix(b *sample, m model.Matrix) error {
	var found, matches bool

	for _, ss := range m {
		for _, sp := range ss.Values {
			ts := time.Unix(0, sp.Timestamp.UnixNano())
			dist := b.ScrapeTime.Sub(ts)
			if dist < 0 {
				dist = -dist
			}

			if dist <= c.cfg.MaxTimestampDelta {
				found = true
				matches = math.Abs(float64(sp.Value)-b.Value) <= c.cfg.ValueEpsilon
			}

			if found && matches {
				break
			}
		}
	}

	if !found || !matches {
		return errValidationFailed{
			missing:  !found,
			mismatch: found && !matches,
		}
	}
	return nil
}

// TestMetrics exposes a collector of test metrics. Each collection will
// schedule a validation job.
func (c *Crow) TestMetrics() prometheus.Collector {
	return &sampleGenerator{
		numSamples: c.cfg.NumSamples,
		sendCh:     c.sampleCh,

		r: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// StateMetrics exposes metrics of Crow itself. These metrics are not validated
// for presence in the remote system.
func (c *Crow) StateMetrics() prometheus.Collector { return c.m }

// Stop stops crow. Panics if Stop is called more than once.
func (c *Crow) Stop() {
	close(c.quit)
	c.wg.Wait()
}
