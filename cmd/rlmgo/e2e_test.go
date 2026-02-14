package main

import (
	"context"
	"path/filepath"
	"strings"
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

func TestEndToEnd_Recursive(t *testing.T) {
	tbl, _ := ops.LoadTable("../../assets/ops.json")
	reg := ops.NewRegistry(tbl)
	ts := store.NewTextStore()
	policy := runtime.Policy{
		MaxSubcalls:       5,
		MaxRecursionDepth: 10,
		AllowedCapabilities: map[string]bool{"llm": true},
	}
	s := runtime.NewSession(policy, ts)
	s.Dispatcher = reg

	mockResult := map[string]interface{}{"found": true}
	s.Host = &mockHost{
		subcallFunc: func(req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
			return runtime.SubcallResponse{
				Result: runtime.Value{Kind: runtime.KindJSON, V: mockResult},
			}, nil
		},
	}

	code := `RLMDSL 0.1
CELL recurse:
  SUBCALL SOURCE "some text" TASK "extract data" DEPTH_COST 2 INTO out
  SET_FINAL SOURCE out
`
	l := lex.NewLexer("recurse.rlm", code)
	p := parse.NewParser(l)
	prog, _ := p.Parse()

	for _, cell := range prog.Cells {
		_ = s.ExecuteCell(context.Background(), cell)
	}

	if s.SubcallCount != 1 {
		t.Errorf("expected 1 subcall")
	}
	if s.RecursionDepth != 2 {
		t.Errorf("expected depth 2")
	}
}

func TestEndToEnd_RegexAndCapabilities(t *testing.T) {
	tbl, _ := ops.LoadTable("../../assets/ops.json")
	reg := ops.NewRegistry(tbl)
	ts := store.NewTextStore()
	
	// Policy WITHOUT 'llm' capability
	policy := runtime.Policy{
		MaxStmtsPerCell: 100,
		AllowedCapabilities: map[string]bool{
			"pure": true, // Not strictly needed but for clarity
		},
	}
	s := runtime.NewSession(policy, ts)
	s.Dispatcher = reg

	prompt := "My email is test@example.com."
	ph := ts.Add(prompt)
	s.Env.Define("PROMPT", runtime.Value{Kind: runtime.KindText, V: ph})

	// 1. Test FIND_REGEX (pure)
	code := `RLMDSL 0.1
CELL regex:
  FIND_REGEX SOURCE PROMPT PATTERN "[a-z]+@[a-z.]+" MODE "FIRST" INTO email_span
  SET_FINAL SOURCE email_span
`
	l := lex.NewLexer("regex.rlm", code)
	p := parse.NewParser(l)
	prog, _ := p.Parse()

	for _, cell := range prog.Cells {
		_ = s.ExecuteCell(context.Background(), cell)
	}

	if s.Final == nil || s.Final.Kind != runtime.KindSpan {
		t.Fatalf("expected span result, got %v", s.Final)
	}
	span := s.Final.V.(runtime.Span)
	if span.Start != 12 { // "My email is " is 12 chars.
		t.Errorf("expected start 12, got %d", span.Start)
	}

	// 2. Test Capability Denial (SUBCALL needs 'llm')
	code2 := `RLMDSL 0.1
CELL denied:
  SUBCALL SOURCE PROMPT TASK "summarize" DEPTH_COST 1 INTO out
`
	l2 := lex.NewLexer("denied.rlm", code2)
	p2 := parse.NewParser(l2)
	prog2, _ := p2.Parse()

	err := s.ExecuteCell(context.Background(), prog2.Cells[0])
	if err == nil || !strings.Contains(err.Error(), "denied by policy") {
		t.Errorf("expected capability denied error, got %v", err)
	}
}

func TestEndToEnd_FileSystem(t *testing.T) {
	tmpDir := t.TempDir()
	tbl, _ := ops.LoadTable("../../assets/ops.json")
	reg := ops.NewRegistry(tbl)
	ts := store.NewTextStore()
	
	policy := runtime.Policy{
		MaxStmtsPerCell: 100,
		AllowedCapabilities: map[string]bool{
			"fs_read":  true,
			"fs_write": true,
		},
		AllowedReadPaths:  []string{tmpDir},
		AllowedWritePaths: []string{tmpDir},
	}
	s := runtime.NewSession(policy, ts)
	s.Dispatcher = reg

	filePath := filepath.Join(tmpDir, "e2e.txt")
	ph := ts.Add(filePath)
	s.Env.Define("FILE_PATH", runtime.Value{Kind: runtime.KindText, V: ph})

	code := `RLMDSL 0.1
CELL fs:
  WRITE_FILE PATH FILE_PATH SOURCE "e2e content" INTO ok
  READ_FILE PATH FILE_PATH INTO content
  SET_FINAL SOURCE content
`
	l := lex.NewLexer("fs.rlm", code)
	p := parse.NewParser(l)
	prog, _ := p.Parse()

	for _, cell := range prog.Cells {
		_ = s.ExecuteCell(context.Background(), cell)
	}

	if s.Final == nil || s.Final.Kind != runtime.KindText {
		t.Fatalf("expected text result, got %v", s.Final)
	}
	resText, _ := ts.Get(s.Final.V.(runtime.TextHandle))
	if resText != "e2e content" {
		t.Errorf("expected 'e2e content', got %q", resText)
	}
}

type mockHost struct {
	subcallFunc func(req runtime.SubcallRequest) (runtime.SubcallResponse, error)
}

func (m *mockHost) Subcall(ctx context.Context, req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
	return m.subcallFunc(req)
}
