package proto

// BuildInfo Struct to store results of calling session.BuildInfo()
type BuildInfo struct {
	Version        string
	VersionArray   []int32
	GitVersion     string
	OpenSSLVersion string
	SysInfo        string
	Bits           int32
	Debug          bool
	MaxObjectSize  int64
}
