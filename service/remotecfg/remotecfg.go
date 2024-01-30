package remotecfg

import (
	"context"
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/go-kit/log"
	agentv1 "github.com/grafana/agent-remote-config/api/gen/proto/go/agent/v1"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/internal/agentseed"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/service"
	"github.com/grafana/agentre-remote-config/api/gen/proto/go/agent/v1/agentv1connect"
	commonconfig "github.com/prometheus/common/config"
)

var fnvHash hash.Hash32 = fnv.New32()

func getHash(in string) string {
	fnvHash.Write([]byte(in))
	defer fnvHash.Reset()

	return fmt.Sprintf("%x", fnvHash.Sum(nil))
}

// Service implements a service for remote configuration.
type Service struct {
	opts Options
	args Arguments

	ctrl service.Controller

	mut               sync.RWMutex
	asClient          agentv1connect.AgentServiceClient
	getConfigRequest  *connect.Request[agentv1.GetConfigRequest]
	ticker            *time.Ticker
	dataPath          string
	currentConfigHash string
}

// ServiceName defines the name used for the remotecfg service.
const ServiceName = "remotecfg"

// Options are used to configure the remotecfg service. Options are
// constant for the lifetime of the remotecfg service.
type Options struct {
	Logger      log.Logger // Where to send logs.
	StoragePath string     // Where to cache configuration on-disk.
}

// Arguments holds runtime settings for the remotecfg service.
type Arguments struct {
	URL              string                   `river:"url,attr,optional"`
	ID               string                   `river:"id,attr,optional"`
	Metadata         map[string]string        `river:"metadata,attr,optional"`
	PollFrequency    time.Duration            `river:"poll_frequency,attr,optional"`
	HTTPClientConfig *config.HTTPClientConfig `river:",squash"`
}

// GetDefaultArguments populates the default values for the Arguments struct.
func GetDefaultArguments() Arguments {
	return Arguments{
		ID:               agentseed.Get().UID,
		Metadata:         make(map[string]string),
		PollFrequency:    1 * time.Minute,
		HTTPClientConfig: config.CloneDefaultHTTPClientConfig(),
	}
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = GetDefaultArguments()
}

// Validate implements river.Validator.
func (a *Arguments) Validate() error {
	// We must explicitly Validate because HTTPClientConfig is squashed and it
	// won't run otherwise
	if a.HTTPClientConfig != nil {
		return a.HTTPClientConfig.Validate()
	}

	return nil
}

// Data includes information associated with the remotecfg service.
type Data struct {
}

// New returns a new instance of the remotecfg service.
func New(opts Options) (*Service, error) {
	basePath := filepath.Join(opts.StoragePath, ServiceName)
	err := os.MkdirAll(basePath, 0750)
	if err != nil {
		return nil, err
	}

	return &Service{
		opts:   opts,
		ticker: time.NewTicker(math.MaxInt64),
	}, nil
}

// Data returns an instance of [Data]. Calls to Data are cachable by the
// caller.
//
// Data must only be called after parsing command-line flags.
func (s *Service) Data() any {
	return map[string]string{}
}

// Definition returns the definition of the remotecfg service.
func (s *Service) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: Arguments{},
		DependsOn:  nil, // remotecfg has no dependencies.
	}
}

var _ service.Service = (*Service)(nil)

