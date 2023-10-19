//go:build windows
// +build windows

// NOTE: This code is derived from Promtail but heavily modified.

package windowsevent

import (
	"github.com/grafana/loki/clients/pkg/promtail/targets/windows/win_eventlog"
	"os"
	"path/filepath"
)

const bookmarkKey = "bookmark"

type bookMark struct {
	handle win_eventlog.EvtHandle
	isNew  bool
	buf    []byte
}

type db struct {
	db   *KVDB
	path string
}

func newBookmarkDB(path string) (*db, error) {

	pdb, err := NewKVDB(path)
	if err != nil {
		// Let's try to recreate the file, it could be mangled.
		err = os.Remove(path)
		if err != nil {
			return nil, err
		}
		pdb, err = NewKVDB(path)
		if err != nil {
			return nil, err
		}
	}
	return &db{
		db:   pdb,
		path: path,
	}, nil
}

// newBookMark creates a new windows event bookmark.
// The bookmark will be saved at the given path. Use save to save the current position for a given event.
func (pdb *db) newBookMark() (*bookMark, error) {
	// 16kb buffer for rendering bookmark
	buf := make([]byte, 16<<10)
	var bookmark string
	pdb.transitionXML()
	valBytes, _, err := pdb.db.Get("bookmark", bookmarkKey)
	if err != nil {
		return nil, err
	}
	bookmark = string(valBytes)

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

func (pdb *db) transitionXML() {
	// See if we can convert the old bookmark to the new path.
	parentPath := filepath.Dir(pdb.path)
	bookmarkXML := filepath.Join(parentPath, "bookmark.xml")
	_, err := os.Stat(bookmarkXML)
	// Only continue if we can access the file.
	if err != nil {
		return
	}
	xmlBytes, err := os.ReadFile(bookmarkXML)
	if err != nil {
		// Try to remove the file so we dont do this again.
		_ = os.Remove(bookmarkXML)
		return
	}

	bookmark := string(xmlBytes)
	_ = pdb.db.Put("bookmark", bookmarkKey, []byte(bookmark))
	_ = os.Remove(bookmarkXML)

}

// save Saves the bookmark at the current event position.
func (pdb *db) save(b *bookMark, event win_eventlog.EvtHandle) error {
	newBookmark, err := win_eventlog.UpdateBookmark(b.handle, event, b.buf)
	if err != nil {
		return err
	}
	return pdb.db.Put("bookmark", bookmarkKey, []byte(newBookmark))
}

func (pdb *db) close() error {
	return pdb.db.Close()
}
