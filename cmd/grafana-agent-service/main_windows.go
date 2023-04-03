package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/sys/windows/svc"
)

const serviceName = "Grafana Agent Flow"

func main() {
	logger, err := newLogger()
	if err != nil {
		// Ideally the logger never fails to be created, since if it does, there's
		// nowhere to send the failure to.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	managerConfig, err := loadConfig()
	if err != nil {
		level.Error(logger).Log("msg", "failed to run service", "err", err)
		os.Exit(1)
	}

	cfg := serviceManagerConfig{
		Path: managerConfig.ServicePath,
		Args: managerConfig.Args,
		Dir:  managerConfig.WorkingDirectory,

		// Send logs directly to the event logger.
		Stdout: logger,
		Stderr: logger,
	}

	as := &agentService{logger: logger, cfg: cfg}
	if err := svc.Run(serviceName, as); err != nil {
		level.Error(logger).Log("msg", "failed to run service", "err", err)
		os.Exit(1)
	}
}

type agentService struct {
	logger log.Logger
	cfg    serviceManagerConfig
}

const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

func (as *agentService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	defer func() {
		s <- svc.Status{State: svc.Stopped}
	}()

	s <- svc.Status{State: svc.StartPending}

	// Run the serviceManager. When the service is exiting, wait until the
	// serviceManager exits completely.
	{
		sm := newServiceManager(as.logger, as.cfg)

		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			defer wg.Done()
			sm.Run(ctx)
		}()
	}

	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	defer func() {
		s <- svc.Status{State: svc.StopPending}
	}()

	for {
		req := <-r
		switch req.Cmd {
		case svc.Interrogate:
			s <- req.CurrentStatus
		case svc.Pause, svc.Continue:
			// no-op
		default:
			// Every other command should terminate the service.
			return false, 0
		}
	}
}
