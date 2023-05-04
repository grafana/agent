package proto

type Asserts struct {
	User      float64 `bson:"user"`
	Warning   float64 `bson:"warning"`
	Msg       float64 `bson:"msg"`
	Regular   float64 `bson:"regular"`
	Rollovers float64 `bson:"rollovers"`
}
