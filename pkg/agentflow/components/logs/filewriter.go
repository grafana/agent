package logs

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"os"
	"strings"
)

type FileWriter struct {
	cfg  config.LogFileWriter
	self *actor.PID
	name string
	file *os.File
}

func (m *FileWriter) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.Metrics}
}

func (m *FileWriter) Output() actorstate.InOutType {
	return actorstate.Metrics
}

func NewFileWriter(name string, cfg config.LogFileWriter) (actorstate.FlowActor, error) {
	return &FileWriter{
		cfg:  cfg,
		name: name,
	}, nil
}

func (m *FileWriter) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Start:
		m.self = c.Self()
		m.file, _ = os.OpenFile(m.cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	case []exchange.Log:
		sb := strings.Builder{}
		//TODO: This actually should NOT use original, but instead decode into logfmt
		for _, l := range msg {
			sb.Write(l.Original())
		}
		m.file.Write([]byte(sb.String()))
	}
}

func (m *FileWriter) Name() string {
	return m.name
}

func (m *FileWriter) PID() *actor.PID {
	return m.self
}
