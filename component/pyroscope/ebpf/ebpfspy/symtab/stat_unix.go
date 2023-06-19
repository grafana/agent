//go:build unix

package symtab

import (
	"os"
	"syscall"
)

type Stat struct {
	dev uint64 `river:"dev,block,optional"`
	ino uint64 `river:"ino,block,optional"`
}

func statFromFileInfo(file os.FileInfo) Stat {
	sys := file.Sys()
	sysStat, ok := sys.(*syscall.Stat_t)
	if !ok || sysStat == nil {
		return Stat{}
	}
	return Stat{
		dev: sysStat.Dev,
		ino: sysStat.Ino,
	}
}
