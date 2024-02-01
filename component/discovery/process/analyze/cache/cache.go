package cache

import (
	"debug/elf"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/discovery/process/analyze"
)

type Cache struct {
	l        log.Logger
	pids     map[uint32]*Entry
	stats    map[Stat]*Entry
	buildIDs map[string]*analyze.Results
}

func New(logger log.Logger) *Cache {
	return &Cache{
		l:        logger,
		pids:     make(map[uint32]*Entry),
		stats:    make(map[Stat]*Entry),
		buildIDs: make(map[string]*analyze.Results),
	}
}

type Entry struct {
	Results *analyze.Results
	Stat    Stat
	BuildID string
}

func (c *Cache) GetPID(pid uint32) *Entry {
	return c.pids[pid]
}

func (c *Cache) Put(pid uint32, a *Entry) {
	c.pids[pid] = a
	if a.Stat.Inode != 0 && a.Stat.Dev != 0 {
		c.stats[a.Stat] = a
	}
	if a.BuildID != "" {
		c.buildIDs[a.BuildID] = a.Results
	}
}

func (c *Cache) GetStat(s Stat) *Entry {
	return c.stats[s]
}

func (c *Cache) GetBuildID(buildID string) *analyze.Results {
	if buildID == "" {
		return nil
	}
	return c.buildIDs[buildID]
}
func (c *Cache) AnalyzePID(pid string) (*analyze.Results, error) {
	ipid, _ := strconv.Atoi(pid)
	exePath := filepath.Join("/proc", pid, "exe")
	return c.AnalyzePIDPath(uint32(ipid), pid, exePath)
}
func (c *Cache) AnalyzePIDPath(pid uint32, pidS string, exePath string) (*analyze.Results, error) {

	e := c.GetPID(pid)
	if e != nil {
		return e.Results, nil
	}

	// check if executable exists
	fi, err := os.Stat(exePath)
	if err != nil {
		return nil, err
	}
	st := StatFromFileInfo(fi)
	e = c.GetStat(st)
	if e != nil {
		c.Put(pid, e)
		return e.Results, nil
	}

	// get path to executable
	f, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ef, err := elf.NewFile(f)
	if err != nil {
		return nil, err
	}
	defer ef.Close()

	buildID, _ := BuildID(ef)
	r := c.GetBuildID(buildID)
	if r != nil {
		c.Put(pid, &Entry{
			Results: r,
			Stat:    st,
			BuildID: buildID,
		})
		return r, nil
	}

	r = analyze.Analyze(c.l, analyze.Input{
		PID:     pid,
		PIDs:    pidS,
		File:    f,
		ElfFile: ef,
	})

	c.Put(pid, &Entry{
		Results: r,
		Stat:    st,
		BuildID: buildID,
	})
	return r, nil
}

func (c *Cache) GC(active map[uint32]struct{}) {
	for pid := range c.pids {
		if _, ok := active[pid]; !ok {
			delete(c.pids, pid)
		}
	}
	reachableStats := make(map[Stat]struct{})
	reachableBuildIDs := make(map[string]struct{})
	for _, e := range c.pids {
		reachableStats[e.Stat] = struct{}{}
		if e.BuildID != "" {
			reachableBuildIDs[e.BuildID] = struct{}{}
		}
	}
	for s := range c.stats {
		if _, ok := reachableStats[s]; !ok {
			delete(c.stats, s)
		}
	}
	for id := range c.buildIDs {
		if _, ok := reachableBuildIDs[id]; !ok {
			delete(c.buildIDs, id)
		}
	}
}
