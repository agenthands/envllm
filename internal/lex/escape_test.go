package lex

import (
	"testing"
)

func TestLexer_EscapedQuotes(t *testing.T) {
	input := `"{\"a\": 1}"`
	l := NewLexer("test.rlm", input)
	tok := l.NextToken()
	
	expected := `{"a": 1}`
	if tok.Type != TypeString {
		t.Errorf("expected string, got %v (val: %q)", tok.Type, tok.Value)
	}
	if tok.Value != expected {
		t.Errorf("expected %q, got %q", expected, tok.Value)
	}
}
