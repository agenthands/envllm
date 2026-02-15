package lex

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `RLMDSL 0.1
CELL plan:
  STATS SOURCE PROMPT INTO stats
SET_FINAL ASSERT PRINT true false
"hello world"
@`
	l := NewLexer("test.rlm", input)
	
	expected := []struct {
		typ   Type
		val   string
	}{
		{TypeRLMDSL, "RLMDSL"},
		{TypeIdent, "0.1"},
		{TypeNewline, "\n"},
		{TypeCELL, "CELL"},
		{TypeIdent, "plan"},
		{TypeColon, ":"},
		{TypeNewline, "\n"},
		{TypeIdent, "STATS"},
		{TypeIdent, "SOURCE"},
		{TypeIdent, "PROMPT"},
		{TypeINTO, "INTO"},
		{TypeIdent, "stats"},
		{TypeNewline, "\n"},
		{TypeSET_FINAL, "SET_FINAL"},
		{TypeASSERT, "ASSERT"},
		{TypePRINT, "PRINT"},
		{TypeBool, "true"},
		{TypeBool, "false"},
		{TypeNewline, "\n"},
		{TypeString, "hello world"},
		{TypeNewline, "\n"},
		{TypeError, "@"},
		{TypeEOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Errorf("[%d] expected type %v, got %v (val: %q)", i, exp.typ, tok.Type, tok.Value)
		}
		if tok.Value != exp.val {
			t.Errorf("[%d] expected value %q, got %q", i, exp.val, tok.Value)
		}
	}
}

func TestLexer_PeekChar(t *testing.T) {
	l := NewLexer("test.rlm", "AB")
	if l.peekChar() != 'B' {
		t.Errorf("expected 'B', got %q", l.peekChar())
	}
}

func TestLoc_String(t *testing.T) {
	loc := Loc{File: "test.rlm", Line: 10, Col: 5}
	expected := "test.rlm:10:5"
	if loc.String() != expected {
		t.Errorf("expected %q, got %q", expected, loc.String())
	}
}

func TestLexer_EOFIdent(t *testing.T) {
	input := "error_id error_snippet error_pos"
	l := NewLexer("test.rlm", input)
	
	expected := []string{"error_id", "error_snippet", "error_pos"}
	for _, exp := range expected {
		tok := l.NextToken()
		if tok.Value != exp {
			t.Errorf("expected %q, got %q", exp, tok.Value)
		}
	}
}
