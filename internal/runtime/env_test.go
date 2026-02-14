package runtime

import (
	"testing"
)

func TestEnv(t *testing.T) {
	e := NewEnv()
	
	val := Value{Kind: KindInt, V: 42}
	if err := e.Define("x", val); err != nil {
		t.Fatalf("Define failed: %v", err)
	}

	got, ok := e.Get("x")
	if !ok || got.V != 42 {
		t.Errorf("Get failed")
	}

	// Test single-assignment
	if err := e.Define("x", val); err == nil {
		t.Errorf("expected error for re-definition of 'x'")
	}

	// Test non-existent var
	if _, ok := e.Get("y"); ok {
		t.Errorf("expected ok=false for non-existent 'y'")
	}
}
