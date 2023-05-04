package proto

type GlobalLock struct {
	ActiveClients *ActiveClients `bson:"activeClients"`
	CurrentQueue  *CurrentQueue  `bson:"currentQueue"`
	TotalTime     int64          `bson:"totalTime"`
}

type ActiveClients struct {
	Readers int64 `bson:"readers"`
	Total   int64 `bson:"total"`
	Writers int64 `bson:"writers"`
}

type CurrentQueue struct {
	Writers int64 `bson:"writers"`
	Readers int64 `bson:"readers"`
	Total   int64 `bson:"total"`
}
