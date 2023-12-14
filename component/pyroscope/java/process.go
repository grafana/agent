package java

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/component/pyroscope/java/asprof"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

var profiler = asprof.NewProfiler("/tmp")

type profilingLoop struct {
	logger log.Logger
	output *pyroscope.Fanout

	wg      sync.WaitGroup
	mutex   sync.Mutex
	pid     int
	target  discovery.Target
	cancel  context.CancelFunc
	error   error
	dist    asprof.Distribution
	jfrFile asprof.File
}

func newProcess(pid int, target discovery.Target, logger log.Logger, output *pyroscope.Fanout) *profilingLoop {
	ctx, cancel := context.WithCancel(context.Background())
	p := &profilingLoop{
		logger: log.With(logger, "pid", pid),
		output: output,
		pid:    pid,
		target: target,
		cancel: cancel,
		dist:   asprof.Glibc, //todo
		jfrFile: asprof.File{
			Path: fmt.Sprintf("/tmp/asprof-%d-%d.jfr", os.Getpid(), pid),
			PID:  pid,
		},
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.loop(ctx)
	}()
	return p
}

func (p *profilingLoop) loop(ctx context.Context) {
	if err := profiler.CopyLib(asprof.Glibc, p.pid); err != nil {
		p.onError(fmt.Errorf("failed to copy libasyncProfiler.so: %w", err))
		return
	}
	defer p.stop()

	timer := time.NewTimer(p.interval())
	err := p.start()
	if err != nil {
		p.onError(fmt.Errorf("failed to start asprof: %w", err))
		return
	}
	for {
		select {
		case <-timer.C:
			if err := p.reset(); err != nil {
				p.onError(fmt.Errorf("failed to reset asprof: %w", err))
				return
			}
			timer.Reset(p.interval())
		case <-ctx.Done():
			return
		}
	}
}

func (p *profilingLoop) reset() error {
	_ = level.Debug(p.logger).Log("msg", "timer tick")
	err := p.stop()
	if err != nil {
		return fmt.Errorf("failed to stop asprof: %w", err)
	}
	jfr, err := p.jfrFile.Read()
	if err != nil {
		return fmt.Errorf("failed to read jfr file: %w", err)
	}
	err = p.jfrFile.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete jfr file: %w", err)
	}
	err = p.start()
	if err != nil {
		return fmt.Errorf("failed to start asprof: %w", err)
	}
	_ = level.Debug(p.logger).Log("msg", "jfr file read", "len", len(jfr))
	return nil
}

func (p *profilingLoop) start() error {
	stdout, stderr, err := profiler.Execute(p.dist, []string{
		"-f", p.jfrFile.Path,
		"-o", "jfr",
		"-e", "itimer",
		"start",
		"--timeout", strconv.Itoa(int(p.interval().Seconds())),
		strconv.Itoa(p.pid),
	})
	if err != nil {
		_ = level.Error(p.logger).Log("msg", "asprof failed to run", "err", err, "stdout", stdout, "stderr", stderr)
		return fmt.Errorf("asprof failed to run: %w", err)
	}
	return nil
}

func (p *profilingLoop) stop() error {
	stdout, stderr, err := profiler.Execute(p.dist, []string{
		"stop",
		strconv.Itoa(p.pid),
	})
	if err != nil {
		_ = level.Error(p.logger).Log("msg", "asprof failed to run", "err", err, "stdout", stdout, "stderr", stderr)
		return fmt.Errorf("asprof failed to run: %w", err)
	}
	return nil
}

func (p *profilingLoop) update(target discovery.Target) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.target = target
}

// Close stops profiling this profilingLoop
func (p *profilingLoop) Close() error {
	p.cancel()
	p.wg.Wait()
	return nil
}

func (p *profilingLoop) onError(err error) {
	_ = level.Error(p.logger).Log("err", err)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.error = err
}

func (p *profilingLoop) interval() time.Duration {
	return time.Second * 15 // todo
}
