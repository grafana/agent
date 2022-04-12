package actorsystem

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/agentflow/components/auth"
	"github.com/grafana/agent/pkg/agentflow/components/integrations"
	"github.com/grafana/agent/pkg/agentflow/components/logs"
	"github.com/grafana/agent/pkg/agentflow/components/metrics"
	"github.com/grafana/agent/pkg/agentflow/components/remotewrites"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"strings"
	"time"
)

type Orchestrator struct {
	cfg config.Config

	actorSystem      *actor.ActorSystem
	rootContext      *actor.RootContext
	nameToPID        map[string]*actor.PID
	pidToName        map[*actor.PID]string
	nameToActor      map[string]actorstate.FlowActor
	parentToChildren map[*actor.PID][]*actor.PID
}

func NewOrchestrator(cfg config.Config) *Orchestrator {
	return &Orchestrator{
		cfg:              cfg,
		nameToPID:        map[string]*actor.PID{},
		pidToName:        map[*actor.PID]string{},
		nameToActor:      map[string]actorstate.FlowActor{},
		parentToChildren: map[*actor.PID][]*actor.PID{},
	}
}

func (u *Orchestrator) StartActorSystem(as *actor.ActorSystem, root *actor.RootContext) error {
	u.actorSystem = as
	u.rootContext = root
	var agentLog *logs.Agent
	// Find if they have defined the agent logger
	for _, nodeCfg := range u.cfg.Nodes {
		if nodeCfg.AgentLogs != nil {
			no, err := logs.NewAgent(nodeCfg.Name, root)
			if err != nil {
				return err
			}
			agentLog = no.(*logs.Agent)
			u.addPID(no)
			break
		}
	}
	// If they have not defined one, then create an internal one
	if agentLog == nil {
		no, err := logs.NewAgent("__agent_log", root)
		if err != nil {
			return err
		}
		agentLog = no.(*logs.Agent)
	}
	logger := log.NewLogfmtLogger(agentLog)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	global := &types.Global{
		Log: logger,
	}
	// Generate the Nodes
	for _, nodeCfg := range u.cfg.Nodes {
		err := u.processNode(nodeCfg, global)
		if err != nil {
			return err
		}
	}
	// Assign all the outputs
	for _, nodeCfg := range u.cfg.Nodes {
		outs := make([]*actor.PID, 0)
		for _, out := range nodeCfg.Outputs {
			pid, found := u.nameToPID[out]
			if !found {
				return fmt.Errorf("unable to find output %s on node named %s", out, nodeCfg.Name)
			}
			outs = append(outs, pid)
		}
		n, found := u.nameToPID[nodeCfg.Name]
		if !found {
			return fmt.Errorf("unable to find %s in name to pid", nodeCfg.Name)
		}
		u.parentToChildren[n] = outs
		u.rootContext.Send(n, actorstate.Init{Children: outs})
	}
	// Start the system
	for _, v := range u.nameToPID {
		u.rootContext.Send(v, actorstate.Start{})
	}
	return nil
}

func (u *Orchestrator) addPID(no actorstate.FlowActor) {
	props := actor.PropsFromProducer(func() actor.Actor { return no })
	pid := u.rootContext.Spawn(props)
	u.nameToPID[no.Name()] = pid
	u.pidToName[pid] = no.Name()
	u.nameToActor[no.Name()] = no
}

func (u *Orchestrator) GeneratePlantUML() string {
	sb := strings.Builder{}
	sb.WriteString("@startuml \n")
	for k, _ := range u.nameToPID {
		sb.WriteString(fmt.Sprintf("[%s] \n", k))
	}

	for parentName, parentPid := range u.nameToPID {
		children, found := u.parentToChildren[parentPid]
		if !found {
			continue
		}
		for _, child := range children {
			childName, found := u.pidToName[child]
			if !found {
				continue
			}
			sb.WriteString(fmt.Sprintf("[%s] -> [%s] \n", parentName, childName))
		}
	}
	sb.WriteString("@enduml \n")
	return sb.String()
}

func (u *Orchestrator) GenerateMermaid() string {
	sb := strings.Builder{}
	sb.WriteString("graph LR \n")
	for nodeName, _ := range u.nameToPID {
		act := u.nameToActor[nodeName]
		sb.WriteString(fmt.Sprintf("\t%s[%s - %T] \n", strings.ToTitle(nodeName), nodeName, act))
	}

	for parentName, parentPid := range u.nameToPID {
		children, found := u.parentToChildren[parentPid]
		if !found {
			continue
		}
		for _, child := range children {
			childName, found := u.pidToName[child]
			if !found {
				continue
			}
			sb.WriteString(fmt.Sprintf("\t%s --> %s \n", strings.ToTitle(parentName), strings.ToTitle(childName)))
		}
	}
	return sb.String()
}

func (u *Orchestrator) NodeList() []string {
	nodes := make([]string, 0)
	for n, _ := range u.nameToPID {
		nodes = append(nodes, n)
	}
	return nodes
}

func (u *Orchestrator) GetNodeStatus(name string) []byte {
	pid, found := u.nameToPID[name]
	if !found {
		return []byte("not found")
	}
	out, _ := u.rootContext.RequestFuture(pid, actorstate.State{}, 10*time.Second).Result()
	return out.([]byte)
}

func (u *Orchestrator) processNode(nodeCfg config.Node, global *types.Global) error {
	var err error
	var no actorstate.FlowActor
	if nodeCfg.MetricGenerator != nil {
		no, err = metrics.NewMetricGenerator(nodeCfg.Name, *nodeCfg.MetricGenerator, global)
	} else if nodeCfg.MetricFilter != nil {
		no, err = metrics.NewMetricFilter(nodeCfg.Name, *nodeCfg.MetricFilter)
	} else if nodeCfg.FakeMetricRemoteWrite != nil {
		no, err = remotewrites.NewFakeMetricRemoteWrite(nodeCfg.Name)
	} else if nodeCfg.LogFileWriter != nil {
		no, err = logs.NewFileWriter(nodeCfg.Name, *nodeCfg.LogFileWriter)
	} else if nodeCfg.Github != nil {
		no, err = integrations.NewGithub(nodeCfg.Name, nodeCfg.Github, *global)
	} else if nodeCfg.PrometheusRemoteWrite != nil {
		no, err = remotewrites.NewPrometheus(nodeCfg.Name, global, nodeCfg.PrometheusRemoteWrite)
	} else if nodeCfg.AgentLogs != nil {
		// AgentLogs is a special case
		return nil
	} else if nodeCfg.Credentials != nil {
		no, err = auth.NewCredentialsManager(nodeCfg.Name, nodeCfg.Credentials)
	}

	if err != nil {
		return err
	}
	if no == nil {
		return fmt.Errorf("unable to handle node named %s", nodeCfg.Name)
	}
	u.addPID(no)
	return nil
}
