package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"github.com/go-kit/log"
	"testing"
	"time"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	db, err := makeDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	tc := &testCompany{CompanyID: 99}
	buffer := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	err = enc.Encode(tc)
	require.NoError(t, err)
	id, err := db.WriteValueWithAutokey(buffer.Bytes(), 10, 1, 0)
	require.NoError(t, err)

	valByte, tt, found, err := db.GetValueByKey(id)
	require.True(t, tt == 10)
	require.NoError(t, err)
	require.True(t, found)
	buffer = bytes.NewBuffer(valByte)
	dec := gob.NewDecoder(buffer)
	tc1 := &testCompany{}
	err = dec.Decode(tc1)
	require.NoError(t, err)
	require.True(t, tc1.CompanyID == tc.CompanyID)
	require.True(t, len(db.keyCache.keys()) == 1)
}

func TestTTL(t *testing.T) {
	db, err := makeDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	tc := &testCompany{CompanyID: 99}
	buffer := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	err = enc.Encode(tc)
	require.NoError(t, err)
	id, err := db.WriteValueWithAutokey(buffer.Bytes(), 1, 1, 3*time.Second)
	require.True(t, len(db.keyCache.keys()) == 1)
	require.True(t, id > 0)
	require.Eventually(t, func() bool {
		val, _, found, _ := db.GetValueByKey(id)
		return found == false && val == nil
	}, 6*time.Second, 500*time.Millisecond)
	require.NoError(t, err)

	require.True(t, len(db.keyCache.keys()) == 0)
}

func TestEvict(t *testing.T) {
	db, err := makeDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	tc := &testCompany{CompanyID: 99}
	buffer := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	err = enc.Encode(tc)
	require.NoError(t, err)
	id, err := db.WriteValueWithAutokey(buffer.Bytes(), 1, 1, 3*time.Second)

	require.True(t, len(db.keyCache.keys()) == 1)
	require.NoError(t, err)
	require.True(t, id > 0)
	// Eventually feels a bit odd here for the evict code.
	time.Sleep(5 * time.Second)
	err = db.Evict()
	require.NoError(t, err)
	// Since we want to get the raw value we need to cheat here.
	// If we went through the normal route then even if it existed in the database it would
	// check the TTL and return nil. Here we call the raw get command so it doesn't
	// check the item.TTL value.
	buf := make([]byte, 8)
	binary.PutUvarint(buf, id)
	res, _, err := db.db.Get(buf)
	require.Nil(t, res)
	require.True(t, errors.Is(err, pebbledb.ErrNotFound))
	require.True(t, len(db.keyCache.keys()) == 0)
}

func makeDB(t *testing.T) (*DB, error) {
	dir := t.TempDir()
	l := log.NewNopLogger()
	return NewDB(dir, l)
}

type testPerson struct {
	Name string
}

type testCompany struct {
	CompanyID int64
}
