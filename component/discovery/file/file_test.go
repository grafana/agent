package file

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grafana/agent/component/discovery"

	"golang.org/x/net/context"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	var mut sync.Mutex
	dir := path.Join(os.TempDir(), "agent_testing", "t1")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	foundFiles := make([]discovery.Target, 0)
	c := createComponent(t, dir, func(e component.Exports) {
		mut.Lock()
		foundFiles = e.(Exports).Targets
		mut.Unlock()
	}, []string{path.Join(dir, "*.txt")}, nil)
	ct := context.Background()
	ct, _ = context.WithTimeout(ct, 5*time.Second)
	c.args.UpdatePeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	ct.Done()
	mut.Lock()
	require.Len(t, foundFiles, 1)
	require.True(t, contains(foundFiles, "t1.txt"))
	mut.Unlock()
}

func TestAddingFile(t *testing.T) {
	var mut sync.Mutex

	dir := path.Join(os.TempDir(), "agent_testing", "t2")
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	foundFiles := make([]discovery.Target, 0)
	c := createComponent(t, dir, func(e component.Exports) {
		mut.Lock()
		foundFiles = e.(Exports).Targets
		mut.Unlock()
	}, []string{path.Join(dir, "*.txt")}, nil)

	ct := context.Background()
	ct, _ = context.WithTimeout(ct, 40*time.Second)
	c.args.UpdatePeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	ct.Done()
	mut.Lock()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	mut.Unlock()
}

func TestAddingFileInSubDir(t *testing.T) {
	var mut sync.Mutex

	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	foundFiles := make([]discovery.Target, 0)
	c := createComponent(t, dir, func(e component.Exports) {
		mut.Lock()
		foundFiles = e.(Exports).Targets
		mut.Unlock()
	}, []string{path.Join(dir, "**", "*.txt")}, nil)

	ct := context.Background()
	ct, _ = context.WithTimeout(ct, 40*time.Second)
	c.args.UpdatePeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	time.Sleep(20 * time.Millisecond)
	err := ioutil.WriteFile(path.Join(subdir, "t3.txt"), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
	ct.Done()
	mut.Lock()
	require.Len(t, foundFiles, 3)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))
	mut.Unlock()
}

func TestAddingRemovingFileInSubDir(t *testing.T) {
	var mut sync.Mutex

	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	foundFiles := make([]discovery.Target, 0)
	c := createComponent(t, dir, func(e component.Exports) {
		mut.Lock()
		foundFiles = e.(Exports).Targets
		mut.Unlock()
	}, []string{path.Join(dir, "**", "*.txt")}, nil)

	ct := context.Background()
	ct, _ = context.WithTimeout(ct, 40*time.Second)
	c.args.UpdatePeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(20 * time.Millisecond)
	writeFile(t, dir, "t2.txt")
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	time.Sleep(100 * time.Millisecond)
	err := ioutil.WriteFile(path.Join(subdir, "t3.txt"), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	mut.Lock()
	require.Len(t, foundFiles, 3)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))
	mut.Unlock()

	err = os.RemoveAll(subdir)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	mut.Lock()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t2.txt"))
	mut.Unlock()

}

func TestExclude(t *testing.T) {
	var mut sync.Mutex

	dir := path.Join(os.TempDir(), "agent_testing", "t3")
	os.MkdirAll(dir, 0755)
	writeFile(t, dir, "t1.txt")
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	foundFiles := make([]discovery.Target, 0)
	c := createComponent(t, dir, func(e component.Exports) {
		mut.Lock()
		foundFiles = e.(Exports).Targets
		mut.Unlock()
	}, []string{path.Join(dir, "**/*.txt")}, []string{path.Join(dir, "**/*.bad")})
	ct := context.Background()
	ct, _ = context.WithTimeout(ct, 40*time.Second)
	c.args.UpdatePeriod = 10 * time.Millisecond
	go c.Run(ct)
	time.Sleep(100 * time.Millisecond)
	subdir := path.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	writeFile(t, subdir, "t3.txt")
	time.Sleep(100 * time.Millisecond)
	mut.Lock()
	require.Len(t, foundFiles, 2)
	require.True(t, contains(foundFiles, "t1.txt"))
	require.True(t, contains(foundFiles, "t3.txt"))
	mut.Unlock()
}

func createComponent(t *testing.T, dir string, foundFiles func(e component.Exports), paths []string, excluded []string) *Component {
	l := util.TestLogger(t)
	c, err := New(component.Options{
		ID:             "test",
		Logger:         l,
		DataPath:       dir,
		OnStateChange:  foundFiles,
		Registerer:     prometheus.DefaultRegisterer,
		Tracer:         nil,
		HTTPListenAddr: "",
		HTTPPath:       "",
	}, Arguments{
		Paths:         paths,
		ExcludedPaths: excluded,
		UpdatePeriod:  1 * time.Second,
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
	err := ioutil.WriteFile(path.Join(dir, name), []byte("asdf"), 0664)
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
}
