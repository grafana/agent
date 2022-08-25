//go:build windows
// +build windows

package initiate

//We must keep this package clear of importing any large packages so that we can initialise the windows service ASAP
import (
	"log"

	"golang.org/x/sys/windows/svc"
)

// Default name for the Grafana Agent under Windows
const (
	ServiceName = "Grafana Agent"
)

func init() {
	// If Windows is trying to run as a service, go through that
	// path instead.
	if IsWindowsService() {
		go func() {
			err := svc.Run(ServiceName, &AgentService{stopCh: ServiceExit})
			if err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
		}()
	}
}

const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

// Channel to inform server of service stop request
var ServiceExit = make(chan bool)

// AgentService runs the Grafana Agent as a service.
type AgentService struct {
	stopCh chan<- bool
}

// Execute starts the AgentService.
func (m *AgentService) Execute(args []string, serviceRequests <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Pause is not accepted, we immediately set the service as running and trigger the entrypoint load in the background
	// this is because the WAL is reloaded and the timeout for a windows service starting is 30 seconds. In this case
	// the service is running but Agent may still be starting up reading the WAL and doing other operations.

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case c := <-serviceRequests:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				//We need to send a message back to stop the process if this happens
				m.stopCh <- true
				break loop
			case svc.Pause:
			case svc.Continue:
			default:
				log.Fatalf("unexpected control request #%d", c)
			}
		}
	}
	// There is a chance the entrypoint may not be setup yet, in that case we don't want to stop.
	// Since it is in another go func it may start after this has returned, in either case the program
	// will exit.
	//if ep != nil {
	//	ep.Stop()
	//}
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
