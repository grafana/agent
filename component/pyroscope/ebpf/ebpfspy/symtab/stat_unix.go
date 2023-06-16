//go:build unix

package symtab

import (
	"os"
	"syscall"
)

type Stat struct {
	dev uint64
	ino uint64
}

func statFromFileInfo(file os.FileInfo) Stat {
	sys := file.Sys()
	sysStat, ok := sys.(*syscall.Stat_t)
	if !ok || sysStat == nil {
		return Stat{}
	}
	return Stat{
		dev: uint64(sysStat.Dev),
		ino: sysStat.Ino,
	}
}
