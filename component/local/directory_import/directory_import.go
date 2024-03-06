package directoryimport

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/internal/featuregate"
	"github.com/grafana/agent/pkg/flow/logging/level"
	httpsvc "github.com/grafana/agent/service/http"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
	"github.com/grafana/river/printer"
	"github.com/grafana/river/rivertypes"
	"github.com/grafana/river/scanner"
	"github.com/grafana/river/token"
)

func init() {
	component.Register(component.Registration{
		Name:      "local.directory_import",
		Stability: featuregate.StabilityExperimental,

		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the local.file component.
type Arguments struct {
	// Filename indicates the file to watch.
	Path string `river:"path,attr"`
	// PollFrequency determines the frequency to check for changes when Type is
	// Poll.
	PollFrequency time.Duration `river:"poll_frequency,attr,optional"`
}

// DefaultArguments provides the default arguments for this component.
var DefaultArguments = Arguments{
	PollFrequency: 15 * time.Second,
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

// Exports holds values which are exported by this component.
type Exports struct {
	// Generated importable module
	Content rivertypes.OptionalSecret `river:"content,attr"`
}

// Component implements the local.file component.
type Component struct {
	opts component.Options

	mut           sync.Mutex
	args          Arguments
	latestContent string
	lastPoll      time.Time

	healthMut sync.RWMutex
	health    component.Health

	url string

	// reloadCh is a buffered channel which is written to when the watched file
	// should be reloaded by the component.
	reloadCh chan struct{}
}

// New creates a new component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{
		opts: o,

		reloadCh: make(chan struct{}, 1),
	}

	c.opts.OnStateChange(Exports{
		Content: rivertypes.OptionalSecret{
			IsSecret: false,
			Value:    initial_config,
		},
	})

	if data, err := c.opts.GetServiceData(httpsvc.ServiceName); err == nil {
		if hdata, ok := data.(httpsvc.Data); ok {
			c.url = fmt.Sprintf("%s%s/raw", hdata.HTTPListenAddr, hdata.HTTPPathForComponent(c.opts.ID))
		}
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.nextPoll()):
			c.poll()
		case <-c.reloadCh:
			c.mut.Lock()
			c.lastPoll = time.Time{}
			c.mut.Unlock()
		}
	}
}

func (c *Component) poll() {
	c.mut.Lock()
	defer func() {
		c.lastPoll = time.Now()
		c.mut.Unlock()
	}()
	data, err := c.scanDirectory()
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "scanning directory", "err", err)
		return
	}
	content := c.buildDynamicModule(data)
	if content != c.latestContent {
		c.latestContent = content
		c.opts.OnStateChange(Exports{
			Content: rivertypes.OptionalSecret{
				IsSecret: false,
				Value:    content,
			},
		})
	}
}

func (c *Component) buildDynamicModule(children map[string][]byte) string {
	bs := &ast.BlockStmt{
		Name:  []string{"declare"},
		Label: "all",
	}
	failCount := 0
	for name, content := range children {
		inner, err := c.innerModuleContent(content)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "invalid dynamic module", "name", name, "err", err)
			failCount++
			continue
		}
		modName, err := sanitizeName(name)
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "sanitizing name", "name", name, "err", err)
			failCount++
			continue
		}
		// build text of inner module
		importStringStmt := &ast.BlockStmt{
			Name:  []string{"import", "string"},
			Label: modName,
			Body: ast.Body{
				&ast.AttributeStmt{
					Name: &ast.Ident{Name: "content"},
					Value: &ast.LiteralExpr{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", inner),
					},
				},
			},
		}
		bs.Body = append(bs.Body, importStringStmt)
		// now generate a usage of main:
		usageStmt := &ast.BlockStmt{
			Name:  []string{modName, "main"},
			Label: "main",
		}
		bs.Body = append(bs.Body, usageStmt)
	}
	bs.Body = append(bs.Body, &ast.BlockStmt{
		Name:  []string{"export"},
		Label: "failedModules",
		Body: ast.Body{
			&ast.AttributeStmt{
				Name: &ast.Ident{Name: "value"},
				Value: &ast.LiteralExpr{
					Kind:  token.NUMBER,
					Value: fmt.Sprint(failCount),
				},
			},
		},
	})
	buf := &bytes.Buffer{}
	printer.Fprint(buf, bs)
	return buf.String()
}

func (c *Component) innerModuleContent(f []byte) (string, error) {
	bs := &ast.BlockStmt{
		Name:  []string{"declare"},
		Label: "main",
	}
	file, err := parser.ParseFile("", f)
	if err != nil {
		return "", err
	}
	// todo: validate body way more
	// todo: extract args to parent
	bs.Body = file.Body
	buf := &bytes.Buffer{}
	printer.Fprint(buf, bs)
	return buf.String(), nil
}

func sanitizeName(s string) (string, error) {
	s = strings.TrimSuffix(s, ".river")
	return scanner.SanitizeIdentifier(s)
}

// scan directory for .river files and put their content in a map
func (c *Component) scanDirectory() (map[string][]byte, error) {
	files := map[string][]byte{}
	fs, err := os.ReadDir(c.args.Path)
	if err != nil {
		return nil, err
	}
	for _, info := range fs {
		if !info.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(info.Name(), ".river") {
			continue
		}
		dat, err := os.ReadFile(filepath.Join(c.args.Path, info.Name()))
		if err != nil {
			level.Error(c.opts.Logger).Log("msg", "reading file", "err", err, "file", info.Name())
			continue
		}
		name := strings.TrimSuffix(info.Name(), ".river")
		files[name] = dat
	}
	return files, nil
}

func (c *Component) nextPoll() time.Duration {
	c.mut.Lock()
	defer c.mut.Unlock()

	nextPoll := c.lastPoll.Add(c.args.PollFrequency)
	now := time.Now()

	if now.After(nextPoll) {
		// Poll immediately; next poll period was in the past.
		return 0
	}
	return nextPoll.Sub(now)
}

// temporary placeholder config so we always start up correctly.
const initial_config = `declare "all" {
	export "const" {
	  value = 44  
	}
  }`

func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	if newArgs.PollFrequency <= 0 {
		return fmt.Errorf("poll_frequency must be greater than 0")
	}

	c.mut.Lock()
	defer c.mut.Unlock()
	// if path is changed, clear the config so it starts valid
	if newArgs.Path != c.args.Path {
		c.opts.OnStateChange(Exports{
			Content: rivertypes.OptionalSecret{
				IsSecret: false,
				Value:    initial_config,
			},
		})
	}
	c.args = newArgs

	select {
	case c.reloadCh <- struct{}{}:
	default:
	}
	return nil
}

// DebugInfo returns debug information for this component.
func (c *Component) DebugInfo() interface{} {
	return struct {
		RawConfigURL string `river:"config_url,attr"`
	}{c.url}
}

func (c *Component) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// very simple path handling
		// only responds to `/raw
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) != 1 || parts[0] != "raw" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, err := w.Write([]byte(c.latestContent))
		if err != nil {
			w.WriteHeader(500)
		}
	})
}
