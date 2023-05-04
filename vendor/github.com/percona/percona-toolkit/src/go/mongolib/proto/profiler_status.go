package proto

// ProfilerStatus is a struct to hold the results of db.getProfilingLevel()
//		var ps proto.ProfilerStatus
//		err := db.Run(bson.M{"profile": -1}, &ps)
type ProfilerStatus struct {
	Was      int64 `bson:"was"`
	SlowMs   int64 `bson:"slowms"`
	GleStats struct {
		ElectionID string `bson:"electionId"`
		LastOpTime int64  `bson:"lastOpTime"`
	} `bson:"$gleStats"`
}
