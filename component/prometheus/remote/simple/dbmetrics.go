package simple

import "github.com/prometheus/client_golang/prometheus"

type dbmetrics struct {
	r prometheus.Registerer
	d *dbstore

	writeTime    prometheus.Histogram
	readTime     prometheus.Histogram
	currentKey   prometheus.Gauge
	totalKeys    *prometheus.Desc
	diskSize     *prometheus.Desc
	serriesCount *prometheus.Desc

	evictionTime prometheus.Histogram
	lastEviction prometheus.Gauge

	totalRecords prometheus.Gauge
}

func newDbMetrics(r prometheus.Registerer, d *dbstore) *dbmetrics {
	dbm := &dbmetrics{
		r: r,
		d: d,
	}

	dbm.writeTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "agent_simple_wal_write_time",
		Help: "The write time for writing to WAL in seconds",
	})

	dbm.readTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "agent_simple_wal_read_time",
		Help: "The read time for reading from the WAL in seconds",
	})

	dbm.currentKey = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_simple_current_key",
		Help: "The current key in the WAL",
	})

	dbm.totalKeys = prometheus.NewDesc("agent_simple_total_keys",
		"Total number of active keys in the WAL", nil, nil)

	dbm.evictionTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "agent_simple_eviction_time",
		Help: "Eviction times for cleaning the WAL in seconds",
	})

	dbm.lastEviction = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_simple_last_eviction",
		Help: "Last eviction in unix timestamp seconds",
	})

	dbm.totalRecords = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_simple_total_signals",
		Help: "Total number of signals in the WAL",
	})

	dbm.diskSize = prometheus.NewDesc("agent_simple_wal_disk_size",
		"Size of WAL in kilobytes", nil, nil)

	dbm.serriesCount = prometheus.NewDesc("agent_simple_wal_sample_count",
		"Total number of samples in the WAL", nil, nil)

	dbm.r.MustRegister(
		dbm.lastEviction,
		dbm.evictionTime,
		dbm.currentKey,
		dbm.readTime,
		dbm.writeTime,
		dbm.totalRecords,
		dbm,
	)

	return dbm
}

func (dbm *dbmetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- dbm.totalKeys
	ch <- dbm.diskSize
	ch <- dbm.serriesCount
}

func (dbm *dbmetrics) Collect(ch chan<- prometheus.Metric) {
	keyCount := dbm.d.getKeyCount()
	ch <- prometheus.MustNewConstMetric(dbm.totalKeys, prometheus.GaugeValue, float64(keyCount))
	fileSize := dbm.d.getFileSize()
	ch <- prometheus.MustNewConstMetric(dbm.diskSize, prometheus.GaugeValue, fileSize)
	sampleCount := dbm.d.sampleCount()
	ch <- prometheus.MustNewConstMetric(dbm.serriesCount, prometheus.GaugeValue, sampleCount)
}
