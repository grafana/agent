package flow

// GetServiceConsumers implements [service.Host]. It returns a slice of
// [component.Component] and [service.Service]s which declared a dependency on
// the named service.
func (f *Flow) GetServiceConsumers(serviceName string) []any {
	// TODO(rfratto): return non-nil once it is possible for a service or
	// component to declare a dependency on a named service.
	return nil
}
