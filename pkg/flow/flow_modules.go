package flow

// registerModule globally registeres a module with the root controller.
func (f *Flow) registerModule(moduleID string, m *module) {
	if f.rootController != nil {
		f.rootController.registerModule(moduleID, m)
		return
	}

	f.modulesMut.Lock()
	defer f.modulesMut.Unlock()

	f.modules[moduleID] = m
}

// unregisterModule globally unregisteres a module from the root controller.
func (f *Flow) unregisterModule(moduleID string) {
	if f.rootController != nil {
		f.rootController.unregisterModule(moduleID)
		return
	}

	f.modulesMut.Lock()
	defer f.modulesMut.Unlock()

	delete(f.modules, moduleID)
}

// getModule gets a module from the root controller.
func (f *Flow) getModule(moduleID string) (*module, bool) {
	if f.rootController != nil {
		return f.rootController.getModule(moduleID)
	}

	f.modulesMut.RLock()
	defer f.modulesMut.RUnlock()

	m, ok := f.modules[moduleID]
	return m, ok
}
