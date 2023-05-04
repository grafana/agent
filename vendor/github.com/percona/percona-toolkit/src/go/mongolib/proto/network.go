package proto

type Network struct {
	BytesIn     float64 `bson:"bytesIn"`
	BytesOut    float64 `bson:"bytesOut"`
	NumRequests float64 `bson:"numRequests"`
}
