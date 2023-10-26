package kv

import "github.com/grafana/agent/component"

type KVProvider interface {
	// GetKV returns a kv implementation sandboxed to the component id.
	GetKV(opts component.Options) KV
}

type KV interface {
	// Put puts a value.
	Put(key string, value []byte) error

	// Get returns the value or nil if not found.
	Get(key string) ([]byte, bool, error)

	// Remove removes the value.
	Remove(key string)

	// PutInBucket puts a value into the specified bucket, creating the bucket if it does not exist.
	PutInBucket(bucket string, key string, value []byte) error

	// GetFromBucket   returns the value or nil if not found. The bucket will be created if it does not exist.
	GetFromBucket(bucket string, key string) ([]byte, bool, error)

	// RemoveFromBucket removes the value and creates the bucket if the bucket does not exist.
	RemoveFromBucket(bucket string, key string)
}
