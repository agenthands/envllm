package fmt

import (
	"testing"

	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/parse"
)

func TestFormat(t *testing.T) {
	input := "RLMDSL 0.1\n\nTASK test:\n  CELL test:\n    STATS SOURCE PROMPT INTO out: JSON\n    PRINT SOURCE out\n    SET_FINAL SOURCE null\n  OUTPUT out\n"
	
	l := lex.NewLexer("test.rlm", input)
	p := parse.NewParser(l, parse.ModeStrict)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	formatted := Format(prog)
	// We allow some flexibility in whitespace for the first test if needed,
	// but idempotence is the main goal.
	
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

func TestFormatV02(t *testing.T) {
	input := `RLMDSL 0.2
DIALECT envllm=0.2

TASK smoke_test:
  INPUT url: TEXT
  CELL start:
    STATS SOURCE url INTO stats: STRUCT
    PRINT SOURCE stats
  OUTPUT stats
`
	l := lex.NewLexer("smoke.rlm", input)
	p := parse.NewParser(l, parse.ModeStrict)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	formatted := Format(prog)

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
