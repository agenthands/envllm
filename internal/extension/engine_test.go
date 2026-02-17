package extension

import (
	"context"
	"testing"

	"github.com/agenthands/envllm/internal/ast"
)

func TestMappingEngine_Upgrade(t *testing.T) {
	e := NewMappingEngine(nil)
	
	// Register a dummy extension with a mapping
	m := Manifest{
		Name:    "web",
		Version: "0.2",
		Compat: Compatibility{
			Mappings: []Mapping{
				{
					ID:             "M001",
					ExtensionRange: "0.1",
					RewriteRuleID:  "RULE_RENAME_NAVIGATE",
				},
			},
		},
	}
	e.Register(m)

	prog := &ast.Program{
		Extensions: map[string]string{
			"web": "0.1",
		},
		Task: &ast.Task{Name: "test"},
	}

	rules, err := e.UpgradeProgram(context.Background(), prog)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}

	if len(rules) != 1 || rules[0] != "RULE_RENAME_NAVIGATE" {
		t.Errorf("expected rule RULE_RENAME_NAVIGATE, got %v", rules)
	}
}

func TestMappingEngine_NotFound(t *testing.T) {
	e := NewMappingEngine(nil)
	prog := &ast.Program{
		Extensions: map[string]string{
			"unknown": "1.0",
		},
	}

	_, err := e.UpgradeProgram(context.Background(), prog)
	if err == nil {
		t.Error("expected error for unknown extension, got nil")
	}
}
