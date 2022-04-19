package exchange

type TargetSet struct {
	source  string
	targets []Target
}

func NewTargetSet(source string, targets []Target) TargetSet {
	return TargetSet{
		source:  source,
		targets: targets,
	}
}

func (t *TargetSet) Source() string {
	return t.source
}

func (t *TargetSet) Targets() []Target {
	newTargets := make([]Target, len(t.targets))
	copy(newTargets, t.targets)
	return newTargets
}
