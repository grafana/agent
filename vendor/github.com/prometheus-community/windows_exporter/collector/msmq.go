// +build windows

package collector

import (
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollectorWithConfig("msmq", func() Config { return &MSMQConfig{} })
}

type MSMQConfig struct {
	MSMQWhereClause string
}

func (m *MSMQConfig) RegisterKingpin(ka *kingpin.Application) {
	ka.Flag(
		"collector.msmq.msmq-where",
		"WQL 'where' clause to use in WMI metrics query. Limits the response to the msmqs you specify and reduces the size of the response.",
	).StringVar(&m.MSMQWhereClause)
}

func (m *MSMQConfig) Build() (Collector, error) {
	return NewMSMQCollector(m)
}

// A Win32_PerfRawData_MSMQ_MSMQQueueCollector is a Prometheus collector for WMI Win32_PerfRawData_MSMQ_MSMQQueue metrics
type Win32_PerfRawData_MSMQ_MSMQQueueCollector struct {
	BytesinJournalQueue    *prometheus.Desc
	BytesinQueue           *prometheus.Desc
	MessagesinJournalQueue *prometheus.Desc
	MessagesinQueue        *prometheus.Desc

	queryWhereClause string
}

// NewWin32_PerfRawData_MSMQ_MSMQQueueCollector ...
func NewMSMQCollector(m *MSMQConfig) (Collector, error) {
	const subsystem = "msmq"

	if m.MSMQWhereClause == "" {
		log.Warn("No where-clause specified for msmq collector. This will generate a very large number of metrics!")
	}

	return &Win32_PerfRawData_MSMQ_MSMQQueueCollector{
		BytesinJournalQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_in_journal_queue"),
			"Size of queue journal in bytes",
			[]string{"name"},
			nil,
		),
		BytesinQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "bytes_in_queue"),
			"Size of queue in bytes",
			[]string{"name"},
			nil,
		),
		MessagesinJournalQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "messages_in_journal_queue"),
			"Count messages in queue journal",
			[]string{"name"},
			nil,
		),
		MessagesinQueue: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "messages_in_queue"),
			"Count messages in queue",
			[]string{"name"},
			nil,
		),
		queryWhereClause: m.MSMQWhereClause,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Win32_PerfRawData_MSMQ_MSMQQueueCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting msmq metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_MSMQ_MSMQQueue struct {
	Name string

	BytesinJournalQueue    uint64
	BytesinQueue           uint64
	MessagesinJournalQueue uint64
	MessagesinQueue        uint64
}

func (c *Win32_PerfRawData_MSMQ_MSMQQueueCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_MSMQ_MSMQQueue
	q := queryAllWhere(&dst, c.queryWhereClause)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, msmq := range dst {

		if msmq.Name == "Computer Queues" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.BytesinJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesinJournalQueue),
			strings.ToLower(msmq.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.BytesinQueue,
			prometheus.GaugeValue,
			float64(msmq.BytesinQueue),
			strings.ToLower(msmq.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.MessagesinJournalQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesinJournalQueue),
			strings.ToLower(msmq.Name),
		)
		ch <- prometheus.MustNewConstMetric(
			c.MessagesinQueue,
			prometheus.GaugeValue,
			float64(msmq.MessagesinQueue),
			strings.ToLower(msmq.Name),
		)
	}
	return nil, nil
}
