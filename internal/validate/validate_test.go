package validate

import (
	"testing"
	"github.com/agenthands/rlm-go/internal/ast"
	"github.com/agenthands/rlm-go/internal/lex"
	"github.com/agenthands/rlm-go/internal/parse"
)

func TestValidator(t *testing.T) {
	v, err := NewValidator("../../schemas/ast.schema.json")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	input := `RLMDSL 0.1
CELL plan:
  STATS SOURCE PROMPT INTO stats
`
	l := lex.NewLexer("test.rlm", input)
	p := parse.NewParser(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if err := v.ValidateAST(prog); err != nil {
		t.Errorf("Validation error: %v", err)
	}
}

func TestValidator_DuplicateCell(t *testing.T) {
	v, err := NewValidator("../../schemas/ast.schema.json")
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	input := `RLMDSL 0.1
CELL plan:
  PRINT SOURCE "a"
CELL plan:
  PRINT SOURCE "b"
`
	l := lex.NewLexer("test.rlm", input)
	p := parse.NewParser(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if err := v.ValidateAST(prog); err == nil {
		t.Error("Expected error for duplicate cell names, got nil")
	}
}

func TestValidator_InvalidSchema(t *testing.T) {
	v, _ := NewValidator("../../schemas/ast.schema.json")
	// Missing version and cells due to omitempty
	prog := &ast.Program{}
	if err := v.ValidateAST(prog); err == nil {
		t.Error("expected schema validation error for missing required fields, got nil")
	}
}

func TestNewValidator_Error(t *testing.T) {
	_, err := NewValidator("non-existent.json")
	if err == nil {
		t.Error("expected error for non-existent schema, got nil")
	}
}
