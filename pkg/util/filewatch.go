package util

import (
	"context"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
)

type ReloadFunc func() error

type FileWatcher struct {
	filePath     string
	reloadFunc   ReloadFunc
	pollInterval time.Duration
}

// NewFileWatcher creates a new file watcher that will call reloadFunc when the file at filePath is modified.
func NewFileWatcher(filePath string, reloadFunc ReloadFunc, pollInterval time.Duration) *FileWatcher {
	return &FileWatcher{
		filePath:     filePath,
		reloadFunc:   reloadFunc,
		pollInterval: pollInterval,
	}
}

// Watch starts watching the file for changes and calls reloadFunc when the file is modified.
func (fw *FileWatcher) Watch(l *logging.Logger, ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		level.Error(l).Log("msg", "failed to create watcher", "err", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(fw.filePath)
	if err != nil {
		level.Error(l).Log("msg", "failed to add file to watcher", "err", err)
		return
	}

	ticker := time.NewTicker(fw.pollInterval)
	defer ticker.Stop()

	var lastModTime time.Time
	if fileInfo, err := os.Stat(fw.filePath); err == nil {
		lastModTime = fileInfo.ModTime()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := fw.reloadFunc(); err != nil {
					level.Error(l).Log("msg", "failed to reload", "err", err)
				}
			}
		case err := <-watcher.Errors:
			level.Error(l).Log("msg", "watcher error", "err", err)
		case <-ticker.C:
			if fileInfo, err := os.Stat(fw.filePath); err == nil {
				modTime := fileInfo.ModTime()
				if modTime.After(lastModTime) {
					lastModTime = modTime
					if err := fw.reloadFunc(); err != nil {
						level.Error(l).Log("msg", "failed to reload", "err", err)
					}
				}
			}
		}
	}
}
