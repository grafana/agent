//go:build linux

package cache

import (
	"os"
	"syscall"
)

// copypaste from https://github.com/grafana/pyroscope/blob/8a7fe2b80c219bfda9be685ff27ca1dee4218a42/ebpf/symtab/stat_linux.go#L14-L13

type Stat struct {
	Dev   uint64
	Inode uint64
}

func StatFromFileInfo(file os.FileInfo) Stat {
	sys := file.Sys()
	sysStat, ok := sys.(*syscall.Stat_t)
	if !ok || sysStat == nil {
		return Stat{}
	}
	return Stat{
		Dev:   sysStat.Dev,
		Inode: sysStat.Ino,
	}
}
