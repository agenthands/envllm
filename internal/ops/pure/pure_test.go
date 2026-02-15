package pure

import (
	"fmt"
	"testing"

	"github.com/agenthands/envllm/internal/runtime"
)

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

func TestPureOps(t *testing.T) {
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{}, ts)
	
	h := ts.Add("hello world\nline 2")
	val := runtime.Value{Kind: runtime.KindText, V: h}

	// Test Stats
	res, _ := Stats(s, val)
	m := res.V.(map[string]interface{})
	if m["lines"] != 2 {
		t.Errorf("expected 2 lines, got %v", m["lines"])
	}

	// Test FindText
	needle := ts.Add("world")
	nval := runtime.Value{Kind: runtime.KindText, V: needle}
	res, _ = FindText(s, val, nval, "FIRST", false)
	if res.V.(int) != 6 {
		t.Errorf("expected pos 6, got %v", res.V)
	}

	// Test JSON
	hj := ts.Add(`{"a": {"b": 1}}`)
	vj := runtime.Value{Kind: runtime.KindText, V: hj}
	pj, _ := JSONParse(s, vj)
	gj, _ := JSONGet(s, pj, "a.b")
	if gj.V.(float64) != 1 {
		t.Errorf("expected 1, got %v", gj.V)
	}

	// Test FindText LAST + IgnoreCase
	needle2 := ts.Add("HELLO")
	nval2 := runtime.Value{Kind: runtime.KindText, V: needle2}
	res, _ = FindText(s, val, nval2, "FIRST", true)
	if res.V.(int) != 0 {
		t.Errorf("expected pos 0, got %v", res.V)
	}
	
	res, _ = FindText(s, val, nval, "LAST", false)
	if res.V.(int) != 6 {
		t.Errorf("expected pos 6 for LAST, got %v", res.V)
	}

	// Test WindowText
	res, _ = WindowText(s, val, 5, 5)
	if res.Kind != runtime.KindText {
		t.Errorf("expected KindText")
	}

	// Test FindRegex
	pat := ts.Add(`[0-9]+`)
	pval := runtime.Value{Kind: runtime.KindText, V: pat}
	res, _ = FindRegex(s, val, pval, "FIRST")
	span := res.V.(runtime.Span)
	if span.Start != 17 { // "line 2" -> "2" is at end. 
		// "hello world\nline 2"
		// 012345678901 2345678
		// Indices: 12=\n, 13=l, 14=i, 15=n, 16=e, 17=' ', 18=2
		// Wait. 18 is '2'.
	}
	if span.Start != 17 {
		t.Errorf("expected span.start 17, got %d", span.Start)
	}

	// Test FindRegex LAST
	res, _ = FindRegex(s, val, pval, "LAST")
	span = res.V.(runtime.Span)
	if span.Start != 17 {
		t.Errorf("expected span.start 17 for LAST, got %d", span.Start)
	}

	// Test FindRegex No Match
	nh := ts.Add("nomatch")
	npat := runtime.Value{Kind: runtime.KindText, V: nh}
	res, _ = FindRegex(s, val, npat, "FIRST")
	span = res.V.(runtime.Span)
	if span.Start != -1 {
		t.Errorf("expected span.start -1, got %d", span.Start)
	}
}

func TestPureOps_Errors(t *testing.T) {
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(runtime.Policy{}, ts)
	
	// JSONParse error
	h := ts.Add("{ invalid")
	val := runtime.Value{Kind: runtime.KindText, V: h}
	_, err := JSONParse(s, val)
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}

	// JSONGet error (not a map)
	pj := runtime.Value{Kind: runtime.KindJSON, V: 123}
	_, err = JSONGet(s, pj, "path")
	if err == nil {
		t.Errorf("expected error for non-map JSONGet")
	}

	// JSONGet error (key not found)
	pj = runtime.Value{Kind: runtime.KindJSON, V: map[string]interface{}{"a": 1}}
	_, err = JSONGet(s, pj, "b")
	if err == nil {
		t.Errorf("expected error for missing key in JSONGet")
	}

	// FindRegex error (invalid pattern)
	ih := ts.Add("[")
	ival := runtime.Value{Kind: runtime.KindText, V: ih}
	_, err = FindRegex(s, val, ival, "FIRST")
	if err == nil {
		t.Errorf("expected error for invalid regex pattern")
	}
}
