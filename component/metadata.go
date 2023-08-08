package component

type DataType string

var (
	// DataTypeTargets represents things that need to be scraped. These are used by multiple telemetry signals
	// scraping components and often require special labels, e.g. __path__ label is required for scraping
	// logs from files using loki.source.file component.
	DataTypeTargets = DataType("Targets")

	// DataTypeLokiLogs represent logs in Loki format
	DataTypeLokiLogs = DataType("Loki Logs")

	DataTypeOTELTelemetry     = DataType("OTEL Telemetry")
	DataTypePromMetrics       = DataType("Prometheus Metrics")
	DataTypePyroscopeProfiles = DataType("Pyroscope Profiles")
)

func TargetDiscoveryMetadata() Metadata {
	return Metadata{
		Outputs: []DataType{DataTypeTargets},
	}
}

func TargetsProcessingMetadata() Metadata {
	return Metadata{
		Accepts: []DataType{DataTypeTargets},
		Outputs: []DataType{DataTypeTargets},
	}
}

func LokiLogsProcessingMetadata() Metadata {
	return Metadata{
		Accepts: []DataType{DataTypeLokiLogs},
		Outputs: []DataType{DataTypeLokiLogs},
	}
}

func LokiLogsScraperMetadata() Metadata {
	return Metadata{
		Accepts: []DataType{DataTypeTargets},
		Outputs: []DataType{DataTypeLokiLogs},
	}
}

func LokiLogsSourceMetadata() Metadata {
	return Metadata{
		Outputs: []DataType{DataTypeLokiLogs},
	}
}

func LokiLogsSinkMetadata() Metadata {
	return Metadata{
		Accepts: []DataType{DataTypeLokiLogs},
	}
}

type Metadata struct {
	Accepts []DataType
	Outputs []DataType
}

func (m Metadata) Empty() bool {
	return len(m.Accepts) == 0 && len(m.Outputs) == 0
}
