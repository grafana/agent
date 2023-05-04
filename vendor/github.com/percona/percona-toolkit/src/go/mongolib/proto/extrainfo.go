package proto

type ExtraInfo struct {
	PageFaults     float64 `bson:"page_faults"`
	HeapUsageBytes float64 `bson:"heap_usage_bytes"`
	Note           string  `bson:"note"`
}
