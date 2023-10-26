package kv

import (
	"context"
	"os"
	"path/filepath"

	"github.com/grafana/agent/component"
	flow_service "github.com/grafana/agent/service"
	"go.etcd.io/bbolt"
)

const ServiceName = "kv"

type KVDB struct {
	db *bbolt.DB
}

type Arguments struct{}

var _ flow_service.Service = (*KVDB)(nil)

// Definition returns the Definition of the Service.
// Definition must always return the same value across all
// calls.
func (kv *KVDB) Definition() flow_service.Definition {
	return flow_service.Definition{
		Name:       ServiceName,
		ConfigType: Arguments{},
		DependsOn:  nil,
	}
}

// Run starts a Service. Run must block until the provided
// context is canceled. Returning an error should be treated
// as a fatal error for the Service.
func (kv *KVDB) Run(ctx context.Context, host flow_service.Host) error {
	<-ctx.Done()
	return nil
}

// Update updates a Service at runtime. Update is never
// called if [Definition.ConfigType] is nil. newConfig will
// be the same type as ConfigType; if ConfigType is a
// pointer to a type, newConfig will be a pointer to the
// same type.
//
// Update will be called once before Run, and may be called
// while Run is active.
func (kv *KVDB) Update(newConfig any) error {
	return nil
}

// Data returns the Data associated with a Service. Data
// must always return the same value across multiple calls,
// as callers are expected to be able to cache the result.
//
// Data may be invoked before Run.
func (kv *KVDB) Data() any {
	return kv
}

// GetKV returns a kv implementation sandboxed to the component id.
func (kv *KVDB) GetKV(opts component.Options) KV {
	return newCompKV(kv, opts.ID)
}

// NewKVDB creates a wrapper around bbolt.
func NewKVDB(path string) (*KVDB, error) {
	_ = os.MkdirAll(filepath.Dir(path), 0600)
	bdb, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &KVDB{db: bdb}, nil
}

// Put writes the value in a specific bucket, creating it if it doesnt exist.
func (kv *KVDB) Put(bucket string, key string, value []byte) error {
	return kv.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

// Get returns the value which is nil if nothing found, true if found and any error.
// If the bucket does not exist then it will be created.
func (kv *KVDB) Get(bucket string, key string) ([]byte, bool, error) {
	var nv []byte
	err := kv.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		v := b.Get([]byte(key))
		if v != nil {
			// Have to copy v since it is reused once update has ended.
			nv = make([]byte, len(v))
			copy(nv, v)
		}
		return nil
	})
	return nv, nv != nil, err
}

func (kv *KVDB) Close() error {
	return kv.db.Close()
}
