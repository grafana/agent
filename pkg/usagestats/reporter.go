package usagestats

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/multierror"
	"github.com/prometheus/common/version"
)

var (
	reportCheckInterval = time.Minute
	reportInterval      = 4 * time.Hour
)

// Reporter holds the agent seed information and sends report of usage
type Reporter struct {
	logger log.Logger
	cfg    *config.Config

	agentSeed  *AgentSeed
	lastReport time.Time
}

// AgentSeed identifies a unique agent
type AgentSeed struct {
	UID       string    `json:"UID"`
	CreatedAt time.Time `json:"created_at"`
	Version   string    `json:"version"`
}

// NewReporter creates a Reporter that will send periodically reports to grafana.com
func NewReporter(logger log.Logger, cfg *config.Config) (*Reporter, error) {
	r := &Reporter{
		logger: logger,
		cfg:    cfg,
	}
	return r, nil
}

func (rep *Reporter) init(ctx context.Context) error {
	path := agentSeedFileName()

	if fileExists(path) {
		seed, err := rep.readSeedFile(path)
		rep.agentSeed = seed
		return err
	}
	rep.agentSeed = &AgentSeed{
		UID:       uuid.NewString(),
		Version:   version.Version,
		CreatedAt: time.Now(),
	}
	return rep.writeSeedFile(*rep.agentSeed, path)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// readSeedFile reads the agent seed file
func (rep *Reporter) readSeedFile(path string) (*AgentSeed, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seed := &AgentSeed{}
	err = json.Unmarshal(data, seed)
	if err != nil {
		return nil, err
	}
	return seed, nil
}

// writeSeedFile writes the agent seed file
func (rep *Reporter) writeSeedFile(seed AgentSeed, path string) error {
	data, err := json.Marshal(seed)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func agentSeedFileName() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "agent_seed.json")
	}
	// linux/mac
	return "/tmp/agent_seed.json"
}

// Start inits the reporter seed and start sending report for every interval
func (rep *Reporter) Start(ctx context.Context) error {
	level.Info(rep.logger).Log("msg", "running usage stats reporter")
	err := rep.init(ctx)
	if err != nil {
		level.Info(rep.logger).Log("msg", "failed to init seed", "err", err)
		return err
	}

	// check every minute if we should report.
	ticker := time.NewTicker(reportCheckInterval)
	defer ticker.Stop()

	// find  when to send the next report.
	next := nextReport(reportInterval, rep.agentSeed.CreatedAt, time.Now())
	if rep.lastReport.IsZero() {
		// if we never reported assumed it was the last interval.
		rep.lastReport = next.Add(-reportInterval)
	}
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if !next.Equal(now) && now.Sub(rep.lastReport) < reportInterval {
				continue
			}
			level.Info(rep.logger).Log("msg", "reporting cluster stats", "date", time.Now())
			if err := rep.reportUsage(ctx, next); err != nil {
				level.Info(rep.logger).Log("msg", "failed to report usage", "err", err)
				continue
			}
			rep.lastReport = next
			next = next.Add(reportInterval)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reportUsage reports the usage to grafana.com.
func (rep *Reporter) reportUsage(ctx context.Context, interval time.Time) error {
	backoff := backoff.New(ctx, backoff.Config{
		MinBackoff: time.Second,
		MaxBackoff: 30 * time.Second,
		MaxRetries: 5,
	})
	var errs multierror.MultiError
	for backoff.Ongoing() {
		if err := sendReport(ctx, rep.agentSeed, interval, rep.getMetrics()); err != nil {
			level.Info(rep.logger).Log("msg", "failed to send usage report", "retries", backoff.NumRetries(), "err", err)
			errs.Add(err)
			backoff.Wait()
			continue
		}
		level.Info(rep.logger).Log("msg", "usage report sent with success")
		return nil
	}
	return errs.Err()
}

func (rep *Reporter) getMetrics() map[string]interface{} {
	return map[string]interface{}{
		"enabled-features": rep.cfg.EnabledFeatures,
	}
}

// nextReport compute the next report time based on the interval.
// The interval is based off the creation of the agent seed to avoid all agents reporting at the same time.
func nextReport(interval time.Duration, createdAt, now time.Time) time.Time {
	duration := math.Ceil(float64(now.Sub(createdAt)) / float64(interval))
	return createdAt.Add(time.Duration(duration) * interval)
}
