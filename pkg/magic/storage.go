package magic

import (
	"context"

	"github.com/prometheus/prometheus/storage"
)

type Storage struct {
	// Embed Queryable/ChunkQueryable for compatibility, but don't actually implement it.
	storage.Queryable
	storage.ChunkQueryable

	a *Appender
}

func newStorage() *Storage {
	a := newAppender()
	return &Storage{
		a: a,
	}
}

func (s *Storage) Appender(ctx context.Context) storage.Appender {
	return s.a
}

func (s Storage) StartTime() (int64, error) {
	return 0, nil
}

func (s Storage) Close() error {
	return nil
}
