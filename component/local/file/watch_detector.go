package file

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type watchDetector struct {
	opts    watchOptions
	watcher *fsnotify.Watcher
	cancel  context.CancelFunc
}

type watchOptions struct {
	Logger   log.Logger
	Filename string
	UpdateCh chan<- struct{} // Where to send detected updates to
}

// newWatchDetector creates a new file update detector which uses filesystem
// events to detect that a file has changed.
func newWatchDetector(opts watchOptions) (*watchDetector, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.Add(opts.Filename); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	wd := &watchDetector{
		opts:    opts,
		watcher: w,
		cancel:  cancel,
	}

	go wd.wait(ctx)
	return wd, nil
}

func (wd *watchDetector) wait(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-wd.watcher.Errors:
			if err != nil {
				level.Warn(wd.opts.Logger).Log("msg", "got error from fsnotify watcher; treating as file updated event", "err", err)
				wd.forwardNotification()
			}
		case ev := <-wd.watcher.Events:
			// We only want events that actually change the file (e.g., ignore chmod)
			if ev.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			level.Debug(wd.opts.Logger).Log("msg", "got fsnotify event", "path", ev.Name, "op", ev.Op.String())
			wd.forwardNotification()
		}
	}
}

func (wd *watchDetector) forwardNotification() {
	select {
	case wd.opts.UpdateCh <- struct{}{}:
	default:
		// Already queued; no need to queue another event.
	}
}

func (wd *watchDetector) Close() error {
	wd.cancel()
	return wd.watcher.Close()
}
