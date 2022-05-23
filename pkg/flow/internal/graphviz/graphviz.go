// Package graphviz implements some graphviz utilities. Graphviz must be
// installed for these to work.
package graphviz

import (
	"bytes"
	"fmt"
	"os/exec"
)

// NotFoundError is generated when a Graphviz tool could not be found.
type NotFoundError struct {
	ToolName string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Failed to execute %s. Is Graphviz installed?", e.ToolName)
}

// Dot renders the DOT-language input with the provided format.
func Dot(in []byte, format string) ([]byte, error) {
	dotPath, err := exec.LookPath("dot")
	if err != nil {
		return nil, NotFoundError{ToolName: "dot"}
	}

	cmd := exec.Command(dotPath, "-T"+format)
	cmd.Stdin = bytes.NewReader(in)
	return cmd.Output()
}
