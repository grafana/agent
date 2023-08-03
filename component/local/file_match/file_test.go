//go:build !windows

// This should run on windows but windows does not like the tight timing of file creation and deletion.
package file_match

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"

	"context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t1")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "*.txt")}, nil)
	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 5*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	ct.Done()
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 1)
	require.True(t, contains(foundFiles, "t1.txt"))
}

func TestDirectoryFile(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t1")
	subdir := path.Join(dir, "subdir")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)
	writeFile(t, subdir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "**/")}, nil)
	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 5*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	ct.Done()
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 1)
	require.True(t, contains(foundFiles, "t1.txt"))
}

func TestAddingFile(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t2")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "*.txt")}, nil)

	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 40*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	ct.Done()
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
}

func TestAddingFileInSubDir(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "**", "*.txt")}, nil)
	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 40*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	time.Sleep(20 * time.Millisecond)
	err := os.WriteFile(path.Join(subdir, "t3.txt"), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
	ct.Done()
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 3)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))
}

func TestAddingRemovingFileInSubDir(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "**", "*.txt")}, nil)

	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 40*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	time.Sleep(100 * time.Millisecond)
	err := os.WriteFile(path.Join(subdir, "t3.txt"), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 3)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))

	err = os.RemoveAll(subdir)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	foundFiles = c.getWatchedFiles()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
}

func TestExclude(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponent(t, dir, []string{path.Join(dir, "**", "*.txt")}, []string{path.Join(dir, "**", "*.bad")})
	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 40*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(100 * time.Millisecond)
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	writeFile(t, subdir, "t3.txt")
	time.Sleep(100 * time.Millisecond)
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))
}

func TestMultiLabels(t *testing.T) {
	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	c := createComponentWithLabels(t, dir, []string{path.Join(dir, "**", "*.txt"), path.Join(dir, "**", "*.txt")}, nil, map[string]string{
		"foo":   "bar",
		"fruit": "apple",
	})
	c.args.PathTargets[0]["newlabel"] = "test"
	ct := context.Background()
	ct, ccl := context.WithTimeout(ct, 40*time.Second)
	defer ccl()
	c.args.SyncPeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(100 * time.Millisecond)
	foundFiles := c.getWatchedFiles()
	require.Len(t, foundFiles, 2)
	require.True(t, contains([]discovery.Target{foundFiles[0]}, "t1.txt"))
	require.True(t, contains([]discovery.Target{foundFiles[1]}, "t1.txt"))
}

func createComponent(t *testing.T, dir string, paths []string, excluded []string) *Component {
	return createComponentWithLabels(t, dir, paths, excluded, nil)
}

func createComponentWithLabels(t *testing.T, dir string, paths []string, excluded []string, labels map[string]string) *Component {
	tPaths := make([]discovery.Target, 0)
	for _, p := range paths {
		tar := discovery.Target{"__path__": p}
		for k, v := range labels {
			tar[k] = v
		}
		tPaths = append(tPaths, tar)
	}
	for _, p := range excluded {
		tar := discovery.Target{"__path_exclude__": p}
		for k, v := range labels {
			tar[k] = v
		}
		tPaths = append(tPaths, tar)
	}
	c, err := New(component.Options{
		ID:       "test",
		Logger:   util.TestFlowLogger(t),
		DataPath: dir,
		OnStateChange: func(e component.Exports) {

		},
		Registerer: prometheus.DefaultRegisterer,
		Tracer:     nil,
	}, Arguments{
		PathTargets: tPaths,
		SyncPeriod:  1 * time.Second,
	})

	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

func contains(sources []discovery.Target, match string) bool {
	for _, s := range sources {
		p := s["__path__"]
		if strings.Contains(p, match) {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, dir string, name string) {
	err := os.WriteFile(path.Join(dir, name), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
}
