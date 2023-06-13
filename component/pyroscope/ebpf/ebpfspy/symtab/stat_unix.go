//go:build unix

package symtab

import (
	"os"
	"syscall"
)

type stat struct {
	dev uint64
	ino uint64
}

func statFromFileInfo(file os.FileInfo) stat {
	sys := file.Sys()
	sysStat, ok := sys.(*syscall.Stat_t)
	if !ok || sysStat == nil {
		return stat{}
	}
	return stat{
		dev: uint64(sysStat.Dev),
		ino: sysStat.Ino,
	}
}
