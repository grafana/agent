package proto

type MasterDoc struct {
	SetName interface{} `bson:"setName"`
	Hosts   interface{} `bson:"hosts"`
	Msg     string      `bson:"msg"`
}
