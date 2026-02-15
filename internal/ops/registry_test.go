package ops

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/runtime"
)

type mockHost struct {
	subcallFunc func(req runtime.SubcallRequest) (runtime.SubcallResponse, error)
}

func (m *mockHost) Subcall(ctx context.Context, req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
	return m.subcallFunc(req)
}

type mockTextStore struct {
	content map[string]string
	nextID  int
}

func (m *mockTextStore) Add(text string) runtime.TextHandle {
	m.nextID++
	id := fmt.Sprintf("t%d", m.nextID)
	m.content[id] = text
	return runtime.TextHandle{ID: id, Bytes: len(text)}
}
func (m *mockTextStore) Get(h runtime.TextHandle) (string, bool) {
	t, ok := m.content[h.ID]
	return t, ok
}
func (m *mockTextStore) Window(h runtime.TextHandle, center, radius int) (runtime.TextHandle, error) {
	return runtime.TextHandle{ID: "w1", Bytes: 10}, nil
}
func (m *mockTextStore) Slice(h runtime.TextHandle, start, end int) (runtime.TextHandle, error) {
	return runtime.TextHandle{ID: "s1", Bytes: 10}, nil
}

func exprToKwArg(kw string, expr ast.Expr) ast.KwArg {
	return ast.KwArg{Keyword: kw, Value: expr}
}

func TestSubcall(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	reg := NewRegistry(tbl)
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{
		MaxSubcalls: 2, 
		MaxRecursionDepth: 10,
		AllowedCapabilities: map[string]bool{"llm": true},
	}, ts)
	
	host := &mockHost{
		subcallFunc: func(req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
			return runtime.SubcallResponse{
				Result: runtime.Value{Kind: runtime.KindJSON, V: map[string]interface{}{"res": 100}},
			}, nil
		},
	}
	s.Host = host

	taskHandle := ts.Add("test")
	s.Env.Define("task_var", runtime.Value{Kind: runtime.KindText, V: taskHandle})
	s.Env.Define("prompt_var", runtime.Value{Kind: runtime.KindText, V: runtime.TextHandle{ID: "t1"}})

	args := []ast.KwArg{
		exprToKwArg("SOURCE", &ast.IdentExpr{Name: "prompt_var"}),
		exprToKwArg("TASK", &ast.IdentExpr{Name: "task_var"}),
		exprToKwArg("DEPTH_COST", &ast.IntExpr{Value: 1}),
	}

	res, err := reg.Dispatch(s, "SUBCALL", args)
	if err != nil {
		t.Fatalf("Dispatch SUBCALL failed: %v", err)
	}

	if res.Kind != runtime.KindJSON {
		t.Errorf("expected KindJSON, got %v", res.Kind)
	}

	if s.SubcallCount != 1 {
		t.Errorf("expected SubcallCount 1, got %d", s.SubcallCount)
	}
	if s.RecursionDepth != 1 {
		t.Errorf("expected RecursionDepth 1, got %d", s.RecursionDepth)
	}
}

func TestSubcall_BudgetExceeded(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	reg := NewRegistry(tbl)
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{
		MaxSubcalls: 1, 
		MaxRecursionDepth: 5,
		AllowedCapabilities: map[string]bool{"llm": true},
	}, ts)
	s.Host = &mockHost{}

	taskHandle := ts.Add("test")
	s.Env.Define("task_var", runtime.Value{Kind: runtime.KindText, V: taskHandle})
	s.Env.Define("prompt_var", runtime.Value{Kind: runtime.KindText, V: runtime.TextHandle{ID: "t1"}})

	args := []ast.KwArg{
		exprToKwArg("SOURCE", &ast.IdentExpr{Name: "prompt_var"}),
		exprToKwArg("TASK", &ast.IdentExpr{Name: "task_var"}),
		exprToKwArg("DEPTH_COST", &ast.IntExpr{Value: 1}),
	}

	// First call ok
	s.Host.(*mockHost).subcallFunc = func(req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
		return runtime.SubcallResponse{Result: runtime.Value{Kind: runtime.KindJSON, V: map[string]interface{}{}}}, nil
	}
	_, err := reg.Dispatch(s, "SUBCALL", args)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should fail (MaxSubcalls = 1)
	_, err = reg.Dispatch(s, "SUBCALL", args)
	if err == nil {
		t.Errorf("expected error for MaxSubcalls exceeded")
	}

	// Test RecursionDepth exceeded
	s2 := runtime.NewSession(runtime.Policy{
		MaxSubcalls: 10, 
		MaxRecursionDepth: 5,
		AllowedCapabilities: map[string]bool{"llm": true},
	}, ts)
	s2.Host = &mockHost{}
	s2.Env.Define("task_var", runtime.Value{Kind: runtime.KindText, V: taskHandle})
	s2.Env.Define("prompt_var", runtime.Value{Kind: runtime.KindText, V: runtime.TextHandle{ID: "t1"}})

	args2 := []ast.KwArg{
		exprToKwArg("SOURCE", &ast.IdentExpr{Name: "prompt_var"}),
		exprToKwArg("TASK", &ast.IdentExpr{Name: "task_var"}),
		exprToKwArg("DEPTH_COST", &ast.IntExpr{Value: 6}),
	}
	_, err = reg.Dispatch(s2, "SUBCALL", args2)
	if err == nil {
		t.Errorf("expected error for MaxRecursionDepth exceeded")
	}
}

