package main

import (
	"context"
	"testing"

	"github.com/agenthands/rlm-go/internal/lex"
	"github.com/agenthands/rlm-go/internal/ops"
	"github.com/agenthands/rlm-go/internal/parse"
	"github.com/agenthands/rlm-go/internal/runtime"
	"github.com/agenthands/rlm-go/internal/store"
)

func TestEndToEnd(t *testing.T) {
	// 1. Setup
	tbl, err := ops.LoadTable("../../assets/ops.json")
	if err != nil {
		t.Fatalf("LoadTable failed: %v", err)
	}
	reg := ops.NewRegistry(tbl)
	ts := store.NewTextStore()
	s := runtime.NewSession(runtime.Policy{MaxStmtsPerCell: 100}, ts)
	s.Dispatcher = reg

	// 2. Prepare Prompt
	prompt := "The secret code is 12345. Use it wisely."
	ph := ts.Add(prompt)
	s.Env.Define("PROMPT", runtime.Value{Kind: runtime.KindText, V: ph})

	// 3. Program
	code := `RLMDSL 0.1
CELL find:
  FIND_TEXT SOURCE PROMPT NEEDLE "12345" MODE "FIRST" IGNORE_CASE false INTO pos
  WINDOW_TEXT SOURCE PROMPT CENTER pos RADIUS 5 INTO snippet
  SET_FINAL SOURCE snippet
`
	l := lex.NewLexer("test.rlm", code)
	p := parse.NewParser(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 4. Execute
	for _, cell := range prog.Cells {
		err = s.ExecuteCell(context.Background(), cell)
		if err != nil {
			t.Fatalf("ExecuteCell %s failed: %v", cell.Name, err)
		}
	}

	// 5. Verify
	if s.Final == nil {
		t.Fatalf("expected Final value, got nil")
	}
	if s.Final.Kind != runtime.KindText {
		t.Errorf("expected KindText, got %s", s.Final.Kind)
	}
	resText, _ := ts.Get(s.Final.V.(runtime.TextHandle))
	// "12345" at pos 19. Radius 5 -> [14, 24] -> "e is 12345"
	expected := "e is 12345"
	if resText != expected {
		t.Errorf("expected %q, got %q", expected, resText)
	}
}
