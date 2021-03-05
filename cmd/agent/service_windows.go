// +build windows

package main

import (
	"flag"
	"log"
	"os"
	"time"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/config"

	"golang.org/x/sys/windows/svc"
)

const ServiceName = "Grafana Cloud Agent"
const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

type AgentService struct{}

func (m *AgentService) Execute(args []string, serviceRequests <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Executable name and any command line parameters will be placed into os.args, this comes from
	// registry key `Computer\HKEY_LOCAL_MACHINE\SYSTEM\ControlSet001\Services\<servicename>\ImagePath`
	// oddly enough args is blank
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg, err := config.Load(fs, os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}
	// After this point we can use util_log.Logger and stop using the log package
	util_log.InitLogger(&cfg.Server)
	logger := util_log.Logger

	srv, err := NewEntryPoint(logger, cfg)
	if err != nil {
		level.Error(logger).Log("msg", "error creating the agent server entrypoint", "err", err)
		os.Exit(1)
	}
	exit := make(chan error)
	// Kick off the server in the background so that we can respond to status queries
	go srv.Start(exit)
	// Pause is not accepted
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-serviceRequests:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				srv.Stop()
				break loop
			case svc.Pause:
			case svc.Continue:
			default:
				srv.Stop()
				break loop
			}
		case err := <-exit:
			level.Error(logger).Log("msg", "error while running agent server entrypoint", "err", err)
			srv.Stop()
			break loop
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func IsWindowsService() bool {
	inService, err := svc.IsWindowsService()
	if inService == false || err != nil {
		return false
	}
	return true
}

func RunService() error {
	return svc.Run(ServiceName, &AgentService{})
}
