package proto

type ChunksByCollection struct {
	ID    string `bson:"_id"` // Namespace
	Count int    `bson:"count"`
}
