package proto

type TimeMs struct {
	WriteToDataFiles   float64 `bson:"writeToDataFiles"`
	WriteToJournal     float64 `bson:"writeToJournal"`
	Commits            float64 `bson:"commits"`
	CommitsInWriteLock float64 `bson:"commitsInWriteLock"`
	Dt                 float64 `bson:"dt"`
	PrepLogBuffer      float64 `bson:"prepLogBuffer"`
	RemapPrivateView   float64 `bson:"remapPrivateView"`
}

type Dur struct {
	TimeMs             *TimeMs `bson:"timeMs"`
	WriteToDataFilesMB float64 `bson:"writeToDataFilesMB"`
	Commits            float64 `bson:"commits"`
	CommitsInWriteLock float64 `bson:"commitsInWriteLock"`
	Compression        float64 `bson:"compression"`
	EarlyCommits       float64 `bson:"earlyCommits"`
	JournaledMB        float64 `bson:"journaledMB"`
}
