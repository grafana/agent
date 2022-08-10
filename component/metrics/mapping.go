package metrics

// remoteWriteMapping maps a remote_write to a set of global ids
type remoteWriteMapping struct {
	RemoteWriteID string
	localToGlobal map[uint64]RefID
	globalToLocal map[RefID]uint64
}

func (rw *remoteWriteMapping) deleteStaleIDs(globalID RefID) {
	localID, found := rw.globalToLocal[globalID]
	if !found {
		return
	}
	delete(rw.globalToLocal, globalID)
	delete(rw.localToGlobal, localID)
}
