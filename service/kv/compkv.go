package kv

import "go.etcd.io/bbolt"

type compkv struct {
	db         *KVDB
	bucketName string
}

func newCompKV(db *KVDB, bucket string) *compkv {
	c := &compkv{
		db:         db,
		bucketName: bucket,
	}
	// Go ahead and create the bucket.
	_ = c.db.db.Update(func(tx *bbolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte(c.bucketName))
		return nil
	})
	return c
}

// Put puts a value into the default bucket.
func (k *compkv) Put(key string, value []byte) error {
	return k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		return b.Put([]byte(key), value)
	})
}

// PutInBucket puts a value into the specified bucket, creating the bucket if it does not exist.
func (k *compkv) PutInBucket(bucket string, key string, value []byte) error {
	return k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		nb, err := b.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return nb.Put([]byte(key), value)
	})
}

// Get returns the value or nil if not found.
func (k *compkv) Get(key string) (_ []byte, _ bool, _ error) {
	var returnBytes []byte
	// This needs to be done in an update since we are creating the bucket if it doesnt exist.
	uerr := k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		// tempVal is only good for the life of the transaction so it needs to be copied.
		tempVal := b.Get([]byte(key))
		returnBytes = make([]byte, len(tempVal))
		copy(returnBytes, tempVal)
		return nil
	})
	return returnBytes, len(returnBytes) > 0, uerr
}

// GetFromBucket returns the value or nil if not found. The bucket will be created if it does not exist.
func (k *compkv) GetFromBucket(bucket string, key string) (_ []byte, _ bool, _ error) {
	var returnBytes []byte
	// This needs to be done in an update since we are creating the bucket if it doesnt exist.
	uerr := k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		nb, err := b.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		// tempVal is only good for the life of the transaction so it needs to be copied.
		tempVal := nb.Get([]byte(key))
		returnBytes = make([]byte, len(tempVal))
		copy(returnBytes, tempVal)
		return nil
	})
	return returnBytes, len(returnBytes) > 0, uerr
}

// Remove removes the value.
func (k *compkv) Remove(key string) {
	_ = k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		return b.Delete([]byte(key))
	})
}

// RemoveFromBucket removes the value and creates the bucket if the bucket does not exist.
func (k *compkv) RemoveFromBucket(bucket string, key string) {
	_ = k.db.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(k.bucketName))
		nb, _ := b.CreateBucketIfNotExists([]byte(bucket))
		return nb.Delete([]byte(key))
	})
}