func TestCapabilityGating(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	reg := NewRegistry(tbl)
	ts := &mockTextStore{content: make(map[string]string)}
	
	// Policy without 'llm' capability
	s := runtime.NewSession(runtime.Policy{
		AllowedCapabilities: map[string]bool{},
	}, ts)
	s.Host = &mockHost{}

	taskHandle := ts.Add("test")
	s.Env.Define("task_var", runtime.Value{Kind: runtime.KindText, V: taskHandle})
	s.Env.Define("prompt_var", runtime.Value{Kind: runtime.KindText, V: runtime.TextHandle{ID: "t1"}})

	args := []ast.KwArg{
		exprToKwArg("SOURCE", &ast.IdentExpr{Name: "prompt_var"}),
		exprToKwArg("TASK", &ast.IdentExpr{Name: "task_var"}),
		exprToKwArg("DEPTH_COST", &ast.IntExpr{Value: 1}),
	}

	_, err := reg.Dispatch(s, "SUBCALL", args)
	if err == nil || !strings.Contains(err.Error(), "denied by policy") {
		t.Errorf("expected capability denied error, got %v", err)
	}

	// Policy with 'llm' capability
	s.Policy.AllowedCapabilities["llm"] = true
	s.Host.(*mockHost).subcallFunc = func(req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
		return runtime.SubcallResponse{Result: runtime.Value{Kind: runtime.KindJSON, V: map[string]interface{}{}}}, nil
	}
	_, err = reg.Dispatch(s, "SUBCALL", args)
	if err != nil {
		t.Errorf("expected success with 'llm' capability, got %v", err)
	}
}

func TestTable_ValidateSignature_Errors(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	
	// Unknown op
	_, err := tbl.ValidateSignature("UNKNOWN", nil)
	if err == nil {
		t.Errorf("expected error for unknown op")
	}

	// Wrong arg count
	_, err = tbl.ValidateSignature("STATS", []ValidatedKwArg{})
	if err == nil {
		t.Errorf("expected error for wrong arg count")
	}
}

func TestRegistry_Dispatch_Errors(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	reg := NewRegistry(tbl)
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{}, ts)

	// Unknown op
	_, err := reg.Dispatch(s, "UNKNOWN", nil)
	if err == nil {
		t.Errorf("expected error for unknown op")
	}

	// Missing impl
	delete(reg.impls, "STATS")
	h := ts.Add("test")
	s.Env.Define("h", runtime.Value{Kind: runtime.KindText, V: h})
	args := []ast.KwArg{exprToKwArg("SOURCE", &ast.IdentExpr{Name: "h"})}
	_, err = reg.Dispatch(s, "STATS", args)
	if err == nil {
		t.Errorf("expected error for missing implementation")
	}

	// Test Result type mismatch
	reg.registerDefaults() // restore STATS
	reg.impls["STATS"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return runtime.Value{Kind: runtime.KindInt, V: 1}, nil
	}
	_, err = reg.Dispatch(s, "STATS", args)
	if err == nil || !strings.Contains(err.Error(), "result type mismatch") {
		t.Errorf("expected result type mismatch error, got %v", err)
	}
}

func TestRegistry_Dispatch_AllPure(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")
	reg := NewRegistry(tbl)
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{}, ts)
	
	h := ts.Add("test")
	s.Env.Define("h", runtime.Value{Kind: runtime.KindText, V: h})
	
	// STATS
	_, err := reg.Dispatch(s, "STATS", []ast.KwArg{exprToKwArg("SOURCE", &ast.IdentExpr{Name: "h"})})
	if err != nil {
		t.Errorf("STATS failed: %v", err)
	}

	// WINDOW_TEXT
	_, err = reg.Dispatch(s, "WINDOW_TEXT", []ast.KwArg{
		exprToKwArg("SOURCE", &ast.IdentExpr{Name: "h"}),
		exprToKwArg("CENTER", &ast.IntExpr{Value: 0}),
		exprToKwArg("RADIUS", &ast.IntExpr{Value: 0}),
	})
	if err != nil {
		t.Errorf("WINDOW_TEXT failed: %v", err)
	}

	// JSON_PARSE
	hj := ts.Add(`{"a":1}`)
	s.Env.Define("hj", runtime.Value{Kind: runtime.KindText, V: hj})
	_, err = reg.Dispatch(s, "JSON_PARSE", []ast.KwArg{exprToKwArg("SOURCE", &ast.IdentExpr{Name: "hj"})})
	if err != nil {
		t.Errorf("JSON_PARSE failed: %v", err)
	}
}

func TestLoadTable_Error(t *testing.T) {
	_, err := LoadTable("non-existent.json")
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}

	// Invalid JSON
	path := "../../assets/dialect_card.md" // Not a JSON
	_, err = LoadTable(path)
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}
