package proto

type Mem struct {
	Bits              float64 `bson:"bits"`
	Mapped            float64 `bson:"mapped"`
	MappedWithJournal float64 `bson:"mappedWithJournal"`
	Resident          float64 `bson:"resident"`
	Supported         bool    `bson:"supported"`
	Virtual           float64 `bson:"virtual"`
}
