package java

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/component/pyroscope/java/asprof"
	"github.com/grafana/agent/pkg/flow/logging/level"
	jfrpprof "github.com/grafana/jfr-parser/pprof"
	jfrpprofPyroscope "github.com/grafana/jfr-parser/pprof/pyroscope"
	"github.com/prometheus/prometheus/model/labels"
	gopsutil "github.com/shirou/gopsutil/v3/process"
)

// we extract to /tmp so all users can read it, but only root write, see asprof.extractPerm
var profiler = asprof.NewProfiler("/tmp")

const spyName = "grafana-agent.java"

type profilingLoop struct {
	logger log.Logger
	output *pyroscope.Fanout

	cfg       ProfilingConfig
	wg        sync.WaitGroup
	mutex     sync.Mutex
	pid       int
	target    discovery.Target
	cancel    context.CancelFunc
	error     error
	dist      *asprof.Distribution
	jfrFile   string
	startTime time.Time
}

func newProfilingLoop(pid int, target discovery.Target, logger log.Logger, output *pyroscope.Fanout, cfg ProfilingConfig) *profilingLoop {
	ctx, cancel := context.WithCancel(context.Background())
	dist, err := asprof.DistributionForProcess(pid)
	p := &profilingLoop{
		logger:  log.With(logger, "pid", pid),
		output:  output,
		pid:     pid,
		target:  target,
		cancel:  cancel,
		dist:    dist,
		jfrFile: fmt.Sprintf("/tmp/asprof-%d-%d.jfr", os.Getpid(), pid),
		cfg:     cfg,
	}
	_ = level.Debug(p.logger).Log("msg", "new process", "target", fmt.Sprintf("%+v", target))

	if err != nil {
		p.onError(fmt.Errorf("failed to select dist for pid %d: %w", pid, err))
		return p
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.loop(ctx)
	}()
	return p
}

func (p *profilingLoop) loop(ctx context.Context) {
	if err := profiler.CopyLib(p.dist, p.pid); err != nil {
		p.onError(fmt.Errorf("failed to copy libasyncProfiler.so: %w", err))
		return
	}
	defer p.stop()
	sleep := func() bool {
		timer := time.NewTimer(p.interval())
		select {
		case <-timer.C:
			return false
		case <-ctx.Done():
			return true
		}
	}
	for {
		err := p.start()
		if err != nil {
			//  could happen when agent restarted - [ERROR] Profiler already started\n
			p.onError(fmt.Errorf("failed to start: %w", err))
			if !p.alive() {
				_ = level.Error(p.logger).Log("msg", "process is not alive")
				return
			}
		}
		done := sleep()
		if done {
			return
		}
		err = p.reset()
		if err != nil {
			p.onError(fmt.Errorf("failed to reset: %w", err))
			if !p.alive() {
				_ = level.Error(p.logger).Log("msg", "process is not alive")
				return
			}
		}
	}
}

func (p *profilingLoop) reset() error {
	jfrFile := asprof.ProcessPath(p.jfrFile, p.pid)
	startTime := p.startTime
	endTime := time.Now()
	p.startTime = endTime
	defer func() {
		os.Remove(jfrFile)
	}()

	err := p.stop()
	if err != nil {
		return fmt.Errorf("failed to stop : %w", err)
	}
	jfrBytes, err := os.ReadFile(jfrFile)
	if err != nil {
		return fmt.Errorf("failed to read jfr file: %w", err)
	}
	_ = level.Debug(p.logger).Log("msg", "jfr file read", "len", len(jfrBytes))

	return p.push(jfrBytes, startTime, endTime)
}
func (p *profilingLoop) push(jfrBytes []byte, startTime time.Time, endTime time.Time) error {
	profiles, err := jfrpprof.ParseJFR(jfrBytes, &jfrpprof.ParseInput{
		StartTime:  startTime,
		EndTime:    endTime,
		SampleRate: int64(p.cfg.SampleRate),
	}, new(jfrpprof.LabelsSnapshot))
	if err != nil {
		return fmt.Errorf("failed to parse jfr: %w", err)
	}
	for _, req := range profiles.Profiles {
		metric := req.Metric
		sz := req.Profile.SizeVT()
		l := log.With(p.logger, "metric", metric, "sz", sz)
		ls := labels.NewBuilder(nil)
		for _, l := range jfrpprofPyroscope.Labels(p.target, profiles.JFREvent, req.Metric, "", spyName) {
			ls.Set(l.Name, l.Value)
		}
		profile, err := req.Profile.MarshalVT()
		samples := []*pyroscope.RawSample{{RawProfile: profile}}
		err = p.output.Appender().Append(context.Background(), ls.Labels(), samples)
		if err != nil {
			_ = l.Log("msg", "failed to push jfr", "err", err)
			continue
		}
		_ = l.Log("msg", "pushed jfr-pprof")
	}
	return nil
}

func (p *profilingLoop) start() error {
	p.startTime = time.Now()
	argv := make([]string, 0, 14)
	argv = append(argv,
		"-f", p.jfrFile,
		"-o", "jfr",
	)
	if p.cfg.CPU {
		argv = append(argv, "-e", "itimer")
	}
	if p.cfg.Alloc != "" {
		argv = append(argv, "--alloc", p.cfg.Alloc)
	}
	if p.cfg.Lock != "" {
		argv = append(argv, "--lock", p.cfg.Lock)
	}
	argv = append(argv,
		"start",
		"--timeout", strconv.Itoa(int(p.interval().Seconds())),
		strconv.Itoa(p.pid),
	)

	_ = level.Debug(p.logger).Log("cmd", fmt.Sprintf("%s %s", p.dist.AsprofPath(), strings.Join(argv, " ")))
	stdout, stderr, err := profiler.Execute(p.dist, argv)
	if err != nil {
		return fmt.Errorf("asprof failed to run: %w %s %s", err, stdout, stderr)
	}
	return nil
}

func (p *profilingLoop) stop() error {
	argv := []string{
		"stop",
		"-o", "jfr",
		strconv.Itoa(p.pid),
	}
	_ = level.Debug(p.logger).Log("msg", "asprof", "cmd", fmt.Sprintf("%s %s", p.dist.AsprofPath(), strings.Join(argv, " ")))
	stdout, stderr, err := profiler.Execute(p.dist, argv)
	if err != nil {
		_ = level.Error(p.logger).Log("msg", "asprof failed to run", "err", err, "stdout", stdout, "stderr", stderr)
		return fmt.Errorf("asprof failed to run: %w", err)
	}
	_ = level.Debug(p.logger).Log("msg", "asprof stopped", "stdout", stdout, "stderr", stderr) //todo this should be empty
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

func (p *profilingLoop) getTarget() discovery.Target {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.target
}

func (p *profilingLoop) alive() bool {
	exists, err := gopsutil.PidExists(int32(p.pid))
	return err != nil && exists
}
