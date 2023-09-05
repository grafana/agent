package pebble

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/go-kit/log"
	"io"
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
	id, err := db.WriteValueWithAutokey(tc, 0)
	require.NoError(t, err)

	tc1, found, err := db.GetValueByKey(id)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, tc1.(*testCompany).CompanyID == tc.CompanyID)

	tp := &testPerson{Name: "Bob Dill"}
	id, err = db.WriteValueWithAutokey(tp, 0)
	require.NoError(t, err)

	tp1, found, err := db.GetValueByKey(id)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, tp1.(*testPerson).Name == tp.Name)

	require.True(t, len(db.keyCache.keys()) == 2)
}

func TestTTL(t *testing.T) {
	db, err := makeDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	tc := &testCompany{CompanyID: 99}
	id, err := db.WriteValueWithAutokey(tc, 3*time.Second)
	require.True(t, len(db.keyCache.keys()) == 1)
	require.True(t, id > 0)
	require.Eventually(t, func() bool {
		val, found, _ := db.GetValueByKey(id)
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
	id, err := db.WriteValueWithAutokey(tc, 3*time.Second)

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
	return NewDB(dir, getValue, getType, l)
}

func getValue(data []byte, t int8) (any, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	switch t {
	case 0:
		tp := &testPerson{}
		err := dec.Decode(tp)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return tp, err
	case 1:
		tc := &testCompany{}
		err := dec.Decode(tc)
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return tc, err
	default:
		return nil, fmt.Errorf("Unknown type found %d", t)
	}
}

func getType(data any) (int8, int, error) {
	switch t := data.(type) {
	case *testPerson:
		return 0, 0, nil
	case *testCompany:
		return 1, 1, nil
	default:
		return 0, 0, fmt.Errorf("unknown type %t", t)
	}
}

type testPerson struct {
	Name string
}

type testCompany struct {
	CompanyID int64
}
