package configstore

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cortexproject/cortex/pkg/ring/kv"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

// Remote loads instance files from a remote KV store. The KV store
// can be swapped out in real time.
type Remote struct {
	log log.Logger
	reg *util.Unregisterer

	kvMut    sync.RWMutex
	kv       kv.Client
	reloadKV chan struct{}

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	configsMut sync.Mutex
	configsCh  chan WatchEvent
}

func NewRemote(l log.Logger, reg prometheus.Registerer, cfg kv.Config, enable bool) (*Remote, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	r := &Remote{
		log: l,
		reg: util.WrapWithUnregisterer(reg),

		reloadKV: make(chan struct{}, 1),

		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,

		configsCh: make(chan WatchEvent),
	}
	if err := r.ApplyConfig(cfg, enable); err != nil {
		return nil, fmt.Errorf("failed to apply config for config store: %w", err)
	}

	go r.run()
	return r, nil
}

// ApplyConfig applies the config for a kv client.
func (r *Remote) ApplyConfig(cfg kv.Config, enable bool) error {
	r.kvMut.Lock()
	defer r.kvMut.Unlock()

	if r.cancelCtx.Err() != nil {
		return fmt.Errorf("remote store already stopped")
	}

	// Unregister all metrics that the previous kv may have registered.
	r.reg.UnregisterAll()

	if !enable {
		r.setClient(nil)
		return nil
	}

	cli, err := kv.NewClient(cfg, GetCodec(), kv.RegistererWithKVName(r.reg, "agent_configs"))
	if err != nil {
		return fmt.Errorf("failed to create kv client: %w", err)
	}

	r.setClient(cli)
	return nil
}

// setClient sets the active client and notifies run to restart the
// kv watcher.
func (r *Remote) setClient(client kv.Client) {
	r.kv = client
	r.reloadKV <- struct{}{}
}

func (r *Remote) run() {
	var (
		kvContext context.Context
		kvCancel  context.CancelFunc
	)

Outer:
	for {
		select {
		case <-r.cancelCtx.Done():
			break Outer
		case <-r.reloadKV:
			r.kvMut.RLock()
			kv := r.kv
			r.kvMut.RUnlock()

			if kvCancel != nil {
				kvCancel()
			}
			kvContext, kvCancel = context.WithCancel(r.cancelCtx)
			go r.watchKV(kvContext, kv)
		}
	}

	if kvCancel != nil {
		kvCancel()
	}
}

func (r *Remote) watchKV(ctx context.Context, client kv.Client) {
	// Edge case: client was unset, nothing to do here.
	if client == nil {
		level.Info(r.log).Log("msg", "not watching the KV, none set")
		return
	}

	client.WatchPrefix(ctx, "", func(key string, v interface{}) bool {
		if ctx.Err() != nil {
			return false
		}

		r.configsMut.Lock()
		defer r.configsMut.Unlock()

		switch {
		case v == nil:
			r.configsCh <- WatchEvent{Key: key, Config: nil}
		default:
			cfg, err := instance.UnmarshalConfig(strings.NewReader(v.(string)))
			if err != nil {
				level.Error(r.log).Log("msg", "could not unmarshal config from store", "name", key, "err", err)
				break
			}

			r.configsCh <- WatchEvent{Key: key, Config: cfg}
		}

		return true
	})
}

func (r *Remote) List(ctx context.Context) ([]string, error) {
	r.kvMut.RLock()
	defer r.kvMut.RUnlock()
	if r.kv == nil {
		return nil, ErrNotConnected
	}

	return r.kv.List(ctx, "")
}

func (r *Remote) Get(ctx context.Context, key string) (instance.Config, error) {
	r.kvMut.RLock()
	defer r.kvMut.RUnlock()
	if r.kv == nil {
		return instance.Config{}, ErrNotConnected
	}

	v, err := r.kv.Get(ctx, key)
	if err != nil {
		return instance.Config{}, fmt.Errorf("failed to get config %s: %w", key, err)
	} else if v == nil {
		return instance.Config{}, NotExistError{Key: key}
	}

	cfg, err := instance.UnmarshalConfig(strings.NewReader(v.(string)))
	if err != nil {
		return instance.Config{}, fmt.Errorf("failed to unmarshal config %s: %w", key, err)
	}
	return *cfg, nil
}

