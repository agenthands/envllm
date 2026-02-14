package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/agenthands/envllm/internal/ast"
)

type mockTextStore struct {
	added []string
}

func (m *mockTextStore) Add(text string) TextHandle {
	m.added = append(m.added, text)
	return TextHandle{ID: "m1", Bytes: len(text)}
}
func (m *mockTextStore) Get(h TextHandle) (string, bool) { return "", false }
func (m *mockTextStore) Window(h TextHandle, center, radius int) (TextHandle, error) {
	return TextHandle{ID: "w1", Bytes: 10}, nil
}

func TestSession_ExecuteCell(t *testing.T) {
	policy := Policy{MaxStmtsPerCell: 10, MaxWallTime: time.Second}
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
	if err != nil { t.Fatalf("ExecuteCell failed: %v", err) }
	if s.Final == nil || s.Final.V != 42 { t.Errorf("expected Final value 42, got %v", s.Final) }
}

func TestSession_Budgets(t *testing.T) {
	policy := Policy{MaxStmtsPerCell: 1}
	s := NewSession(policy, nil)
	cell := &ast.Cell{Stmts: []ast.Stmt{&ast.PrintStmt{Source: &ast.IntExpr{Value: 1}}, &ast.PrintStmt{Source: &ast.IntExpr{Value: 2}}}}
	err := s.ExecuteCell(context.Background(), cell)
	if err == nil { t.Errorf("expected budget error") }
}

func TestSession_EvalExpr_Ident(t *testing.T) {
	s := NewSession(Policy{}, nil)
	s.Env.Define("x", Value{Kind: KindInt, V: 100})
	val, err := s.evalExpr(&ast.IdentExpr{Name: "x"})
	if err != nil { t.Fatalf("evalExpr failed: %v", err) }
	if val.V != 100 { t.Errorf("expected 100, got %v", val.V) }
	val, err = s.evalExpr(&ast.IdentExpr{Name: "y"})
	if err != nil { t.Fatalf("evalExpr failed for unknown ident: %v", err) }
	if val.Kind != KindString || val.V != "y" { t.Errorf("expected bareword string, got %v", val) }
}

func TestSession_ValidatePath(t *testing.T) {
	policy := Policy{
		AllowedReadPaths:  []string{"/tmp/rlm"},
		AllowedWritePaths: []string{"/tmp/rlm"},
	}
	s := NewSession(policy, nil)
	tests := []struct {
		path  string
		write bool
		want  bool
	}{
		{"/tmp/rlm/file.txt", false, true},
		{"/tmp/rlm_secret/file.txt", false, false},
		{"/tmp/rlm", true, true},
	}
	for _, tt := range tests {
		err := s.ValidatePath(tt.path, tt.write)
		if (err == nil) != tt.want { t.Errorf("ValidatePath(%q) = %v, want success=%v", tt.path, err, tt.want) }
	}
}
