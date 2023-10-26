//go:build windows
// +build windows

// NOTE: This code is derived from Promtail but heavily modified.

package windowsevent

import (
	"github.com/grafana/loki/clients/pkg/promtail/targets/windows/win_eventlog"
)

type bookMark struct {
	handle win_eventlog.EvtHandle
	isNew  bool
	buf    []byte
}

type db struct {
	file *BookmarkFile
	path string
}

func newBookmarkDB(path string, legacyPath string) (*db, error) {
	pdb, err := NewBookmarkFile(path, legacyPath)
	if err != nil {
		return nil, err
	}
	return &db{
		file: pdb,
		path: path,
	}, nil
}

// newBookMark creates a new windows event bookmark.
// The bookmark will be saved at the given path. Use save to save the current position for a given event.
func (pdb *db) newBookMark() (*bookMark, error) {
	// 16kb buffer for rendering bookmark
	buf := make([]byte, 16<<10)
	bookmark := pdb.file.Get()
	// creates a new bookmark file if none exists.
	if bookmark == "" {
		bm, err := win_eventlog.CreateBookmark("")
		if err != nil {
			return nil, err
		}
		return &bookMark{
			handle: bm,
			isNew:  true,
			buf:    buf,
		}, nil
	}

	// load the current bookmark.
	bm, err := win_eventlog.CreateBookmark(bookmark)
	if err != nil {
		return nil, err
	}
	return &bookMark{
		handle: bm,
		isNew:  bookmark == "",
		buf:    buf,
	}, nil
}

// save Saves the bookmark at the current event position.
func (pdb *db) save(b *bookMark, event win_eventlog.EvtHandle) error {
	newBookmark, err := win_eventlog.UpdateBookmark(b.handle, event, b.buf)
	if err != nil {
		return err
	}
	return pdb.file.Put(newBookmark)
}