// Run implements [service.Service] and starts the remotecfg service. It will
// run until the provided context is canceled or there is a fatal error.
func (s *Service) Run(ctx context.Context, host service.Host) error {
	s.ctrl = host.NewController(ServiceName)

	// Run the service's own controller.
	go func() {
		s.ctrl.Run(ctx)
	}()

	// We're on the initial start-up of the service.
	// Let's try to read from the API, parse and load the response contents.
	// If either the request itself or parsing its contents fails, try to read
	// from the on-disk cache on a best-effort basis.
	var (
		gcr              *connect.Response[agentv1.GetConfigResponse]
		src              *flow.Source
		getErr, parseErr error
	)

	s.mut.RLock()
	gcr, getErr = s.asClient.GetConfig(ctx, s.getConfigRequest)
	if getErr == nil {
		src, parseErr = flow.ParseSource(ServiceName, []byte(gcr.Msg.GetContent()))
	}

	// Reading from the API succeeded, let's try to load the contents.
	if getErr == nil && parseErr == nil {
		err := s.ctrl.LoadSource(src, nil)
		if err != nil {
			level.Error(s.opts.Logger).Log("msg", "could not load the API response contents", "err", err)
		} else {
			s.currentConfigHash = getHash(gcr.Msg.Content)
		}
	} else {
		// Either the API call or parsing its contents failed, let's try the
		// on-disk cache.
		level.Info(s.opts.Logger).Log("msg", "falling back to the on-disk cache")
		b, err := os.ReadFile(s.dataPath)
		if err != nil {
			level.Error(s.opts.Logger).Log("msg", "could not read from the on-disk cache", "err", err)
		}
		if len(b) > 0 {
			src, err := flow.ParseSource(ServiceName, b)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "could not parse the on-disk cache contents", "err", err)
			} else {
				err = s.ctrl.LoadSource(src, nil)
				if err != nil {
					level.Error(s.opts.Logger).Log("msg", "could not load the on-disk cache contents", "err", err)
				} else {
					s.currentConfigHash = getHash(string(b))
				}
			}
		}
	}
	s.mut.RUnlock()

	for {
		select {
		case <-s.ticker.C:
			s.mut.RLock()
			gcr, err := s.asClient.GetConfig(ctx, s.getConfigRequest)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to fetch configuration from the API", "err", err)
				continue
			}

			newConfig := []byte(gcr.Msg.Content)

			// The polling loop got the same configuration, no need to reload.
			newConfigHash := getHash(gcr.Msg.Content)
			if s.currentConfigHash == newConfigHash {
				continue
			}

			// The polling loop got new configuration contents.
			src, err := flow.ParseSource(ServiceName, newConfig)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to parse configuration from the API", "err", err)
				continue
			}
			err = s.ctrl.LoadSource(src, nil)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to load configuration from the API", "err", err)
				continue
			}

			// If successful, flush to disk and keep a copy.
			err = os.WriteFile(s.dataPath, newConfig, 0750)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to flush remote_configuration contents the on-disk cache", "err", err)
			}
			s.mut.RUnlock()

			s.mut.Lock()
			s.currentConfigHash = newConfigHash
			s.mut.Unlock()
		case <-ctx.Done():
			s.ticker.Stop()
			return nil
		default:
		}
	}
}

// Update implements [service.Service] and applies settings.
func (s *Service) Update(newConfig any) error {
	newArgs := newConfig.(Arguments)

	// We either never set the block on the first place, or recently removed
	// it. Make sure we stop everything gracefully before returning.
	if newArgs.URL == "" {
		s.mut.Lock()
		defer s.mut.Unlock()
		s.ticker.Reset(math.MaxInt64)
		s.getConfigRequest = nil
		s.asClient = nil
		s.args.HTTPClientConfig = nil
		s.getConfigRequest = nil
		s.currentConfigHash = ""
		return nil
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	s.dataPath = filepath.Join(s.opts.StoragePath, ServiceName, getHash(newArgs.URL))
	s.ticker.Reset(newArgs.PollFrequency)

	if !reflect.DeepEqual(s.args.HTTPClientConfig, newArgs.HTTPClientConfig) {
		httpClient, err := commonconfig.NewClientFromConfig(*newArgs.HTTPClientConfig.Convert(), "remoteconfig")
		if err != nil {
			return err
		}
		s.asClient = agentv1connect.NewAgentServiceClient(
			httpClient,
			newArgs.URL,
		)
	}

	s.getConfigRequest = connect.NewRequest(&agentv1.GetConfigRequest{
		Id:       newArgs.ID,
		Metadata: newArgs.Metadata,
	})
	s.args = newArgs

	return nil
}
