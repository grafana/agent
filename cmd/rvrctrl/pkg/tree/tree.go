package tree

import (
	"bytes"
	"fmt"

	"github.com/grafana/agent/pkg/river/printer"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/parser"
)

type Tree struct {
	file *ast.File
}

func (t *Tree) Parse(name string, data []byte) error {
	file, err := parser.ParseFile(name, data)
	t.file = file
	return err
}

func (t *Tree) AddComponent(rawComponent []byte) error {
	comp, err := parser.ParseFile("", rawComponent)
	if err != nil {
		return err
	}
	if len(comp.Body) == 0 {
		return fmt.Errorf("component not found for %s", string(rawComponent))
	}
	if len(comp.Body) > 1 {
		return fmt.Errorf("two many components found for %s", string(rawComponent))
	}

	t.file.Body = append(t.file.Body, comp.Body[0])
	return nil
}

func (t *Tree) Print() (string, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, t.file); err != nil {
		return "", err
	}

	// Add a newline at the endi
	_, _ = buf.Write([]byte{'\n'})
	return buf.String(), nil
}
