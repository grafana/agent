package proto

type ShardsMap struct {
	Map map[string]string `bson:"map"`
	OK  int               `bson:"ok"`
}
