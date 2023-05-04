package proto

type AcquiredLocks struct {
	AcquireCount        *AcquireCount `bson:"acquireCount"`
	AcquireWaitCount    float64       `bson:"acquireWaitCount.W"`
	TimeAcquiringMicros float64       `bson:"timeAcquiringMicros.W"`
}
