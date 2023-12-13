package java

import (
	"context"
	_ "embed"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

type process struct {
	logger log.Logger
	output *pyroscope.Fanout

	wg     sync.WaitGroup
	mutex  sync.Mutex
	pid    int
	target discovery.Target
	cancel context.CancelFunc
	error  error
}

func newProcess(pid int, target discovery.Target, logger log.Logger, output *pyroscope.Fanout) *process {
	ctx, cancel := context.WithCancel(context.Background())
	p := &process{
		logger: log.With(logger, "pid", pid),
		output: output,
		pid:    pid,
		target: target,
		cancel: cancel,
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.loop(ctx)
	}()
	return p
}

func (p *process) loop(ctx context.Context) {
	err := p.installAsyncProfiler()
	if err != nil {
		p.onError(err)
		return
	}
	defer func() {
		err := p.uninstallAsyncProfiler()
		if err != nil {
			p.onError(err)
		}
	}()
	timer := time.NewTimer(p.interval())
	for {
		select {
		case <-timer.C:
			timer.Reset(time.Second * 15)
			_ = level.Debug(p.logger).Log("msg", "timer tick")
		case <-ctx.Done():
			return
		}
	}
}

func (p *process) update(target discovery.Target) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.target = target
}

func (p *process) stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *process) onError(err error) {
	_ = level.Error(p.logger).Log("err", err)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.error = err
}

func (p *process) installAsyncProfiler() error {
	//dstDir := p.profilerDir()
	return nil
}

// todo do a checksum before executing
func (p *process) profilerDir() string {
	return fmt.Sprintf("/proc/%d/root/tmp/asprof-%d", p.pid, p.pid)
}

func (p *process) uninstallAsyncProfiler() error {
	return nil
}

func (p *process) interval() time.Duration {
	return time.Second * 15 // todo
}

//go:embed async-profiler-3.0-ea-linux-x64.tar.gz
var asprof []byte
