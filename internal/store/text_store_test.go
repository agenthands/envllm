package store

import (
	"testing"
)

func TestTextStore(t *testing.T) {
	s := NewTextStore()
	h := s.Add("Hello World")

	if h.Bytes != 11 {
		t.Errorf("expected 11 bytes, got %d", h.Bytes)
	}

	text, ok := s.Get(h)
	if !ok || text != "Hello World" {
		t.Errorf("Get failed")
	}

	// Test Window
	wh, err := s.Window(h, 6, 2)
	if err != nil {
		t.Fatalf("Window failed: %v", err)
	}
	wtext, _ := s.Get(wh)
	if wtext != " Wo" { // center 6 (space), radius 2 -> 4 to 8 -> "o Wo" Wait. 
		// "Hello World"
		// 01234567890
		// center 6 is 'W'. radius 2 -> 4 to 8. 
		// text[4:8] is "o Wo"
		// indices: 4='o', 5=' ', 6='W', 7='o'
	}
	// Let's re-verify my math.
	// H e l l o   W o r l d
	// 0 1 2 3 4 5 6 7 8 9 0
	// center=6 ('W'), radius=2.
	// start = 6-2=4 ('o')
	// end = 6+2=8 ('r')
	// text[4:8] -> indices 4,5,6,7 -> "o Wo"
	if wtext != "o Wo" {
		t.Errorf("expected 'o Wo', got %q", wtext)
	}

	// Test Slice
	sh, err := s.Slice(h, 0, 5)
	if err != nil {
		t.Fatalf("Slice failed: %v", err)
	}
	stext, _ := s.Get(sh)
	if stext != "Hello" {
		t.Errorf("expected 'Hello', got %q", stext)
	}

	// Test boundary cases
	bh, _ := s.Window(h, 0, 100)
	btext, _ := s.Get(bh)
	if btext != "Hello World" {
		t.Errorf("expected full text, got %q", btext)
	}

	bh2, _ := s.Slice(h, -10, 100)
	btext2, _ := s.Get(bh2)
	if btext2 != "Hello World" {
		t.Errorf("expected full text, got %q", btext2)
	}

	bh3, _ := s.Slice(h, 10, 5)
	btext3, _ := s.Get(bh3)
	if btext3 != "" {
		t.Errorf("expected empty text, got %q", btext3)
	}
}

func TestTextStore_Errors(t *testing.T) {
	s := NewTextStore()
	h := s.Add("test")
	delete(s.content, h.ID)

	if _, err := s.Window(h, 0, 0); err == nil {
		t.Errorf("expected error for missing text in Window")
	}
	if _, err := s.Slice(h, 0, 0); err == nil {
		t.Errorf("expected error for missing text in Slice")
	}
}
