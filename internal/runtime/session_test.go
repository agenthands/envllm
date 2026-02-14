package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/agenthands/rlm-go/internal/ast"
	"github.com/agenthands/rlm-go/internal/lex"
)

// MockTextStore for testing
type mockTextStore struct {
	added []string
}

func (m *mockTextStore) Add(text string) TextHandle {
	m.added = append(m.added, text)
	return TextHandle{ID: "m1", Bytes: len(text)}
}
func (m *mockTextStore) Get(h TextHandle) (string, bool) { return "", false }

func TestSession_ExecuteCell(t *testing.T) {
	policy := Policy{
		MaxStmtsPerCell: 10,
		MaxWallTime:     time.Second,
	}
	ts := &mockTextStore{}
	s := NewSession(policy, ts)

	cell := &ast.Cell{
		Name: "test",
		Stmts: []ast.Stmt{
			&ast.PrintStmt{Source: &ast.StringExpr{Value: "hello"}},
			&ast.AssertStmt{Cond: &ast.BoolExpr{Value: true}, Message: "ok"},
			&ast.SetFinalStmt{Source: &ast.IntExpr{Value: 42}},
		},
	}

	err := s.ExecuteCell(context.Background(), cell)
	if err != nil {
		t.Fatalf("ExecuteCell failed: %v", err)
	}

	if s.Final == nil || s.Final.V != 42 {
		t.Errorf("expected Final value 42, got %v", s.Final)
	}

	if len(ts.added) != 1 || ts.added[0] != "hello" {
		t.Errorf("expected 'hello' added to TextStore")
	}
}

func TestSession_WallTime(t *testing.T) {
	policy := Policy{
		MaxWallTime: time.Nanosecond, // Extremely small limit
	}
	s := NewSession(policy, nil)

	cell := &ast.Cell{
		Stmts: []ast.Stmt{
			&ast.PrintStmt{Source: &ast.IntExpr{Value: 1}},
			&ast.PrintStmt{Source: &ast.IntExpr{Value: 2}},
		},
	}
	
	err := s.ExecuteCell(context.Background(), cell)
	if err == nil {
		t.Errorf("expected wall time budget error")
	}
}

func TestSession_Errors(t *testing.T) {
	s := NewSession(Policy{}, nil)

	tests := []struct {
		name string
		stmt ast.Stmt
	}{
		{"Assert Fail", &ast.AssertStmt{Cond: &ast.BoolExpr{Value: false}, Message: "fail"}},
		{"Assert Type", &ast.AssertStmt{Cond: &ast.IntExpr{Value: 1}}},
		{"Undefined Op", &ast.OpStmt{OpName: "NOP"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ExecuteStmt(context.Background(), tt.stmt)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestEnv_Vars(t *testing.T) {
	e := NewEnv()
	e.Define("a", Value{Kind: KindInt, V: 1})
	vars := e.Vars()
	if len(vars) != 1 || vars["a"].V != 1 {
		t.Errorf("Vars() failed")
	}
}

func TestSession_Budgets(t *testing.T) {
	policy := Policy{
		MaxStmtsPerCell: 1,
	}
	s := NewSession(policy, nil)

	cell := &ast.Cell{
		Stmts: []ast.Stmt{
			&ast.PrintStmt{Source: &ast.IntExpr{Value: 1}},
			&ast.PrintStmt{Source: &ast.IntExpr{Value: 2}},
		},
	}

	err := s.ExecuteCell(context.Background(), cell)
	if err == nil {
		t.Errorf("expected budget error")
	}
}

func TestSession_EvalExpr_Ident(t *testing.T) {
	s := NewSession(Policy{}, nil)
	s.Env.Define("x", Value{Kind: KindInt, V: 100})

	val, err := s.evalExpr(&ast.IdentExpr{Name: "x"})
	if err != nil {
		t.Fatalf("evalExpr failed: %v", err)
	}
	if val.V != 100 {
		t.Errorf("expected 100, got %v", val.V)
	}

	_, err = s.evalExpr(&ast.IdentExpr{Name: "y", Loc: lex.Loc{File: "f", Line: 1}})
	if err == nil {
		t.Errorf("expected error for undefined variable")
	}
}

func TestSession_EvalExpr_Errors(t *testing.T) {
	s := NewSession(Policy{}, nil)
	
	// Test nil TextStore for StringExpr
	val, _ := s.evalExpr(&ast.StringExpr{Value: "raw"})
	if val.Kind != KindJSON {
		t.Errorf("expected KindJSON for raw string when TextStore is nil")
	}

	// Test unknown expression type
	type unknownExpr struct{ ast.Expr }
	_, err := s.evalExpr(&unknownExpr{})
	if err == nil {
		t.Errorf("expected error for unknown expression type")
	}
}

func TestSession_ExecuteStmt_Errors(t *testing.T) {
	s := NewSession(Policy{}, nil)
	
	// Unknown statement type
	type unknownStmt struct{ ast.Stmt }
	err := s.ExecuteStmt(context.Background(), &unknownStmt{})
	if err == nil {
		t.Errorf("expected error for unknown statement type")
	}
	
	// SetFinal error (undefined var)
	err = s.ExecuteStmt(context.Background(), &ast.SetFinalStmt{Source: &ast.IdentExpr{Name: "undef"}})
	if err == nil {
		t.Errorf("expected error for SetFinal with undefined var")
	}

	// Print error (undefined var)
	err = s.ExecuteStmt(context.Background(), &ast.PrintStmt{Source: &ast.IdentExpr{Name: "undef"}})
	if err == nil {
		t.Errorf("expected error for Print with undefined var")
	}
}
