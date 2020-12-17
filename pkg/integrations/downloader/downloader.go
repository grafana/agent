// Package downloader implements a file downloader that can download files from
// a URL and verify checksums of the file.
package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Global is a globally shared downloader.
var Global = New(nil)

type Downloader struct {
	client *http.Client

	// locks holds a lock for downloading a specific file.
	lockMtx sync.Mutex
	locks   map[string]*sync.Mutex
}

// New creates a new Downloader. If c is nil, http.DefaultClient will be used.
func New(c *http.Client) *Downloader {
	if c == nil {
		c = http.DefaultClient
	}
	return &Downloader{
		client: c,
		locks:  make(map[string]*sync.Mutex),
	}
}

// Download downloads a file from a given URL to a path on disk. The checksum
// is used to validate that the file has the expected contents. If the file
// already exists on disk and has the expected checksum, the existing file will
// be kept and the download will be skipped.
//
// Download is safe to call concurrently with the same target file path.
func (d *Downloader) Download(ctx context.Context, url, file, checksum string) error {
	d.lock(file)
	defer d.unlock(file)

	if exist, err := d.exists(ctx, file, checksum); exist {
		return nil
	} else if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer f.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not form HTTP request: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status code from %s: %s", url, resp.Status)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Seek back to beginning for checksum validation.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to start of downloaded file %s: %w", file, err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to validate checksum: %w", err)
	}
	if downloadedHash := hex.EncodeToString(h.Sum(nil)); downloadedHash != checksum {
		return fmt.Errorf("checksums do not match: expected %s, got %s", checksum, downloadedHash)
	}

	return nil
}

// Exists returns true if the target file exists and has the expected checksum.
func (d *Downloader) Exists(ctx context.Context, file, checksum string) (bool, error) {
	d.lock(file)
	defer d.unlock(file)

	return d.exists(ctx, file, checksum)
}

// exists is Exists but doesn't grab a lock on the file.
func (d *Downloader) exists(ctx context.Context, file, checksum string) (bool, error) {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return hex.EncodeToString(h.Sum(nil)) == checksum, nil
}

func (d *Downloader) lock(file string) {
	d.lockMtx.Lock()
	defer d.lockMtx.Unlock()

	mtx := d.locks[file]
	if mtx == nil {
		mtx = &sync.Mutex{}
		d.locks[file] = mtx
	}
	mtx.Lock()
}

func (d *Downloader) unlock(file string) {
	d.lockMtx.Lock()
	defer d.lockMtx.Unlock()

	mtx := d.locks[file]
	if mtx == nil {
		panic(file + " not locked")
	}
	mtx.Unlock()
}
