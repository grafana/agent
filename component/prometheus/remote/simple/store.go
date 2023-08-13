package simple

type BookmarkStore interface {
	WriteBookmark(wal string, forwardTo string, value *Bookmark) error
	GetBookmark(wal string, forwardTo string) (*Bookmark, bool)
}

type Bookmark struct {
	Key uint64
}
