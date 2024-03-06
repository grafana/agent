package importsource

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/river/ast"
	"github.com/grafana/river/parser"
	"github.com/grafana/river/printer"
	"github.com/grafana/river/scanner"
	"github.com/grafana/river/token"
	"github.com/grafana/river/vm"
)

// ImportDir loads all river files in a directory into independent submodules.
type ImportDir struct {
	args            DirArguments
	managedOpts     component.Options
	eval            *vm.Evaluator
	onContentChange func(string)
	content         string
}

var _ ImportSource = (*ImportDir)(nil)

func NewImportDir(managedOpts component.Options, eval *vm.Evaluator, onContentChange func(string)) *ImportDir {
	opts := managedOpts
	return &ImportDir{
		managedOpts:     opts,
		eval:            eval,
		onContentChange: onContentChange,
		content:         dir_base,
	}
}

type DirArguments struct {
	Path string `river:"path,attr"`
}

func (im *ImportDir) Evaluate(scope *vm.Scope) error {
	level.Error(im.managedOpts.Logger).Log("msg", "UPDATE", "path", im.args.Path)
	var arguments DirArguments
	if err := im.eval.Evaluate(scope, &arguments); err != nil {
		return fmt.Errorf("decoding River: %w", err)
	}

	if reflect.DeepEqual(im.args, arguments) {
		return nil
	}
	im.args = arguments
	im.onContentChange(dir_base)

	return nil
}

// temporary placeholder config so we always start up correctly.
const dir_base = `declare "main" {
	export "const" {
	  value = 44  
	}
  }`

func (im *ImportDir) Run(ctx context.Context) error {
	level.Error(im.managedOpts.Logger).Log("msg", "RUN", "path", im.args.Path)
	im.findFiles()

	<-ctx.Done()
	return nil
}

func (im *ImportDir) findFiles() error {
	files := map[string]string{}
	fs, err := os.ReadDir(im.args.Path)
	if err != nil {
		return err
	}
	for _, info := range fs {
		if !info.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(info.Name(), ".river") {
			continue
		}
		dat, err := os.ReadFile(filepath.Join(im.args.Path, info.Name()))
		if err != nil {
			level.Error(im.managedOpts.Logger).Log("msg", "reading file", "err", err, "file", info.Name())
			continue
		}
		name, err := sanitizeName(info.Name())
		if err != nil {
			level.Error(im.managedOpts.Logger).Log("msg", "sanitizing file name", "err", err, "file", info.Name())
			continue
		}
		files[name] = string(dat)
	}
	fmt.Println("!!!!!!", files)
	newContent := im.buildDynamicModule(files)
	im.content = newContent
	im.onContentChange(newContent)
	return nil
}

func sanitizeName(s string) (string, error) {
	s = strings.TrimSuffix(s, ".river")
	return scanner.SanitizeIdentifier(s)
}

// CurrentHealth returns the health of the file component.
func (im *ImportDir) CurrentHealth() component.Health {
	return component.Health{
		Health: component.HealthTypeHealthy,
	}
}

func (im *ImportDir) buildDynamicModule(children map[string]string) string {
	bs := &ast.BlockStmt{
		Name:  []string{"declare"},
		Label: "main",
	}
	failCount := 0
	for name, content := range children {
		inner, err := innerModuleContent(content)
		if err != nil {
			level.Error(im.managedOpts.Logger).Log("msg", "invalid dynamic module", "name", name)
			failCount++
			continue
		}
		// name should already be sanitized
		modName := name
		// build text of inner module
		importStringStmt := &ast.BlockStmt{
			Name:  []string{"import", "string"},
			Label: modName,
			Body: ast.Body{
				&ast.AttributeStmt{
					Name: &ast.Ident{Name: "content"},
					Value: &ast.LiteralExpr{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", inner),
					},
				},
			},
		}
		bs.Body = append(bs.Body, importStringStmt)
		// now generate a usage of main:
		usageStmt := &ast.BlockStmt{
			Name:  []string{modName, "main"},
			Label: "main",
		}
		bs.Body = append(bs.Body, usageStmt)
	}
	bs.Body = append(bs.Body, &ast.BlockStmt{
		Name:  []string{"export"},
		Label: "failedModules",
		Body: ast.Body{
			&ast.AttributeStmt{
				Name: &ast.Ident{Name: "value"},
				Value: &ast.LiteralExpr{
					Kind:  token.NUMBER,
					Value: fmt.Sprint(failCount),
				},
			},
		},
	})
	buf := &bytes.Buffer{}
	printer.Fprint(buf, bs)
	fmt.Println(buf.String())
	return buf.String()
}

func innerModuleContent(f string) (string, error) {
	bs := &ast.BlockStmt{
		Name:  []string{"declare"},
		Label: "main",
	}
	file, err := parser.ParseFile("", []byte(f))
	if err != nil {
		return "", err
	}
	// todo: validate body way more
	bs.Body = file.Body
	buf := &bytes.Buffer{}
	printer.Fprint(buf, bs)
	return buf.String(), nil
}

func (im *ImportDir) DebugInfo() interface{} {
	return struct {
		Content string `river:"content,attr"`
	}{Content: im.content}
}
