package lint

import (
	"testing"

	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/parse"
)

func TestLinter(t *testing.T) {
	tbl, _ := ops.LoadTable("../../assets/ops.json")
	lnt := NewLinter(tbl)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			"Correct order",
			"CELL test:\n  FIND_TEXT SOURCE PROMPT NEEDLE \"x\" MODE FIRST IGNORE_CASE true INTO out\n",
			false,
		},
		{
			"Wrong order",
			"CELL test:\n  FIND_TEXT NEEDLE \"x\" SOURCE PROMPT MODE FIRST IGNORE_CASE true INTO out\n",
			true,
		},
		{
			"Type mismatch",
			"CELL test:\n  WINDOW_TEXT SOURCE PROMPT CENTER \"wrong_type\" RADIUS 10 INTO out\n",
			true,
		},
		{
			"Variable Shadowing",
			"CELL test:\n  STATS SOURCE PROMPT INTO out\n  STATS SOURCE PROMPT INTO out\n",
			true,
		},
		{
			"JSON_GET on Struct",
			"CELL test:\n  STATS SOURCE PROMPT INTO stats: STRUCT\n  JSON_GET SOURCE stats PATH \"cost\" INTO cost\n",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lex.NewLexer("test.rlm", tt.input)
			p := parse.NewParser(l, parse.ModeCompat) 
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			errs := lnt.Lint(prog)
			if (len(errs) > 0) != tt.wantErr {
				for _, e := range errs {
					t.Logf("Linter Error: [%s] %s", e.Code, e.Message)
				}
				t.Errorf("Lint() error count = %d, wantErr %v", len(errs), tt.wantErr)
			}
		})
	}
}
