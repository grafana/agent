//go:build windows

package positions

// This code is copied from Promtail. The positions package allows logging
// components to keep track of read file offsets on disk and continue from the
// same place in case of a restart.

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// writePositionFile is a fallback for Windows because renameio does not support Windows.
// See https://github.com/google/renameio#windows-support
func writePositionFile(filename string, positions map[Entry]string) error {
	buf, err := yaml.Marshal(File{
		Positions: positions,
	})
	if err != nil {
		return err
	}

	target := filepath.Clean(filename)
	temp := target + "-new"

	err = os.WriteFile(temp, buf, os.FileMode(positionFileMode))
	if err != nil {
		return err
	}

	return os.Rename(temp, target)
}
