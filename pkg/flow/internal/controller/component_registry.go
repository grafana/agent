package controller

import "github.com/grafana/agent/component"

// ComponentRegistry is a collection of registered components.
type ComponentRegistry interface {
	// Get looks up a component by name.
	Get(name string) (component.Registration, bool)
}

// DefaultComponentRegistry is the default [ComponentRegistry] which gets
// components registered to github.com/grafana/agent/component.
type DefaultComponentRegistry struct{}

// Get retrieves a component using [component.Get].
func (reg DefaultComponentRegistry) Get(name string) (component.Registration, bool) {
	return component.Get(name)
}

// RegistryMap is a map which implements [ComponentRegistry].
type RegistryMap map[string]component.Registration

// Get retrieves a component using [component.Get].
func (m RegistryMap) Get(name string) (component.Registration, bool) {
	reg, ok := m[name]
	return reg, ok
}
