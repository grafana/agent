package remotewrites

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
)

type FakeMetric struct {
	name string
	self *actor.PID
}

func (f *FakeMetric) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.Metrics}
}

func (f *FakeMetric) Output() actorstate.InOutType {
	return actorstate.None
}

func NewFakeMetricRemoteWrite(name string) (actorstate.FlowActor, error) {
	return &FakeMetric{
		name: name,
	}, nil
}

func (f *FakeMetric) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Start:
		f.self = c.Self()
	case []exchange.Metric:
		fmt.Printf("recieved %d metrics \n", len(msg))
	}
}

func (f *FakeMetric) Name() string {
	return f.name
}

func (f FakeMetric) PID() *actor.PID {
	return f.PID()
}
