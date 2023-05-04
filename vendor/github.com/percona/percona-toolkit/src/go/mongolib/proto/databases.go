package proto

// Database item plus struct to hold collections stats
type Database struct {
	Name       string `bson:"name"`
	SizeOnDisk int64  `bson:"sizeOnDisk"`
	Empty      bool   `bson:"empty"`
}

// Database struct for listDatabases command
type Databases struct {
	Databases []Database `bson:"databases"`
}
