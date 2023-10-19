package windowsevent

import (
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
)

type KVDB struct {
	db *bbolt.DB
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
