//go:build windows

package windowsevent

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// BookmarkFile represents reading and writing to a bookmark.xml file.
// These files are written sequentially in the format of bookmark.<number>.xml.
// Each individual bookmark file is immutable once written and is either read or deleted.
// The folder should ONLY contain bookmark files since all other files will be deleted.
type BookmarkFile struct {
	mut            sync.Mutex
	index          int
	directory      string
	oldDefaultFile string
	currentPath    string
}

var pathMatch, _ = regexp.Compile("bookmark.[0-9]+.xml")

// NewBookmarkFile creates a wrapper around saving the bookmark file.
func NewBookmarkFile(directory string, oldpath string) (*BookmarkFile, error) {
	_ = os.MkdirAll(filepath.Dir(directory), 0600)
	index, currentPath := findMostRecentAndPurge(directory, oldpath)
	return &BookmarkFile{
		directory:      directory,
		oldDefaultFile: oldpath,
		index:          index,
		currentPath:    currentPath,
	}, nil
}

// Put writes the value in to the newest file, and deletes the old one.
func (bf *BookmarkFile) Put(value string) error {
	bf.mut.Lock()
	defer bf.mut.Unlock()

	previousIndex := bf.index
	bf.index++
	newFile := fmt.Sprintf("bookmark.%d.xml", bf.index)
	newPath := filepath.Join(bf.directory, newFile)
	err := os.WriteFile(newPath, []byte(value), 0600)
	if err != nil {
		return err
	}
	writtenVal, err := os.ReadFile(newPath)
	if err != nil {
		_ = os.Remove(newPath)
		return err
	}
	// If for some reason the file was not written correctly then bail.
	if string(writtenVal) != value {
		_ = os.Remove(newPath)
		return fmt.Errorf("unable to save data , contents written differ from value")
	}
	// Finally we can delete the old file.
	oldFile := fmt.Sprintf("bookmark.%d.xml", previousIndex)
	// We don't care if it errors.
	_ = os.Remove(filepath.Join(bf.directory, oldFile))
	return nil
}

// findMostRecentAndPurge will find the most recent file, including the legacy bookmark.
// If found will return the index and the path to the newest file.
func findMostRecentAndPurge(dir string, legacyPath string) (int, string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 1, ""
	}
	index := 0
	var path string
	for _, f := range files {
		if pathMatch.MatchString(f.Name()) {
			stripped := strings.ReplaceAll(f.Name(), "bookmark.", "")
			number := strings.ReplaceAll(stripped, ".xml", "")
			foundNum, err := strconv.Atoi(number)
			if err != nil {
				continue
			}
			// Cant read the file.
			content, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				continue
			}

			// Need to ensure the file was properly saved previously.
			if xml.Unmarshal(content, new(interface{})) != nil {
				continue
			}
			if foundNum > index {
				index = foundNum
				path = f.Name()
			}
		}
	}
	// If we don't have a path then see if we can transition.
	if path == "" && legacyPath != "" {
		_, err = os.Stat(legacyPath)
		if err == nil {
			index = 1
			contents, _ := os.ReadFile(legacyPath)
			// Try to write the file if we have some contents.
			if len(contents) > 0 {
				newFile := fmt.Sprintf("bookmark.%d.xml", 1)
				newPath := filepath.Join(dir, newFile)
				_ = os.WriteFile(newPath, contents, 0600)
				_ = os.Remove(legacyPath)
			}
		}
	}
	if index == 0 {
		index = 1
	}

	// Finally delete all files other than the found.
	for _, f := range files {
		if f.Name() != path {
			_ = os.Remove(filepath.Join(dir, f.Name()))
		}
	}
	return index, path
}

// Get returns the value which is "" if nothing found.
// If the bucket does not exist then it will be created.
func (bf *BookmarkFile) Get() string {
	val, _ := os.ReadFile(filepath.Join(bf.directory, bf.currentPath))
	// We don't want to propagate the path error up the stack if its does not exist.
	return string(val)
}
