package fmt

import (
	"testing"

	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/parse"
)

func TestFormat(t *testing.T) {
	input := "RLMDSL 0.1\n\nCELL test:\n  STATS SOURCE PROMPT INTO out: JSON\n  PRINT SOURCE out\n  SET_FINAL SOURCE null\n"
	
	l := lex.NewLexer("test.rlm", input)
	p := parse.NewParser(l, parse.ModeStrict)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	formatted := Format(prog)
	if formatted != input {
		t.Errorf("Format mismatch.\nGot:\n%s\nWant:\n%s", formatted, input)
	}

	// Idempotence check
	l2 := lex.NewLexer("formatted.rlm", formatted)
	p2 := parse.NewParser(l2, parse.ModeStrict)
	prog2, err := p2.Parse()
	if err != nil {
		t.Fatalf("Parse of formatted failed: %v", err)
	}
	formatted2 := Format(prog2)
	if formatted2 != formatted {
		t.Errorf("Format not idempotent")
	}
}
