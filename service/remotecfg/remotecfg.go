package remotecfg

import (
	"context"
	"fmt"
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
	"github.com/grafana/agent-remote-config/api/gen/proto/go/agent/v1/agentv1connect"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/internal/agentseed"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/agent/service"
	commonconfig "github.com/prometheus/common/config"
)

func getHash(in []byte) string {
	fnvHash := fnv.New32()

	fnvHash.Write(in)
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

// New returns a new instance of the remotecfg service.
func New(opts Options) (*Service, error) {
	basePath := filepath.Join(opts.StoragePath, ServiceName)
	err := os.MkdirAll(basePath, 0750)
	if err != nil {
		return nil, err
	}

	return &Service{
		opts:     opts,
		asClient: noopClient{},
		ticker:   time.NewTicker(math.MaxInt64),
	}, nil
}

// Data does not expose anything for the remotecfg service during runtime.
func (s *Service) Data() any {
	return nil
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

	// Let's try to read from the API or the local cache as a fallback.
	var b1, b2 []byte
	b1, err := s.getAPIConfig()
	b2, err2 := s.getCachedConfig()

	if err != nil {
		level.Debug(s.opts.Logger).Log("msg", "failed to get any configuration during startup", "apiErr", err, "cacheErr", err2)
	}

	err = s.parseAndLoad(b1)
	if err != nil {
		err = s.parseAndLoad(b2)
	}

	if err != nil {
		level.Error(s.opts.Logger).Log("msg", "failed to load remote cfg during startup", "err", err)
	}

	// Run the service's own controller.
	go func() {
		s.ctrl.Run(ctx)
	}()

	for {
		select {
		case <-s.ticker.C:
			b, err := s.getAPIConfig()
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to fetch configuration from the API", "err", err)
				continue
			}

			// The polling loop got the same configuration, no need to reload.
			newConfigHash := getHash(b)
			if s.getCfgHash() == newConfigHash {
				level.Debug(s.opts.Logger).Log("msg", "skipping to the next polling loop")
				continue
			}

			// We have a new configuration let's try to load it
			err = s.parseAndLoad(b)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to load configuration from the API", "err", err)
				continue
			}

			level.Info(s.opts.Logger).Log("msg", "new remote configuration loaded successfully")

			// If successful, flush to disk and keep a copy.
			err = os.WriteFile(s.dataPath, b, 0750)
			if err != nil {
				level.Error(s.opts.Logger).Log("msg", "failed to flush remote configuration contents the on-disk cache", "err", err)
			}
			s.setCfgHash(newConfigHash)
		case <-ctx.Done():
			s.ticker.Stop()
			return nil
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
		s.ticker.Reset(math.MaxInt64)
		s.asClient = noopClient{}
		s.args.HTTPClientConfig = config.CloneDefaultHTTPClientConfig()
		s.mut.Unlock()

		s.setCfgHash("") // TODO(@tpaschalis) Should we clear this out?
		return nil
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	s.dataPath = filepath.Join(s.opts.StoragePath, ServiceName, getHash([]byte(newArgs.URL)))
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

	s.args = newArgs

	return nil
}

func (s *Service) getAPIConfig() ([]byte, error) {
	s.mut.RLock()
	req := connect.NewRequest(&agentv1.GetConfigRequest{
		Id:       s.args.ID,
		Metadata: s.args.Metadata,
	})
	client := s.asClient
	s.mut.RUnlock()

	gcr, err := client.GetConfig(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return []byte(gcr.Msg.GetContent()), nil
}

func (s *Service) getCachedConfig() ([]byte, error) {
	s.mut.RLock()
	p := s.dataPath
	s.mut.RUnlock()

	return os.ReadFile(p)
}

func (s *Service) parseAndLoad(b []byte) error {
	s.mut.RLock()
	ctrl := s.ctrl
	s.mut.RUnlock()

	if len(b) == 0 {
		return nil
	}
	src, err := flow.ParseSource(ServiceName, b)
	if err != nil {
		return err
	}

	err = ctrl.LoadSource(src, nil)
	if err != nil {
		return err
	}

	s.setCfgHash(getHash(b))
	return nil
}

func (s *Service) getCfgHash() string {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.currentConfigHash
}

func (s *Service) setCfgHash(h string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.currentConfigHash = h
}
