package proto

type Cursors struct {
	ClientCursorsSize float64 `bson:"clientCursors_size"`
	Note              string  `bson:"note"`
	Pinned            float64 `bson:"pinned"`
	TimedOut          float64 `bson:"timedOut"`
	TotalNoTimeout    float64 `bson:"totalNoTimeout"`
	TotalOpen         float64 `bson:"totalOpen"`
}
