package proto

type BackgroundFlushing struct {
	AverageMs    float64 `bson:"average_ms"`
	Flushes      float64 `bson:"flushes"`
	LastFinished string  `bson:"last_finished"`
	LastMs       float64 `bson:"last_ms"`
	TotalMs      float64 `bson:"total_ms"`
}
