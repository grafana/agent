package printer

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/river/ast"
	"github.com/grafana/agent/pkg/river/token"
)

// A walker walks an AST and sends lexical tokens and formatting information to
// a printer.
type walker struct {
	p *printer
}

func (w *walker) Walk(node ast.Node) error {
	switch node := node.(type) {
	case *ast.File:
		w.walkFile(node)
	case ast.Body:
		w.walkStmts(node)
	case ast.Stmt:
		w.walkStmt(node)
	case ast.Expr:
		w.walkExpr(node)
	default:
		return fmt.Errorf("unsupported node type %T", node)
	}

	return nil
}

func (w *walker) walkFile(f *ast.File) {
	w.p.SetComments(f.Comments)
	w.walkStmts(f.Body)
}

func (w *walker) walkStmts(ss []ast.Stmt) {
	for i, s := range ss {
		var addedSpacing bool

		// Two blocks should always be separated by a blank line.
		if _, isBlock := s.(*ast.BlockStmt); i > 0 && isBlock {
			w.p.Write(wsNewline)
			addedSpacing = true
		}

		// A blank line should always be added if there is a blank line in the
		// source between two statements.
		if i > 0 && !addedSpacing {
			var (
				prevLine = ast.EndPos(ss[i-1]).Position().Line
				curLine  = ast.StartPos(ss[i-0]).Position().Line

				lineDiff = curLine - prevLine
			)

			if lineDiff > 1 {
				w.p.Write(wsFormfeed)
			}
		}

		w.walkStmt(s)

		// Statements which cross multiple lines don't belong to the same row run.
		// Add a formfeed to start a new row run if the node crossed more than one
		// line, otherwise add the normal newline.
		if nodeLines(s) > 1 {
			w.p.Write(wsFormfeed)
		} else {
			w.p.Write(wsNewline)
		}
	}
}

func nodeLines(n ast.Node) int {
	var (
		startLine = ast.StartPos(n).Position().Line
		endLine   = ast.EndPos(n).Position().Line
	)

	return endLine - startLine + 1
}

func (w *walker) walkStmt(s ast.Stmt) {
	switch s := s.(type) {
	case *ast.AttributeStmt:
		w.walkAttributeStmt(s)
	case *ast.BlockStmt:
		w.walkBlockStmt(s)
	}
}

func (w *walker) walkAttributeStmt(s *ast.AttributeStmt) {
	w.p.Write(s.Name.NamePos, s.Name, wsVTab, token.ASSIGN, wsBlank)
	w.walkExpr(s.Value)
}

func (w *walker) walkBlockStmt(s *ast.BlockStmt) {
	joined := strings.Join(s.Name, ".")

	// TODO(rfratto): Should blocks have a oneline format if they're short or
	// empty? e.g.: `empty_block { attr = 5 }`, `empty_block {}`

	w.p.Write(
		s.NamePos,
		&ast.IdentifierExpr{Name: joined, NamePos: s.NamePos},
	)

	if s.Label != "" {
		label := fmt.Sprintf("%q", s.Label)

		w.p.Write(
			wsBlank,
			&ast.LiteralExpr{Kind: token.STRING, Value: label},
		)
	}

	w.p.Write(
		wsBlank,
		s.LCurly, token.LCURLY, wsIndent,
		wsNewline,
	)

	w.walkStmts(s.Body)

	w.p.Write(wsUnindent, s.RCurly, token.RCURLY)
}

