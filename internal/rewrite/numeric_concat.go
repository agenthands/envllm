package rewrite

import (
	"context"
	"fmt"
	"github.com/agenthands/envllm/internal/ast"
)

type NumericConcatRule struct{}

func NewNumericConcatRule() *NumericConcatRule {
	return &NumericConcatRule{}
}

func (r *NumericConcatRule) ID() RuleID {
	return "RULE_NUMERIC_CONCAT_TO_TEXT"
}

func (r *NumericConcatRule) Description() string {
	return "Convert numeric or offset values to TEXT before using them in CONCAT_TEXT."
}

type concatVisitor struct {
	matches []MatchResult
}

func (v *concatVisitor) Visit(node ast.Node) ast.Visitor {
	if op, ok := node.(*ast.OpStmt); ok && op.OpName == "CONCAT_TEXT" {
		v.matches = append(v.matches, MatchResult{Node: op})
	}
	return v
}

func (r *NumericConcatRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	v := &concatVisitor{}
	ast.Walk(v, prog)
	if len(v.matches) == 0 {
		return nil, false
	}
	return v.matches, true
}

func (r *NumericConcatRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
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
				if op, ok := stmt.(*ast.OpStmt); ok && op.OpName == "CONCAT_TEXT" {
					// Check each argument A, B
					for i := range op.Args {
						kw := op.Args[i].Keyword
						if kw == "A" || kw == "B" {
							expr := op.Args[i].Value
							// If it's a literal INT, it needs TO_TEXT
							if _, ok := expr.(*ast.IntExpr); ok {
								varSeq++
								newName := fmt.Sprintf("text_val_%d", varSeq)
								
								toTextOp := &ast.OpStmt{
									OpName: "TO_TEXT",
									Args: []ast.KwArg{
										{Keyword: "VALUE", Value: expr},
									},
									Into:     newName,
									IntoType: "TEXT",
									Type:     "op",
								}
								newCell.Stmts = append(newCell.Stmts, toTextOp)
								op.Args[i].Value = &ast.IdentExpr{Name: newName, Kind: "IDENT"}
							}
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
