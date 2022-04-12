package logs

import (
	"bytes"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/scheduler"
	"github.com/go-logfmt/logfmt"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"os"
	"sync"
	"time"
)

type Agent struct {
	self *actor.PID
	outs []*actor.PID
	root *actor.RootContext
	name string

	logsBuffer [][]byte
	logsMutex  sync.Mutex

	cancel scheduler.CancelFunc
}

func NewAgent(name string, root *actor.RootContext) (actorstate.FlowActor, error) {
	return &Agent{
		name:       name,
		root:       root,
		logsBuffer: make([][]byte, 0),
	}, nil
}

func (a *Agent) PID() *actor.PID {
	return a.self
}

func (a *Agent) Write(p []byte) (n int, err error) {
	n, err = os.Stdout.Write(p)
	a.logsMutex.Lock()
	a.logsBuffer = append(a.logsBuffer, p)
	defer a.logsMutex.Unlock()
	return n, err
}

func (a *Agent) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.None}
}

func (a *Agent) Output() actorstate.InOutType {
	return actorstate.Logs
}

func (a *Agent) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		a.outs = msg.Children
	case actorstate.Start:
		a.self = c.Self()
		sched := scheduler.NewTimerScheduler(c)
		a.cancel = sched.SendRepeatedly(1*time.Millisecond, 1*time.Second, c.Self(), "flush")
	case string:
		if msg != "flush" {
			return
		}
		a.logsMutex.Lock()
		if len(a.logsBuffer) == 0 {
			a.logsMutex.Unlock()
			return
		}
		cpy := make([][]byte, len(a.logsBuffer))
		copy(cpy, a.logsBuffer)
		a.logsBuffer = a.logsBuffer[:0]
		a.logsMutex.Unlock()
		logs := make([]exchange.Log, 0)
		for _, b := range cpy {
			bb := bytes.Buffer{}
			bb.Write(b)
			d := logfmt.NewDecoder(&bb)
			labels := make(map[string]string)
			for d.ScanRecord() {
				for d.ScanKeyval() {
					labels[string(d.Key())] = string(d.Value())
				}
			}
			//TODO: Time.now is incorrect but don't feel like parsing the time
			l := exchange.NewLog(time.Now(), labels, b)
			logs = append(logs, l)
		}
		for _, o := range a.outs {
			a.root.Send(o, logs)
		}
	}
}

func (m *Agent) Name() string {
	return m.name
}
