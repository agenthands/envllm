package parse

import (
	"testing"

	"github.com/agenthands/envllm/internal/lex"
)

func TestParseHeaders(t *testing.T) {
	input := `DIALECT envllm=0.2.2
EXT web=0.1
EXT net=0.3

TASK test:
  OUTPUT out
`
	l := lex.NewLexer("test.rlm", input)
	p := NewParser(l, ModeStrict)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if prog.Dialect != "envllm=0.2.2" {
		t.Errorf("expected dialect envllm=0.2.2, got %q", prog.Dialect)
	}

	if len(prog.Extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(prog.Extensions))
	}

	if v, ok := prog.Extensions["web"]; !ok || v != "0.1" {
		t.Errorf("expected web=0.1, got %q", v)
	}
	if v, ok := prog.Extensions["net"]; !ok || v != "0.3" {
		t.Errorf("expected net=0.3, got %q", v)
	}
}
