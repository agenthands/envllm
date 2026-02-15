package runtime

import (
	"context"
	"testing"

	"github.com/agenthands/envllm/internal/ast"
)

func FuzzSession(f *testing.F) {
	f.Add(int(10)) // Max statements seed
	
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 || n > 100 {
			return
		}
		
		policy := Policy{
			MaxStmtsPerCell: 50,
		}
		s := NewSession(policy, nil)
		
		stmts := []ast.Stmt{
			&ast.PrintStmt{Source: &ast.IntExpr{Value: 1}},
			&ast.AssertStmt{Cond: &ast.BoolExpr{Value: true}, Message: "ok"},
			&ast.SetFinalStmt{Source: &ast.BoolExpr{Value: false}},
		}
		
		cell := &ast.Cell{Name: "fuzz"}
		for i := 0; i < n; i++ {
			cell.Stmts = append(cell.Stmts, stmts[i%len(stmts)])
		}
		
		_ = s.ExecuteCell(context.Background(), cell)
	})
}
