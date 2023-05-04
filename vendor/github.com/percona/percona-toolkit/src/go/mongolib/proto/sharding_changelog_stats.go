package proto

type ShardingChangelogSummaryId struct {
	Event string `bson:"event"`
	Note  string `bson:"note"`
}

type ShardingChangelogSummary struct {
	Id    *ShardingChangelogSummaryId `bson:"_id"`
	Count float64                     `bson:"count"`
}

type ShardingChangelogStats struct {
	Items *[]ShardingChangelogSummary
}
