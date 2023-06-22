package badger

import (
	badgerdb "github.com/dgraph-io/badger/v3"
	"github.com/grafana/agent/pkg/flow/logging"
)

type bookmark struct {
	d   *badgerdb.DB
	log *logging.Logger
}

func newBookmark(dir string, l *logging.Logger) (*bookmark, error) {
	bdb, err := badgerdb.Open(badgerdb.DefaultOptions(dir))
	if err != nil {
		return nil, err
	}

	newDb := &bookmark{
		d:   bdb,
		log: l,
	}
	if err != nil {
		return nil, err
	}
	return newDb, nil
}

func (d *bookmark) getValueForKey(k string) ([]byte, bool, error) {
	var value []byte
	var found bool
	err := d.d.View(func(txn *badgerdb.Txn) error {
		item, err := txn.Get([]byte(k))
		if err == badgerdb.ErrKeyNotFound {
			found = false
			return nil
		}
		found = true
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, found, err
}

func (d *bookmark) writeBookmark(k string, v []byte) error {
	if len(v) == 0 {
		return nil
	}

	err := d.d.Update(func(txn *badgerdb.Txn) error {
		inErr := txn.SetEntry(&badgerdb.Entry{
			Key:   []byte(k),
			Value: v,
		})
		return inErr
	})
	if err != nil {
		return err
	}
	return nil
}
