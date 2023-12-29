package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "loki.test.logs",
		Args:    Arguments{},
		Exports: Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewComponent(opts, args.(Arguments))
		},
	})
}

type Component struct {
	mut         sync.Mutex
	index       int
	files       []string
	args        Arguments
	argsChan    chan Arguments
	writeTicker *time.Ticker
	churnTicker *time.Ticker
}

func NewComponent(o component.Options, c Arguments) (*Component, error) {
	return &Component{
		args:        c,
		index:       1,
		files:       make([]string, 0),
		writeTicker: time.NewTicker(c.WriteCadence),
		churnTicker: time.NewTicker(c.FileRefresh),
		argsChan:    make(chan Arguments),
	}, nil
}

func (c *Component) Run(ctx context.Context) error {
	defer c.writeTicker.Stop()
	defer c.churnTicker.Stop()
	// Create the initial set of files.
	c.churnFiles()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.writeTicker.C:
			c.writeFiles()
		case <-c.churnTicker.C:
			c.churnFiles()
		case args := <-c.argsChan:
			c.args = args
			c.writeTicker.Reset(c.args.WriteCadence)
			c.churnTicker.Reset(c.args.FileRefresh)
		}
	}
}

func (c *Component) Update(args component.Arguments) error {
	c.argsChan <- args.(Arguments)
	return nil
}

func (c *Component) writeFiles() {
	c.mut.Lock()
	defer c.mut.Unlock()

	for _, f := range c.files {
		attributes := make(map[string]string)
		attributes["ts"] = time.Now().Format(time.RFC3339)
		msgLen := 0
		if c.args.MessageMaxLength == c.args.MessageMinLength {
			msgLen = c.args.MessageMinLength
		} else {
			msgLen = rand.Intn(c.args.MessageMaxLength-c.args.MessageMinLength) + c.args.MessageMinLength

		}
		attributes["msg"] = gofakeit.Sentence(msgLen)
		for k, v := range c.args.Labels {
			attributes[k] = v
		}
		data, err := json.Marshal(attributes)
		if err != nil {
			continue
		}
		fh, err := os.OpenFile(f, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		_, _ = fh.WriteString(string(data) + "\n")
		fh.Close()
	}
}

func (c *Component) churnFiles() {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.args.NumberOfFiles > len(c.files) {
		fullpath := filepath.Join(c.args.Directory, strconv.Itoa(c.index)+".log")
		c.files = append(c.files, fullpath)
		_ = os.WriteFile(fullpath, []byte(""), 0644)
		c.index++
	} else if c.args.NumberOfFiles < len(c.files) {
		c.files = c.files[:c.args.NumberOfFiles]
	}

	churn := int(float64(c.args.NumberOfFiles) * c.args.FileChurnPercent)
	for i := 0; i < churn; i++ {
		candidate := rand.Intn(len(c.files))
		fullpath := filepath.Join(c.args.Directory, strconv.Itoa(c.index)+".log")
		c.files = append(c.files, fullpath)
		_ = os.WriteFile(fullpath, []byte(""), 0644)
		c.index++
		c.files[candidate] = fullpath
	}
}

type Arguments struct {
	Directory string `river:"directory,attribute"`
	// WriteCadance is the interval at which it will write to a file.
	WriteCadence     time.Duration     `river:"write_cadence,attribute,optional"`
	NumberOfFiles    int               `river:"number_of_files,attribute,optional"`
	Labels           map[string]string `river:"labels,attribute,optional"`
	MessageMaxLength int               `river:"message_max_length,attribute,optional"`
	MessageMinLength int               `river:"message_min_length,attribute,optional"`
	FileChurnPercent float64           `river:"file_churn_percent,attribute,optional"`
	// FileRefresh is the interval at which it will stop writing to a number of files equal to churn percent and start new ones.
	FileRefresh time.Duration `river:"file_refresh,attribute,optional"`
}

// SetToDefault implements river.Defaulter.
func (r *Arguments) SetToDefault() {
	*r = DefaultArguments()
}

func DefaultArguments() Arguments {
	return Arguments{
		WriteCadence:     1 * time.Second,
		NumberOfFiles:    1,
		MessageMaxLength: 10,
		MessageMinLength: 100,
		FileChurnPercent: 0.1,
		FileRefresh:      1 * time.Minute,
	}
}

// Validate implements river.Validator.
func (r *Arguments) Validate() error {
	if r.NumberOfFiles <= 0 {
		return fmt.Errorf("number_of_files must be greater than 0")
	}
	if r.MessageMaxLength < r.MessageMinLength {
		return fmt.Errorf("message_max_length must be greater than or equal to message_min_length")
	}
	if r.FileChurnPercent < 0 || r.FileChurnPercent > 1 {
		return fmt.Errorf("file_churn_percent must be between 0 and 1")
	}
	if r.WriteCadence < 0 {
		return fmt.Errorf("write_cadence must be greater than 0")
	}
	if r.FileRefresh < 0 {
		return fmt.Errorf("file_refresh must be greater than 0")
	}
	return nil
}

type Exports struct {
	Directory string `river:"directory,attribute"`
}
