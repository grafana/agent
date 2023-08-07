package component

type DataType string

var (
	// Targets represents things that need to be scraped. These are used by multiple telemetry signals
	// scraping components and often require special labels, e.g. __path__ label is required for scraping
	// logs from files using loki.source.file component.
	Targets = DataType("Targets")

	// LokiLogs represent logs in Loki format
	LokiLogs = DataType("Loki Logs")

	OTELTelemetry     = DataType("OTEL Telemetry")
	PromMetrics       = DataType("Prometheus Metrics")
	PyroscopeProfiles = DataType("Pyroscope Profiles")
)

type Metadata struct {
	Accepts []DataType
	Outputs []DataType
}

func (m Metadata) Empty() bool {
	return len(m.Accepts) == 0 && len(m.Outputs) == 0
}
