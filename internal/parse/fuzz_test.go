package parse

import (
	"testing"
	"github.com/agenthands/rlm-go/internal/lex"
)

func FuzzParser(f *testing.F) {
	seed := []string{
		"RLMDSL 0.1\nCELL plan:\n  STATS SOURCE PROMPT INTO stats\n",
		"CELL solve:\n  SET_FINAL SOURCE \"done\"\n",
	}
	for _, s := range seed {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		l := lex.NewLexer("fuzz.rlm", input)
		p := NewParser(l)
		p.Parse()
	})
}
