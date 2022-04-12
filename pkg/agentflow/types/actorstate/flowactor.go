package actorstate

import "github.com/AsynkronIT/protoactor-go/actor"

type FlowActor interface {
	actor.Actor
	Name() string
	PID() *actor.PID
	AllowableInputs() []InOutType
	Output() InOutType
}

type InOutType = int

const (
	Metrics InOutType = iota
	Targets
	Credentials
	Logs
	None
)
