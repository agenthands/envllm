package rewrite

import (
	"context"
	"fmt"
	"strings"
	"github.com/agenthands/envllm/internal/ast"
)

type DotAccessRule struct{}

func NewDotAccessRule() *DotAccessRule {
	return &DotAccessRule{}
}

func (r *DotAccessRule) ID() RuleID {
	return "RULE_DOT_ACCESS_TO_GETTER"
}

func (r *DotAccessRule) Description() string {
	return "Convert dot access (e.g. stats.cost) to explicit GET_FIELD or specialized getter calls."
}

type dotVisitor struct {
	matches []MatchResult
}

func (v *dotVisitor) Visit(node ast.Node) ast.Visitor {
	if ident, ok := node.(*ast.IdentExpr); ok {
		if strings.Contains(ident.Name, ".") {
			v.matches = append(v.matches, MatchResult{Node: ident})
		}
	}
	return v
}

func (r *DotAccessRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	v := &dotVisitor{}
	ast.Walk(v, prog)
	if len(v.matches) == 0 {
		return nil, false
	}
	return v.matches, true
}

func (r *DotAccessRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
	if prog.Task == nil {
		return prog, nil
	}

	newBody := []ast.BodyItem{}
	varSeq := 0

	for _, item := range prog.Task.Body {
		switch it := item.(type) {
		case *ast.Cell:
			newCell := &ast.Cell{Loc: it.Loc, Name: it.Name}
			for _, stmt := range it.Stmts {
				// Search for dot access in this statement's expressions
				found := false
				var targetIdent *ast.IdentExpr
				
				ast.Walk(&exprSearchVisitor{callback: func(e ast.Expr) {
					if id, ok := e.(*ast.IdentExpr); ok && strings.Contains(id.Name, ".") {
						targetIdent = id
						found = true
					}
				}}, stmt)

				if found {
					parts := strings.Split(targetIdent.Name, ".")
					obj := parts[0]
					field := parts[1]
					
					varSeq++
					newName := fmt.Sprintf("%s_%s_%d", obj, field, varSeq)
					
					// 1. Insert extraction op
					extractOp := &ast.OpStmt{
						OpName: "GET_FIELD",
						Args: []ast.KwArg{
							{Keyword: "SOURCE", Value: &ast.IdentExpr{Name: obj, Kind: "IDENT"}},
							{Keyword: "FIELD", Value: &ast.StringExpr{Value: field, Kind: "STRING"}},
						},
						Into:     newName,
						IntoType: "JSON",
						Type:     "op",
					}
					newCell.Stmts = append(newCell.Stmts, extractOp)
					
					// 2. Replace original argument in the statement
					// This is a bit tricky with generic walk. 
					// For now, we'll manually handle common stmts.
					switch s := stmt.(type) {
					case *ast.PrintStmt:
						if id, ok := s.Source.(*ast.IdentExpr); ok && id == targetIdent {
							s.Source = &ast.IdentExpr{Name: newName, Kind: "IDENT"}
						}
					case *ast.OpStmt:
						for i := range s.Args {
							if id, ok := s.Args[i].Value.(*ast.IdentExpr); ok && id == targetIdent {
								s.Args[i].Value = &ast.IdentExpr{Name: newName, Kind: "IDENT"}
							}
						}
					case *ast.SetFinalStmt:
						if id, ok := s.Source.(*ast.IdentExpr); ok && id == targetIdent {
							s.Source = &ast.IdentExpr{Name: newName, Kind: "IDENT"}
						}
					}
				}
				newCell.Stmts = append(newCell.Stmts, stmt)
			}
			newBody = append(newBody, newCell)
		default:
			newBody = append(newBody, item)
		}
	}

	prog.Task.Body = newBody
	return prog, nil
}

type exprSearchVisitor struct {
	callback func(ast.Expr)
}

func (v *exprSearchVisitor) Visit(node ast.Node) ast.Visitor {
	if e, ok := node.(ast.Expr); ok {
		v.callback(e)
	}
	return v
}