func (r *Remote) Put(ctx context.Context, c instance.Config) (bool, error) {
	// We need to use a write lock here since two Applies can't run concurrently
	// (given the current need to perform a store-wide validation.)
	r.kvMut.Lock()
	defer r.kvMut.Unlock()
	if r.kv == nil {
		return false, ErrNotConnected
	}

	bb, err := instance.MarshalConfig(&c, false)
	if err != nil {
		return false, fmt.Errorf("failed to marshal config: %w", err)
	}

	cfgCh, err := r.all(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check validity of config: %w", err)
	}
	if err := checkUnique(cfgCh, &c); err != nil {
		return false, fmt.Errorf("failed to check uniqueness of config: %w", err)
	}

	var created bool
	err = r.kv.CAS(ctx, c.Name, func(in interface{}) (out interface{}, retry bool, err error) {
		// The configuration is new if there's no previous value from the CAS
		created = (in == nil)
		return string(bb), false, nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to put config: %w", err)
	}
	return created, nil
}

func (r *Remote) Delete(ctx context.Context, key string) error {
	r.kvMut.RLock()
	defer r.kvMut.RUnlock()
	if r.kv == nil {
		return ErrNotConnected
	}

	// Some KV stores don't return an error if something failed to be
	// deleted, so we'll try to get it first. This isn't perfect, and
	// it may fail, so we'll silently ignore any errors here unless
	// we know for sure the config doesn't exist.
	v, err := r.kv.Get(ctx, key)
	if err != nil {
		level.Warn(r.log).Log("msg", "error validating key existence for deletion", "err", err)
	} else if v == nil {
		return NotExistError{Key: key}
	}

	err = r.kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("error deleting configuration: %w", err)
	}

	return nil
}

// All retrieves the set of all configs in the store.
func (r *Remote) All(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error) {
	r.kvMut.RLock()
	defer r.kvMut.RUnlock()
	return r.all(ctx, keep)
}

// all can only be called if the kvMut lock is already held.
func (r *Remote) all(ctx context.Context, keep func(key string) bool) (<-chan instance.Config, error) {
	if r.kv == nil {
		return nil, ErrNotConnected
	}

	keys, err := r.kv.List(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	ch := make(chan instance.Config)

	var wg sync.WaitGroup
	wg.Add(len(keys))
	go func() {
		wg.Wait()
		close(ch)
	}()

	for _, key := range keys {
		go func(key string) {
			defer wg.Done()

			if keep != nil && !keep(key) {
				level.Debug(r.log).Log("msg", "skipping key that was filtered out", "key", key)
				return
			}

			// TODO(rfratto): retries might be useful here
			v, err := r.kv.Get(ctx, key)
			if err != nil {
				level.Error(r.log).Log("msg", "failed to get config with key", "key", key, "err", err)
				return
			} else if v == nil {
				// Config was deleted since we called list, skip it.
				level.Debug(r.log).Log("msg", "skipping key that was deleted after list was called", "key", key)
				return
			}

			cfg, err := instance.UnmarshalConfig(strings.NewReader(v.(string)))
			if err != nil {
				level.Error(r.log).Log("msg", "failed to unmarshal config from store", "key", key, "err", err)
				return
			}
			ch <- *cfg
		}(key)
	}

	return ch, nil
}

// Watch watches the Store for changes.
func (r *Remote) Watch() <-chan WatchEvent {
	return r.configsCh
}

// Close closes the Remote store.
func (r *Remote) Close() error {
	r.kvMut.Lock()
	defer r.kvMut.Unlock()
	r.cancelFunc()
	return nil
}
