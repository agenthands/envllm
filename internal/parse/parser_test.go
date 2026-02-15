package parse

import (
	"strings"
	"testing"
	"github.com/agenthands/envllm/internal/lex"
)

func TestParser_Basic(t *testing.T) {
	input := `RLMDSL 0.1
CELL plan:
  STATS SOURCE PROMPT INTO stats
`
	l := lex.NewLexer("test.rlm", input)
	p := NewParser(l, ModeCompat)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if prog.Version != "0.1" {
		t.Errorf("expected version 0.1, got %s", prog.Version)
	}

	if len(prog.Cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(prog.Cells))
	}

	cell := prog.Cells[0]
	if cell.Name != "plan" {
		t.Errorf("expected cell name 'plan', got %s", cell.Name)
	}

	if len(cell.Stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(cell.Stmts))
	}
}

func TestParser_Full(t *testing.T) {
	input := `RLMDSL 0.1
CELL plan:
  STATS SOURCE PROMPT INTO stats
  FIND_TEXT SOURCE PROMPT NEEDLE "login" MODE FIRST IGNORE_CASE true INTO pos

CELL solve:
  ASSERT COND true MESSAGE "it works"
  PRINT SOURCE pos
  WINDOW_TEXT SOURCE PROMPT CENTER 100 RADIUS 50 INTO snippet
  SET_FINAL SOURCE "done"
`
	l := lex.NewLexer("full.rlm", input)
	p := NewParser(l, ModeCompat)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(prog.Cells) != 2 {
		t.Fatalf("expected 2 cells, got %d", len(prog.Cells))
	}

	cell1 := prog.Cells[0]
	if len(cell1.Stmts) != 2 {
		t.Errorf("cell 1: expected 2 stmts, got %d", len(cell1.Stmts))
	}

	cell2 := prog.Cells[1]
	if len(cell2.Stmts) != 4 {
		t.Errorf("cell 2: expected 4 stmts, got %d", len(cell2.Stmts))
	}
}

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Missing version", "RLMDSL\n"},
		{"Missing newline after header", "RLMDSL 0.1 CELL plan:"},
		{"Expected CELL", "NOTCELL name:"},
		{"Missing cell name", "CELL :"},
		{"Missing colon", "CELL plan\n"},
		{"Unexpected token in cell", "CELL plan:\n INVALID"},
		{"Op missing INTO", "CELL plan:\n OP SOURCE x\n"},
		{"Op missing ident after INTO", "CELL plan:\n OP SOURCE x INTO\n"},
		{"SetFinal missing SOURCE", "CELL plan:\n SET_FINAL x\n"},
		{"Assert missing COND", "CELL plan:\n ASSERT x\n"},
		{"Assert missing MESSAGE", "CELL plan:\n ASSERT COND true x\n"},
		{"Assert missing string", "CELL plan:\n ASSERT COND true MESSAGE 123\n"},
		{"Print missing SOURCE", "CELL plan:\n PRINT x\n"},
		{"Invalid expression", "CELL plan:\n PRINT SOURCE @\n"},
		{"Invalid indentation", "CELL plan:\nPRINT SOURCE x\n"},
		{"Too much indentation", "CELL plan:\n   PRINT SOURCE x\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := ModeCompat
			if strings.Contains(tt.name, "indentation") {
				mode = ModeStrict
			}
			l := lex.NewLexer("err.rlm", tt.input)
			p := NewParser(l, mode)
			_, err := p.Parse()
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.input)
			}
		})
	}
}

func TestParser_StrictMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			"Valid strict",
			"RLMDSL 0.1\nCELL test:\n  STATS SOURCE PROMPT INTO out: JSON\n",
			false,
		},
		{
			"Invalid indentation in strict",
			"RLMDSL 0.1\nCELL test:\nSTATS SOURCE PROMPT INTO out: JSON\n",
			true,
		},
		{
			"Missing type in strict",
			"RLMDSL 0.1\nCELL test:\n  STATS SOURCE PROMPT INTO out\n",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lex.NewLexer("test.rlm", tt.input)
			p := NewParser(l, ModeStrict)
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("ModeStrict Parse(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
