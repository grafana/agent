package proto

type Opcounters struct {
	Command float64 `bson:"command"`
	Delete  float64 `bson:"delete"`
	Getmore float64 `bson:"getmore"`
	Insert  float64 `bson:"insert"`
	Query   float64 `bson:"query"`
	Update  float64 `bson:"update"`
}
