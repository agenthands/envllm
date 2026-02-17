package rewrite

import (
	"context"
	"testing"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
)

func TestRewriteEngine_AutoRepair(t *testing.T) {
	// 1. Setup Table with an op that requires capability
	table := &ops.Table{
		Ops: map[string]*ops.Op{
			"READ_FILE": {
				Name:         "READ_FILE",
				Capabilities: []string{"fs_read"},
			},
		},
	}

	// 2. Setup Program with dot access and missing requires
	prog := &ast.Program{
		Task: &ast.Task{
			Name: "test",
			Body: []ast.BodyItem{
				&ast.Cell{
					Name: "main",
					Stmts: []ast.Stmt{
						&ast.OpStmt{
							OpName: "READ_FILE",
							Args: []ast.KwArg{
								{Keyword: "PATH", Value: &ast.IdentExpr{Name: "stats.path", Kind: "IDENT"}},
							},
							Into: "content",
						},
					},
				},
			},
		},
	}

	engine := NewEngine(DefaultRegistry(table))
	
	patched, applied, err := engine.AutoRepair(context.Background(), prog)
	if err != nil {
		t.Fatalf("AutoRepair failed: %v", err)
	}

	// Check applied rules
	hasDot := false
	hasReq := false
	for _, id := range applied {
		if id == "RULE_DOT_ACCESS_TO_GETTER" {
			hasDot = true
		}
		if id == "RULE_MISSING_REQUIRES" {
			hasReq = true
		}
	}

	if !hasDot {
		t.Error("RULE_DOT_ACCESS_TO_GETTER was not applied")
	}
	if !hasReq {
		t.Error("RULE_MISSING_REQUIRES was not applied")
	}

	// Check results
	if len(patched.Task.Body) < 2 {
		t.Errorf("expected at least 2 body items (Requirement + Cell), got %d", len(patched.Task.Body))
	}

	foundReq := false
	for _, item := range patched.Task.Body {
		if req, ok := item.(*ast.Requirement); ok && req.Capability == "fs_read" {
			foundReq = true
		}
	}
	if !foundReq {
		t.Error("Missing REQUIRES capability=\"fs_read\" in patched program")
	}
}
