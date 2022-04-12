package pogo

type Target struct {
	address  string
	labels   map[string]string
	metadata map[string]string
}

func NewTarget(address string, labels map[string]string, metadata map[string]string) Target {
	return Target{
		address:  address,
		labels:   labels,
		metadata: metadata,
	}
}

func CopyTarget(in Target) Target {
	return Target{
		address:  in.Address(),
		labels:   in.Labels(),
		metadata: in.Metadata(),
	}
}

func (t *Target) Address() string {
	return t.address
}

func (t *Target) Labels() map[string]string {
	return copyMap(t.labels)
}

func (t *Target) Metadata() map[string]string {
	return copyMap(t.metadata)
}