func (w *walker) walkExpr(e ast.Expr) {
	switch e := e.(type) {
	case *ast.LiteralExpr:
		w.p.Write(e.ValuePos, e)

	case *ast.ArrayExpr:
		w.walkArrayExpr(e)

	case *ast.ObjectExpr:
		w.walkObjectExpr(e)

	case *ast.IdentifierExpr:
		w.p.Write(e.NamePos, e)

	case *ast.AccessExpr:
		w.walkExpr(e.Value)
		w.p.Write(token.DOT, e.Name)

	case *ast.IndexExpr:
		w.walkExpr(e.Value)
		w.p.Write(e.LBrack, token.LBRACK)
		w.walkExpr(e.Index)
		w.p.Write(e.RBrack, token.RBRACK)

	case *ast.CallExpr:
		// TODO(rfratto): allow arguments to be on a new line
		w.walkExpr(e.Value)
		w.p.Write(token.LPAREN)
		for i, arg := range e.Args {
			w.walkExpr(arg)

			if i+1 < len(e.Args) {
				w.p.Write(token.COMMA, wsBlank)
			}
		}
		w.p.Write(token.RPAREN)

	case *ast.UnaryExpr:
		w.p.Write(e.KindPos, e.Kind)
		w.walkExpr(e.Value)

	case *ast.BinaryExpr:
		// TODO(rfratto):
		//
		//   1. allow RHS to be on a new line
		//
		//   2. remove spacing between some operators to make precedence
		//      clearer like Go does
		w.walkExpr(e.Left)
		w.p.Write(wsBlank, e.KindPos, e.Kind, wsBlank)
		w.walkExpr(e.Right)

	case *ast.ParenExpr:
		w.p.Write(token.LPAREN)
		w.walkExpr(e.Inner)
		w.p.Write(token.RPAREN)
	}
}

func (w *walker) walkArrayExpr(e *ast.ArrayExpr) {
	w.p.Write(e.LBrack, token.LBRACK, wsIndent)

	prevPos := e.LBrack

	for i := 0; i < len(e.Elements); i++ {
		elementPos := ast.StartPos(e.Elements[i])

		// Add a newline if this element starts on a different line from the last
		// element.
		if differentLines(prevPos, elementPos) {
			w.p.Write(wsFormfeed)
		} else if i > 0 {
			// Make sure a space is injected before the next element if two
			// successive elements are on the same line.
			w.p.Write(wsBlank)
		}
		prevPos = elementPos

		// Write the expression.
		w.walkExpr(e.Elements[i])

		// Always add commas in between successive elements.
		if i+1 < len(e.Elements) {
			w.p.Write(token.COMMA)
		}
	}

	// If the closing bracket is on a different line than the final element,
	// we need to add a trailing comma.
	if len(e.Elements) > 0 && differentLines(prevPos, e.RBrack) {
		w.p.Write(token.COMMA, wsFormfeed)
	}

	w.p.Write(wsUnindent, e.RBrack, token.RBRACK)
}

func (w *walker) walkObjectExpr(e *ast.ObjectExpr) {
	w.p.Write(token.LCURLY, wsIndent)

	prevPos := e.LCurly

	for i := 0; i < len(e.Fields); i++ {
		field := e.Fields[i]
		elementPos := ast.StartPos(field.Name)

		// Add a newline if this element starts on a different line from the last
		// element.
		if differentLines(prevPos, elementPos) {
			w.p.Write(wsFormfeed)
		} else if i > 0 {
			// Make sure a space is injected before the next element if two successive
			// elements are on the same line.
			w.p.Write(wsBlank)
		}
		prevPos = elementPos

		w.p.Write(field.Name.NamePos)

		// Write the field.
		if field.Quoted {
			w.p.Write(&ast.LiteralExpr{
				Kind:     token.STRING,
				ValuePos: field.Name.NamePos,
				Value:    fmt.Sprintf("%q", field.Name.Name),
			})
		} else {
			w.p.Write(field.Name)
		}

		w.p.Write(wsVTab, token.ASSIGN, wsBlank)
		w.walkExpr(field.Value)

		// Always add commas in between successive elements.
		if i+1 < len(e.Fields) {
			w.p.Write(token.COMMA)
		}
	}

	// If the closing bracket is on a different line than the final element,
	// we need to add a trailing comma.
	if len(e.Fields) > 0 && differentLines(prevPos, e.RCurly) {
		w.p.Write(token.COMMA, wsFormfeed)
	}

	w.p.Write(wsUnindent, e.RCurly, token.RCURLY)
}

// differentLines returns true if a and b are on different lines.
func differentLines(a, b token.Pos) bool {
	return a.Position().Line != b.Position().Line
}
