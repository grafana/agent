package file

import (
	"context"
	"encoding"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Detector is used to specify how changes to the file should be detected.
type Detector int

const (
	// DetectorInvalid indicates an invalid UpdateType.
	DetectorInvalid Detector = iota
	// DetectorFSNotify uses filesystem events to wait for changes to the file.
	DetectorFSNotify
	// DetectorPoll will re-read the file on an interval to detect changes.
	DetectorPoll

	// DetectorDefault holds the default UpdateType.
	DetectorDefault = DetectorFSNotify
)

var (
	_ encoding.TextMarshaler   = Detector(0)
	_ encoding.TextUnmarshaler = (*Detector)(nil)
)

// String returns the string representation of the UpdateType.
func (ut Detector) String() string {
	switch ut {
	case DetectorFSNotify:
		return "fsnotify"
	case DetectorPoll:
		return "poll"
	default:
		return fmt.Sprintf("Detector(%d)", ut)
	}
}

// MarshalText implements encoding.TextMarshaler.
func (ut Detector) MarshalText() (text []byte, err error) {
	return []byte(ut.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ut *Detector) UnmarshalText(text []byte) error {
	switch string(text) {
	case "":
		*ut = DetectorDefault
	case "fsnotify":
		*ut = DetectorFSNotify
	case "poll":
		*ut = DetectorPoll
	default:
		return fmt.Errorf("unrecognized detector %q, expected fsnotify or poll", string(text))
	}
	return nil
}

type fsNotify struct {
	opts    fsNotifyOptions
	watcher *fsnotify.Watcher
	cancel  context.CancelFunc
}

type fsNotifyOptions struct {
	Logger      log.Logger
	Filename    string
	UpdateCh    chan<- struct{} // Where to send detected updates to
	RewatchWait time.Duration   // How often to try to re-watch the file
}

// newFSNotify creates a new fsnotify detector which uses filesystem events to
// detect that a file has changed.
func newFSNotify(opts fsNotifyOptions) (*fsNotify, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := w.Add(opts.Filename); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	wd := &fsNotify{
		opts:    opts,
		watcher: w,
		cancel:  cancel,
	}

	go wd.wait(ctx)
	return wd, nil
}

func (fsn *fsNotify) wait(ctx context.Context) {
	rewatchTick := time.NewTicker(fsn.opts.RewatchWait)
	defer rewatchTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rewatchTick.C:
			// The fsnotify watcher may have removed our file from the watch list
			// (i.e., if the file got deleted).
			//
			// Continually re-adding it to the watcher acts as a fallback polling
			// mechanism, where we'll eventually rewatch the file once it gets added
			// again.
			err := fsn.watcher.Add(fsn.opts.Filename)
			if err != nil {
				level.Warn(fsn.opts.Logger).Log("msg", "failed re-watch file", "path", fsn.opts.Filename, "err", err)
			} else {
				fsn.forwardNotification()
			}
		case err := <-fsn.watcher.Errors:
			if err != nil {
				level.Warn(fsn.opts.Logger).Log("msg", "got error from fsnotify watcher; treating as file updated event", "err", err)
				fsn.forwardNotification()
			}
		case ev := <-fsn.watcher.Events:
			level.Debug(fsn.opts.Logger).Log("msg", "got fsnotify event", "path", ev.Name, "op", ev.Op.String())
			fsn.forwardNotification()
		}
	}
}

func (fsn *fsNotify) forwardNotification() {
	select {
	case fsn.opts.UpdateCh <- struct{}{}:
	default:
		// Already queued; no need to queue another event.
	}
}

func (fsn *fsNotify) Close() error {
	fsn.cancel()
	return fsn.watcher.Close()
}

type poller struct {
	opts   pollerOptions
	cancel context.CancelFunc
}

type pollerOptions struct {
	Filename      string
	UpdateCh      chan<- struct{} // Where to send detected updates to
	PollFrequency time.Duration
}

// newPoller creates a new poll-based file update detector.
func newPoller(opts pollerOptions) *poller {
	ctx, cancel := context.WithCancel(context.Background())

	pw := &poller{
		opts:   opts,
		cancel: cancel,
	}

	go pw.run(ctx)
	return pw
}

func (p *poller) run(ctx context.Context) {
	t := time.NewTicker(p.opts.PollFrequency)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// Always tell the component to re-check the file. This avoids situations
			// where the file changed without changing any of the stats (like modify
			// time).
			p.forwardNotification()
		}
	}
}

func (p *poller) forwardNotification() {
	select {
	case p.opts.UpdateCh <- struct{}{}:
	default:
		// Already queued; no need to queue another event.
	}
}

// Close terminates the poller.
func (p *poller) Close() error {
	p.cancel()
	return nil
}
