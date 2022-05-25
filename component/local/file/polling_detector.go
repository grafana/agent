package file

import (
	"context"
	"os"
	"time"
)

type pollingDetector struct {
	opts        pollingOptions
	lastModTime time.Time
	cancel      context.CancelFunc
}

type pollingOptions struct {
	Filename      string
	UpdateCh      chan<- struct{} // Where to send detected updates to
	PollFrequency time.Duration
}

// newPollingDetector creates a new poll-based file update detactor.
func newPollingDetector(opts pollingOptions) (*pollingDetector, error) {
	fi, err := os.Stat(opts.Filename)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	pw := &pollingDetector{
		opts:        opts,
		lastModTime: fi.ModTime(),
		cancel:      cancel,
	}

	go pw.run(ctx)
	return pw, nil
}

func (pw *pollingDetector) run(ctx context.Context) {
	t := time.NewTicker(pw.opts.PollFrequency)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			pw.checkFileUpdated()
		}
	}
}

func (pw *pollingDetector) checkFileUpdated() {
	fi, err := os.Stat(pw.opts.Filename)
	if err != nil {
		// We failed to stat the file. We send an event over the channel so that
		// the component can process the error when it tries to re-read the file.

		select {
		case pw.opts.UpdateCh <- struct{}{}:
		default:
			// Event already queued; no need to process more than one.
		}
		return
	}

	if modTime := fi.ModTime(); modTime.After(pw.lastModTime) {
		pw.lastModTime = modTime

		select {
		case pw.opts.UpdateCh <- struct{}{}:
		default:
			// Event already queued; no need to process more than one.
		}
	}
}

// Close terminates the poll-based file watcher.
func (pw *pollingDetector) Close() error {
	pw.cancel()
	return nil
}
