package ha

import (
	"context"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grafana/agent/pkg/agentproto"
)

// Reshard initiates an entire reshard of the current HA scraping service instance.
// All configs will be reloaded from the KV store and the scraping service instance
// will see what should be managed locally.
//
// Satisfies agentproto.ScrapingServiceServer.
func (s *Server) Reshard(ctx context.Context, _ *agentproto.ReshardRequest) (_ *empty.Empty, err error) {
	s.configManagerMut.Lock()
	defer s.configManagerMut.Unlock()

	start := time.Now()
	defer func() {
		success := "1"
		if err != nil {
			success = "0"
		}
		reshardDuration.WithLabelValues(success).Observe(time.Since(start).Seconds())
	}()

	var (
		// configs found in the KV store. currentConfigs - discoveredConfigs is the
		// list of configs that was removed from the KV store since the last reshard.
		discoveredConfigs = map[string]struct{}{}
	)

	configCh, err := s.store.All(ctx, func(key string) bool {
		owns, err := s.owns(key)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to detect if key was owned", "key", key, "err", err)
			return false
		}
		return owns
	})
	if err != nil {
		level.Error(s.logger).Log("msg", "failed getting config list when resharding", "err", err)
		return nil, err
	}
	for ch := range configCh {
		if s.applyConfig(ch.Name, &ch) {
			discoveredConfigs[ch.Name] = struct{}{}
		}
	}

	// Find the set of configs that disappeared from AllConfigs from the last
	// time this ran and remove them.
	for runningConfig := range s.configs {
		_, keyInStore := discoveredConfigs[runningConfig]
		if keyInStore {
			continue
		}

		level.Info(s.logger).Log("msg", "deleting config removed from store", "name", runningConfig)
		err := s.im.DeleteConfig(runningConfig)
		if err != nil {
			level.Error(s.logger).Log("msg", "failed to delete stale config", "err", err)
		}
	}

	// Update the set of running configs to what we last got from the server.
	s.configs = discoveredConfigs

	return &empty.Empty{}, nil
}

// owns checks to see if a config name is owned by this Server. owns will
// return an error if the ring is empty or if there aren't enough
// healthy nodes.
func (s *Server) owns(key string) (bool, error) {
	rs, err := s.ring.Get(keyHash(key), ring.Write, nil, nil, nil)
	if err != nil {
		return false, err
	}
	for _, r := range rs.Ingesters {
		if r.Addr == s.addr {
			return true, nil
		}
	}
	return false, nil
}

func keyHash(key string) uint32 {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// ReadRing is a subset of the Cortex ring.ReadRing interface with only the
// functionality used by the HA server.
type ReadRing interface {
	http.Handler

	Get(key uint32, op ring.Operation, bufDescs []ring.InstanceDesc, bufHosts, bufZones []string) (ring.ReplicationSet, error)
	GetAllHealthy(op ring.Operation) (ring.ReplicationSet, error)
}
