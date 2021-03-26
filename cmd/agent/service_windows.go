// +build windows

package main

import (
	"flag"
	"log"
	"os"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/util"

	"golang.org/x/sys/windows/svc"
)

const ServiceName = "Grafana Agent"
const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

// AgentService runs the Grafana Agent as a service.
type AgentService struct{}

// Execute starts the AgentService.
func (m *AgentService) Execute(args []string, serviceRequests <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Executable name and any command line parameters will be placed into os.args, this comes from
	// registry key `Computer\HKEY_LOCAL_MACHINE\SYSTEM\ControlSet001\Services\<servicename>\ImagePath`
	// oddly enough args is blank

	reloader := func() (*config.Config, error) {
		fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		return config.Load(fs, os.Args[1:])
	}
	cfg, err := reloader()
	if err != nil {
		log.Fatalln(err)
	}

	// After this point we can start using go-kit logging.
	logger := util.NewLogger(&cfg.Server)
	util_log.Logger = logger

	ep, err := NewEntrypoint(logger, cfg, reloader)
	if err != nil {
		level.Error(logger).Log("msg", "error creating the agent server entrypoint", "err", err)
		os.Exit(1)
	}
	entrypointExit := make(chan error)
	// Kick off the server in the background so that we can respond to status queries
	go func() {
		entrypointExit <- ep.Start()
	}()
	// Pause is not accepted
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-serviceRequests:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				ep.Stop()
				break loop
			case svc.Pause:
			case svc.Continue:
			default:
				ep.Stop()
				break loop
			}
		case err := <-entrypointExit:
			level.Error(logger).Log("msg", "error while running agent server entrypoint", "err", err)
			ep.Stop()
			break loop
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

// IsWindowsService returns whether the current process is running as a Windows
// Service. On non-Windows platforms, this always returns false.
func IsWindowsService() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// RunService runs the current process as a Windows servce. On non-Windows platforms,
// this is always a no-op.
func RunService() error {
	return svc.Run(ServiceName, &AgentService{})
}
