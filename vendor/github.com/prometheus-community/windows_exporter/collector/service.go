// +build windows

package collector

import (
	"strconv"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollectorWithConfig("service", func() Config { return &ServiceConfig{} })
}

type ServiceConfig struct {
	ServiceWhereClause string
}

func (s *ServiceConfig) RegisterKingpin(ka *kingpin.Application) {
	ka.Flag(
		"collector.service.services-where",
		"WQL 'where' clause to use in WMI metrics query. Limits the response to the services you specify and reduces the size of the response.",
	).Default("").StringVar(&s.ServiceWhereClause)
}

func (s *ServiceConfig) Build() (Collector, error) {
	return NewserviceCollector(s)
}

// A serviceCollector is a Prometheus collector for WMI Win32_Service metrics
type serviceCollector struct {
	Information *prometheus.Desc
	State       *prometheus.Desc
	StartMode   *prometheus.Desc
	Status      *prometheus.Desc

	queryWhereClause string
}

// NewserviceCollector ...
func NewserviceCollector(s *ServiceConfig) (Collector, error) {
	const subsystem = "service"

	if s.ServiceWhereClause == "" {
		log.Warn("No where-clause specified for service collector. This will generate a very large number of metrics!")
	}

	return &serviceCollector{
		Information: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"A metric with a constant '1' value labeled with service information",
			[]string{"name", "display_name", "process_id", "run_as"},
			nil,
		),
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The state of the service (State)",
			[]string{"name", "state"},
			nil,
		),
		StartMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_mode"),
			"The start mode of the service (StartMode)",
			[]string{"name", "start_mode"},
			nil,
		),
		Status: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "status"),
			"The status of the service (Status)",
			[]string{"name", "status"},
			nil,
		),
		queryWhereClause: s.ServiceWhereClause,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *serviceCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting service metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_Service docs:
// - https://msdn.microsoft.com/en-us/library/aa394418(v=vs.85).aspx
type Win32_Service struct {
	DisplayName string
	Name        string
	ProcessId   uint32
	State       string
	Status      string
	StartMode   string
	StartName   *string
}

var (
	allStates = []string{
		"stopped",
		"start pending",
		"stop pending",
		"running",
		"continue pending",
		"pause pending",
		"paused",
		"unknown",
	}
	allStartModes = []string{
		"boot",
		"system",
		"auto",
		"manual",
		"disabled",
	}
	allStatuses = []string{
		"ok",
		"error",
		"degraded",
		"unknown",
		"pred fail",
		"starting",
		"stopping",
		"service",
		"stressed",
		"nonrecover",
		"no contact",
		"lost comm",
	}
)

func (c *serviceCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_Service
	q := queryAllWhere(&dst, c.queryWhereClause)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	for _, service := range dst {
		pid := strconv.FormatUint(uint64(service.ProcessId), 10)

		runAs := ""
		if service.StartName != nil {
			runAs = *service.StartName
		}
		ch <- prometheus.MustNewConstMetric(
			c.Information,
			prometheus.GaugeValue,
			1.0,
			strings.ToLower(service.Name),
			service.DisplayName,
			pid,
			runAs,
		)

		for _, state := range allStates {
			isCurrentState := 0.0
			if state == strings.ToLower(service.State) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				strings.ToLower(service.Name),
				state,
			)
		}

		for _, startMode := range allStartModes {
			isCurrentStartMode := 0.0
			if startMode == strings.ToLower(service.StartMode) {
				isCurrentStartMode = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				strings.ToLower(service.Name),
				startMode,
			)
		}

		for _, status := range allStatuses {
			isCurrentStatus := 0.0
			if status == strings.ToLower(service.Status) {
				isCurrentStatus = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.Status,
				prometheus.GaugeValue,
				isCurrentStatus,
				strings.ToLower(service.Name),
				status,
			)
		}
	}
	return nil, nil
}
