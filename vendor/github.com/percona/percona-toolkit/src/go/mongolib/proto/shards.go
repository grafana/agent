package proto

type Shard struct {
	ID   string `bson:"_id"`
	Host string `bson:"host"`
}

type ShardsInfo struct {
	Shards []Shard `bson:"shards"`
	OK     int     `bson:"ok"`
}
