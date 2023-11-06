package wal

import "github.com/prometheus/client_golang/prometheus"

type WatcherMetrics struct {
	recordsRead               *prometheus.CounterVec
	recordDecodeFails         *prometheus.CounterVec
	droppedWriteNotifications *prometheus.CounterVec
	segmentRead               *prometheus.CounterVec
	currentSegment            *prometheus.GaugeVec
	replaySegment             *prometheus.GaugeVec
	watchersRunning           *prometheus.GaugeVec
}

func NewWatcherMetrics(reg prometheus.Registerer) *WatcherMetrics {
	m := &WatcherMetrics{
		recordsRead: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "records_read_total",
				Help:      "Number of records read by the WAL watcher from the WAL.",
			},
			[]string{"id"},
		),
		recordDecodeFails: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "record_decode_failures_total",
				Help:      "Number of records read by the WAL watcher that resulted in an error when decoding.",
			},
			[]string{"id"},
		),
		droppedWriteNotifications: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "dropped_write_notifications_total",
				Help:      "Number of dropped write notifications due to having one already buffered.",
			},
			[]string{"id"},
		),
		segmentRead: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "segment_read_total",
				Help:      "Number of segment reads triggered by the backup timer firing.",
			},
			[]string{"id", "reason"},
		),
		currentSegment: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "current_segment",
				Help:      "Current segment the WAL watcher is reading records from.",
			},
			[]string{"id"},
		),
		replaySegment: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "replay_segment",
				Help:      "Segment the WAL watcher will start replaying the WAL from on startup.",
			},
			[]string{"id"},
		),
		watchersRunning: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "loki_write",
				Subsystem: "wal_watcher",
				Name:      "running",
				Help:      "Number of WAL watchers running.",
			},
			nil,
		),
	}

	if reg != nil {
		m.recordsRead = mustRegisterOrGet(reg, m.recordsRead).(*prometheus.CounterVec)
		m.recordDecodeFails = mustRegisterOrGet(reg, m.recordDecodeFails).(*prometheus.CounterVec)
		m.droppedWriteNotifications = mustRegisterOrGet(reg, m.droppedWriteNotifications).(*prometheus.CounterVec)
		m.segmentRead = mustRegisterOrGet(reg, m.segmentRead).(*prometheus.CounterVec)
		m.currentSegment = mustRegisterOrGet(reg, m.currentSegment).(*prometheus.GaugeVec)
		m.watchersRunning = mustRegisterOrGet(reg, m.watchersRunning).(*prometheus.GaugeVec)
	}

	return m
}

func mustRegisterOrGet(reg prometheus.Registerer, c prometheus.Collector) prometheus.Collector {
	if err := reg.Register(c); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return are.ExistingCollector
		}
		panic(err)
	}
	return c
}
