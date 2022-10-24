package main

import (
	"github.com/grafana/agent/component"
	_ "github.com/grafana/agent/component/all"
)

func main() {

	components := component.GetRegisteredComponents()
	for _, c := range components {

	}
}
