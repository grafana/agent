package sql_exporter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/burningalchemist/sql_exporter/config"
	"github.com/burningalchemist/sql_exporter/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

var enablePing = flag.Bool("config.enable-ping", true, "Enable ping for targets")

const (
	// Capacity for the channel to collect metrics.
	capMetricChan = 1000

	upMetricName       = "up"
	upMetricHelp       = "1 if the target is reachable, or 0 if the scrape failed"
	scrapeDurationName = "scrape_duration_seconds"
	scrapeDurationHelp = "How long it took to scrape the target in seconds"
)

// Target collects SQL metrics from a single sql.DB instance. It aggregates one or more Collectors and it looks much
// like a prometheus.Collector, except its Collect() method takes a Context to run in.
type Target interface {
	// Collect is the equivalent of prometheus.Collector.Collect(), but takes a context to run in.
	Collect(ctx context.Context, ch chan<- Metric)
}

// target implements Target. It wraps a sql.DB, which is initially nil but never changes once instantianted.
type target struct {
	name               string
	dsn                string
	collectors         []Collector
	constLabels        prometheus.Labels
	globalConfig       *config.GlobalConfig
	upDesc             MetricDesc
	scrapeDurationDesc MetricDesc
	logContext         string

	conn *sql.DB
}

// NewTarget returns a new Target with the given instance name, data source name, collectors and constant labels.
// An empty target name means the exporter is running in single target mode: no synthetic metrics will be exported.
func NewTarget(
	logContext, name, dsn string, ccs []*config.CollectorConfig, constLabels prometheus.Labels, gc *config.GlobalConfig) (
	Target, errors.WithContext,
) {
	if name != "" {
		logContext = fmt.Sprintf("%s, target=%q", logContext, name)
	}

	constLabelPairs := make([]*dto.LabelPair, 0, len(constLabels))
	for n, v := range constLabels {
		constLabelPairs = append(constLabelPairs, &dto.LabelPair{
			Name:  proto.String(n),
			Value: proto.String(v),
		})
	}
	sort.Sort(labelPairSorter(constLabelPairs))

	collectors := make([]Collector, 0, len(ccs))
	for _, cc := range ccs {
		c, err := NewCollector(logContext, cc, constLabelPairs)
		if err != nil {
			return nil, err
		}
		collectors = append(collectors, c)
	}

	upDesc := NewAutomaticMetricDesc(logContext, upMetricName, upMetricHelp, prometheus.GaugeValue, constLabelPairs)
	scrapeDurationDesc := NewAutomaticMetricDesc(logContext, scrapeDurationName, scrapeDurationHelp, prometheus.GaugeValue, constLabelPairs)
	t := target{
		name:               name,
		dsn:                dsn,
		collectors:         collectors,
		constLabels:        constLabels,
		globalConfig:       gc,
		upDesc:             upDesc,
		scrapeDurationDesc: scrapeDurationDesc,
		logContext:         logContext,
	}
	return &t, nil
}

// Collect implements Target.
func (t *target) Collect(ctx context.Context, ch chan<- Metric) {
	var (
		scrapeStart = time.Now()
		targetUp    = true
	)

	err := t.ping(ctx)
	if err != nil {
		ch <- NewInvalidMetric(errors.Wrap(t.logContext, err))
		targetUp = false
	}
	if t.name != "" {
		// Export the target's `up` metric as early as we know what it should be.
		ch <- NewMetric(t.upDesc, boolToFloat64(targetUp))
	}

	var wg sync.WaitGroup
	// Don't bother with the collectors if target is down.
	if targetUp {
		wg.Add(len(t.collectors))
		for _, c := range t.collectors {
			// If using a single DB connection, collectors will likely run sequentially anyway. But we might have more.
			go func(collector Collector) {
				defer wg.Done()
				collector.Collect(ctx, t.conn, ch)
			}(c)
		}
	}
	// Wait for all collectors (if any) to complete.
	wg.Wait()

	if t.name != "" {
		// And export a `scrape duration` metric once we're done scraping.
		ch <- NewMetric(t.scrapeDurationDesc, float64(time.Since(scrapeStart))*1e-9)
	}
}

func (t *target) ping(ctx context.Context) errors.WithContext {
	// Create the DB handle, if necessary. It won't usually open an actual connection, so we'll need to ping afterwards.
	// We cannot do this only once at creation time because the sql.Open() documentation says it "may" open an actual
	// connection, so it "may" actually fail to open a handle to a DB that's initially down.
	if t.conn == nil {
		conn, err := OpenConnection(ctx, t.logContext, t.dsn, t.globalConfig.MaxConns, t.globalConfig.MaxIdleConns, t.globalConfig.MaxConnLifetime)
		if err != nil {
			if err != ctx.Err() {
				return errors.Wrap(t.logContext, err)
			}
			// if err == ctx.Err() fall through
		} else {
			t.conn = conn
		}
	}

	// If we have a handle and the context is not closed, test whether the database is up.
	// FIXME: we ping the database during each request even with cacheCollector. It leads
	// to additional charges for paid database services.
	if t.conn != nil && ctx.Err() == nil && *enablePing {
		var err error
		// Ping up to max_connections + 1 times as long as the returned error is driver.ErrBadConn, to purge the connection
		// pool of bad connections. This might happen if the previous scrape timed out and in-flight queries got canceled.
		for i := 0; i <= t.globalConfig.MaxConns; i++ {
			if err = PingDB(ctx, t.conn); err != driver.ErrBadConn {
				break
			}
		}
		if err != nil {
			return errors.Wrap(t.logContext, err)
		}
	}

	if ctx.Err() != nil {
		return errors.Wrap(t.logContext, ctx.Err())
	}
	return nil
}

// boolToFloat64 converts a boolean flag to a float64 value (0.0 or 1.0).
func boolToFloat64(value bool) float64 {
	if value {
		return 1.0
	}
	return 0.0
}
