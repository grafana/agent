package actorstate

import "github.com/AsynkronIT/protoactor-go/actor"

type Init struct {
	Children []*actor.PID
}

type Start struct{}

type Stop struct{}
