package proto

type Connections struct {
	Available    float64 `bson:"available"`
	Current      float64 `bson:"current"`
	TotalCreated float64 `bson:"totalCreated"`
}
