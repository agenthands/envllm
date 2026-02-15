package pure

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agenthands/envllm/internal/runtime"
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

// SliceText implements the SLICE_TEXT operation.
func SliceText(s *runtime.Session, source runtime.Value, start, end int) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	
	wh, err := s.Stores.Text.Slice(h, start, end)
	if err != nil {
		return runtime.Value{}, err
	}

	return runtime.Value{Kind: runtime.KindText, V: wh}, nil
}

// FindRegex implements the FIND_REGEX operation.
func FindRegex(s *runtime.Session, source runtime.Value, pattern runtime.Value, mode string) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	text, _ := s.Stores.Text.Get(h)
	
	ph := pattern.V.(runtime.TextHandle)
	pat, _ := s.Stores.Text.Get(ph)

	re, err := regexp.Compile(pat)
	if err != nil {
		return runtime.Value{}, fmt.Errorf("FIND_REGEX invalid pattern %q: %v", pat, err)
	}

	indices := re.FindAllStringIndex(text, -1)
	if len(indices) == 0 {
		return runtime.Value{Kind: runtime.KindSpan, V: runtime.Span{Start: -1, End: -1}}, nil
	}

	var match []int
	if mode == "FIRST" {
		match = indices[0]
	} else if mode == "LAST" {
		match = indices[len(indices)-1]
	}

	return runtime.Value{Kind: runtime.KindSpan, V: runtime.Span{Start: match[0], End: match[1]}}, nil
}

// GetSpanStart implements the GET_SPAN_START operation.
func GetSpanStart(s *runtime.Session, source runtime.Value) (runtime.Value, error) {
	span := source.V.(runtime.Span)
	return runtime.Value{Kind: runtime.KindInt, V: span.Start}, nil
}

// GetSpanEnd implements the GET_SPAN_END operation.
func GetSpanEnd(s *runtime.Session, source runtime.Value) (runtime.Value, error) {
	span := source.V.(runtime.Span)
	return runtime.Value{Kind: runtime.KindInt, V: span.End}, nil
}
