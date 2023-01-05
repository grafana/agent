// Package exportedcomments exposes an Analyzer which will validate that all
// exported identifiers have comments.
package exportedcomments

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer implements the exportedcomments analyzer.
var Analyzer = &analysis.Analyzer{
	Name: "exportedcomments",
	Doc:  "ensure eported identifiers have documentation",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		// Ignore test and generated files.
		if strings.HasSuffix(pass.Fset.File(f.Pos()).Name(), "_test.go") {
			continue
		} else if isGeneratedFile(f) {
			continue
		}

		for _, decl := range f.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				lintFunction(pass, decl)
			case *ast.GenDecl:
				lintGenDecl(pass, decl)
			}
		}
	}

	return nil, nil
}

func isGeneratedFile(file *ast.File) bool {
	var (
		generatedPrefix = "// Code generated"
		generatedSuffix = " DO NOT EDIT."
	)

	for _, comment := range file.Comments {
		for _, line := range comment.List {
			if strings.HasPrefix(line.Text, generatedPrefix) && strings.HasSuffix(line.Text, generatedSuffix) {
				return true
			}
		}
	}

	return false
}

func lintFunction(pass *analysis.Pass, fn *ast.FuncDecl) {
	// Ignore non-exported functions.
	if !ast.IsExported(fn.Name.Name) {
		return
	}

	// Ignore exported functions where the receiver is non-exported.
	// For example:
	//
	//   func (a unexportedType) ExportedFunction() {}
	if fn.Recv != nil {
		typ := pass.TypesInfo.Types[fn.Recv.List[0].Type]
		if !isExportedReceiverType(typ.Type) {
			return
		}

		// golint didn't require doc comments on the implementation of
		// sort.Interface, so we emulate the same check here.
		switch fn.Name.Name {
		case "Len", "Less", "Swap":
			if types.Implements(typ.Type, sortInterface) {
				return
			}
		}
	}

	if fn.Doc == nil {
		pass.Report(analysis.Diagnostic{
			Pos:     fn.Pos(),
			Message: fmt.Sprintf("exported function %s should have comment or be unexported", fn.Name.Name),
		})
	}
}

func isExportedReceiverType(ty types.Type) bool {
	switch ty := ty.(type) {
	case *types.Pointer:
		return isExportedReceiverType(ty.Elem())
	case *types.Named:
		return ty.Obj().Exported()
	default:
		// This shouldn't be possible to hit for valid functions, since valid
		// receivers have to be to named types or pointers to named types.
		return false
	}
}

func lintGenDecl(pass *analysis.Pass, fn *ast.GenDecl) {
	// Ignore any gen decl with comments:
	//
	//   // Comment
	//   var (
	//     SomeVariable1 bool
	//     SomeVariable2 int
	//   )
	if fn.Doc != nil {
		return
	}

	for _, spec := range fn.Specs {
		res := analyzeSpec(spec)

		// Even if the group didn't have a comment, the individual specification
		// still might:
		//
		//   var (
		//     // Comment
		//     SomeVariable1 bool
		//   )
		//
		// We only want to report an error if there's not a comment on the grouping
		// AND on the individual spec.
		if res.Exported && !res.HasDoc {
			pass.Report(analysis.Diagnostic{
				Pos:     spec.Pos(),
				Message: fmt.Sprintf("exported %s %s should have comment or be unexported", res.Kind, res.Name),
			})
		}
	}
}

type specAnalysis struct {
	Name     string // Name of the specification
	Kind     string // Kind of the specification (identifier, type)
	HasDoc   bool   // Whether the specification has docs
	Exported bool   // Whether the specification is exported
}

func analyzeSpec(spec ast.Spec) specAnalysis {
	var analysis specAnalysis

	switch spec := spec.(type) {
	case *ast.ValueSpec:
		analysis.Name = spec.Names[0].Name
		analysis.Kind = "identifier"
		analysis.HasDoc = (spec.Doc != nil)
	case *ast.TypeSpec:
		analysis.Name = spec.Name.Name
		analysis.Kind = "type"
		analysis.HasDoc = (spec.Doc != nil)
	}

	analysis.Exported = ast.IsExported(analysis.Name)
	return analysis
}
