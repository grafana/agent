package prometheus

// remoteWriteMapping maps a remote_write to a set of global ids
type remoteWriteMapping struct {
	RemoteWriteID string
	localToGlobal map[uint64]uint64
	globalToLocal map[uint64]uint64
}

func (rw *remoteWriteMapping) deleteStaleIDs(globalID uint64) {
	localID, found := rw.globalToLocal[globalID]
	if !found {
		return
	}
	delete(rw.globalToLocal, globalID)
	delete(rw.localToGlobal, localID)
}
