package usagestats

import (
	"context"
	"errors"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/dskit/backoff"
	"github.com/grafana/dskit/multierror"
	"github.com/grafana/dskit/services"
	"github.com/grafana/loki/pkg/util/build"
)

const (
	// File name for the cluster seed file.
	clusterSeedFileName = "cluster_seed.json"
)

var (
	reportCheckInterval = time.Minute
	reportInterval      = 1 * time.Hour
)

// Reporter holds the cluster information and sends report of usage
type Reporter struct {
	logger log.Logger
	cfg    *config.Config
	services.Service

	cluster    *ClusterSeed
	lastReport time.Time
}

// NewReporter creates a Reporter that will send periodically reports to grafana.com
func NewReporter(logger log.Logger, cfg *config.Config) (*Reporter, error) {
	r := &Reporter{
		logger: logger,
		cfg:    cfg,
	}

	if cfg.EnableUsageReport {
		r.Service = services.NewBasicService(nil, r.start, nil)
		r.Service.StartAsync(context.Background())
	} else {
		// builds an empty service
		r.Service = services.NewBasicService(nil, nil, nil)
	}
	return r, nil
}

func (rep *Reporter) init(ctx context.Context) {
	if fileExists(clusterSeedFileName) {
		rep.cluster, _ = rep.readSeedFile()
	} else {
		rep.cluster = &ClusterSeed{
			UID:               uuid.NewString(),
			PrometheusVersion: build.GetVersion(),
			CreatedAt:         time.Now(),
		}
		rep.writeSeedFile(*rep.cluster)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// readSeedFile reads the cluster seed file
func (rep *Reporter) readSeedFile() (*ClusterSeed, error) {
	data, err := ioutil.ReadFile(clusterSeedFileName)
	if err != nil {
		return nil, err
	}
	seed, err := JSONCodec.Decode(data)
	if err != nil {
		return nil, err
	}
	return seed.(*ClusterSeed), nil
}

// writeSeedFile writes the cluster seed file
func (rep *Reporter) writeSeedFile(seed ClusterSeed) error {
	data, err := JSONCodec.Encode(seed)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(clusterSeedFileName, data, 0644)
}

// start inits the reporter seed and start sending report for every interval
func (rep *Reporter) start(ctx context.Context) error {
	level.Info(rep.logger).Log("msg", "running usage stats reporter")
	rep.init(ctx)

	// check every minute if we should report.
	ticker := time.NewTicker(reportCheckInterval)
	defer ticker.Stop()

	// find  when to send the next report.
	next := nextReport(reportInterval, rep.cluster.CreatedAt, time.Now())
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
		if err := sendReport(ctx, rep.cluster, interval, rep.getMetrics()); err != nil {
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
// The interval is based off the creation of the cluster seed to avoid all cluster reporting at the same time.
func nextReport(interval time.Duration, createdAt, now time.Time) time.Time {
	// createdAt * (x * interval ) >= now
	return createdAt.Add(time.Duration(math.Ceil(float64(now.Sub(createdAt))/float64(interval))) * interval)
}
