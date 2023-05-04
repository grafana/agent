package proto

import "time"

type ProcInfo struct {
	CreateTime time.Time
	Path       string
	UserName   string
	Error      error
}

type GetHostInfo struct {
	Hostname          string
	HostOsType        string
	HostSystemCPUArch string
	HostDatabases     int
	HostCollections   int
	DBPath            string

	ProcPath         string
	ProcUserName     string
	ProcCreateTime   time.Time
	ProcProcessCount int

	// Server Status
	ProcessName    string
	ReplicasetName string
	Version        string
	NodeType       string
}
