package pure

import (
	"strings"

	"github.com/agenthands/rlm-go/internal/runtime"
)

// Stats implements the STATS operation.
func Stats(s *runtime.Session, source runtime.Value) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	text, _ := s.Stores.Text.Get(h)

	res := map[string]interface{}{
		"bytes": len(text),
		"lines": len(strings.Split(text, "\n")),
	}
	return runtime.Value{Kind: runtime.KindJSON, V: res}, nil
}

// FindText implements the FIND_TEXT operation.
func FindText(s *runtime.Session, source runtime.Value, needle runtime.Value, mode string, ignoreCase bool) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	text, _ := s.Stores.Text.Get(h)
	n := needle.V.(runtime.TextHandle)
	ntext, _ := s.Stores.Text.Get(n)

	searchText := text
	searchNeedle := ntext
	if ignoreCase {
		searchText = strings.ToLower(text)
		searchNeedle = strings.ToLower(ntext)
	}

	pos := -1
	if mode == "FIRST" {
		pos = strings.Index(searchText, searchNeedle)
	} else if mode == "LAST" {
		pos = strings.LastIndex(searchText, searchNeedle)
	}

	return runtime.Value{Kind: runtime.KindInt, V: pos}, nil
}

// WindowText implements the WINDOW_TEXT operation.
func WindowText(s *runtime.Session, source runtime.Value, center int, radius int) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	
	wh, err := s.Stores.Text.Window(h, center, radius)
	
	if err != nil {
		return runtime.Value{}, err
	}

	return runtime.Value{Kind: runtime.KindText, V: wh}, nil
}
