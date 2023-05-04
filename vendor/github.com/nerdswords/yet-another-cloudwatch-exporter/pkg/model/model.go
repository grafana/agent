package model

const (
	DefaultPeriodSeconds = int64(300)
	DefaultLengthSeconds = int64(300)
	DefaultDelaySeconds  = int64(300)
)

type LabelSet map[string]struct{}

type Tag struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}
