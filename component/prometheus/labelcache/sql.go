package labelcache

import (
	"arena"
	"context"
	"database/sql"
	"fmt"
	"github.com/prometheus/prometheus/model/labels"
	"os"
	"path"
	"sync"
	"time"
)
import _ "modernc.org/sqlite"

type sqlcache struct {
	mut sync.Mutex
	db *sql.DB
}

func newSQLCache(directory string) (*sqlcache, error) {
	fullpath := path.Join(directory, "cache.db")
	_ = os.Remove(fullpath)
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s", fullpath))
	if err != nil {
		return nil, err
	}
	return &sqlcache{db: db}, nil
}

func createTables(c *sqlcache) {
	_, _ = c.db.Exec(`
		CREATE TABLE LABEL_TO_ID (
			LABEL TEXT PRIMARY KEY,
			ID INTEGER,
		)`)

	_, _ = c.db.Exec(`
		CREATE TABLE ID_TO_LABEL (
			ID INTEGER PRIMARY KEY,
			LABEL TEXT,
		)`)
}

func (s *sqlcache) WriteLabels(lbls [][]labels.Label, ttl time.Duration, mem *arena.Arena) ([]uint64, error) {
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	returnKeys := arena.MakeSlice[uint64](mem,len(lbls),len(lbls))
	for _,l := range lbls {
		lblStr := (labels.Labels)(l).String()
		res, err := tx.Query("INSERT OR IGNORE INTO LABEL_TO_ID SELECT ?; SELECT ID FROM LABEL_TO_ID WHERE LABEL = ?",lblStr,lblStr)
		if err != nil {
			return nil,err
		}
		var k int64
		res.Scan(&k)
	}
	 tx.Exec("INSERT INTO LABEL_TO_ID SELECT LABEL FROM LABEL_MATCH WHERE LABEL NOT IN")

	if err != nil {
		return nil, err
	}
	for
}

func (s sqlcache) GetLabels(keys []uint64, mem *arena.Arena) ([]labels.Labels, error) {
	//TODO implement me
	panic("implement me")
}

func (s sqlcache) GetOrAddLink(componentID string, localRefID uint64, lbls labels.Labels) uint64 {
	//TODO implement me
	panic("implement me")
}

func (s sqlcache) GetOrAddGlobalRefID(l labels.Labels) uint64 {
	//TODO implement me
	panic("implement me")
}

func (s sqlcache) GetGlobalRefID(componentID string, localRefID uint64) uint64 {
	//TODO implement me
	panic("implement me")
}

func (s sqlcache) GetLocalRefID(componentID string, globalRefID uint64) uint64 {
	//TODO implement me
	panic("implement me")
}
