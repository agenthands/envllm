package store

import (
	"crypto/sha256"
	"fmt"

	"github.com/agenthands/rlm-go/internal/runtime"
)

// TextStore manages text content and provides handle-based access.
type TextStore struct {
	content map[string]string
}

func NewTextStore() *TextStore {
	return &TextStore{
		content: make(map[string]string),
	}
}

// Add adds text to the store and returns a TextHandle.
func (s *TextStore) Add(text string) runtime.TextHandle {
	id := fmt.Sprintf("t:%x", sha256.Sum256([]byte(text)))
	s.content[id] = text
	return runtime.TextHandle{
		ID:    id,
		Bytes: len(text),
	}
}

// Get returns the text content for a given handle.
func (s *TextStore) Get(h runtime.TextHandle) (string, bool) {
	text, ok := s.content[h.ID]
	return text, ok
}

// Window creates a new snippet based on a center and radius, returning a new handle.
func (s *TextStore) Window(h runtime.TextHandle, center, radius int) (runtime.TextHandle, error) {
	text, ok := s.content[h.ID]
	if !ok {
		return runtime.TextHandle{}, fmt.Errorf("text not found: %s", h.ID)
	}

	start := center - radius
	end := center + radius

	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start > end {
		start = end
	}

	snippet := text[start:end]
	return s.Add(snippet), nil
}

// Slice creates a new snippet based on start and end indices.
func (s *TextStore) Slice(h runtime.TextHandle, start, end int) (runtime.TextHandle, error) {
	text, ok := s.content[h.ID]
	if !ok {
		return runtime.TextHandle{}, fmt.Errorf("text not found: %s", h.ID)
	}

	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start > end {
		return s.Add(""), nil
	}

	snippet := text[start:end]
	return s.Add(snippet), nil
}
