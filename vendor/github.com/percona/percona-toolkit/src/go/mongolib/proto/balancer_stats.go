package proto

type BalancerStats struct {
	Success  int64
	Failed   int64
	Splits   int64
	Drops    int64
	Settings map[string]interface{}
}
