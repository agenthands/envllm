package pure

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agenthands/envllm/internal/runtime"
)

// JSONParse implements the JSON_PARSE operation.
func JSONParse(s *runtime.Session, source runtime.Value) (runtime.Value, error) {
	h := source.V.(runtime.TextHandle)
	text, _ := s.Stores.Text.Get(h)

	var v interface{}
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		return runtime.Value{}, fmt.Errorf("JSON_PARSE failed: %v", err)
	}

	return runtime.Value{Kind: runtime.KindJSON, V: v}, nil
}

// JSONGet implements the JSON_GET operation.
// Simplistic implementation for now: only works for maps.
func JSONGet(s *runtime.Session, source runtime.Value, path string) (runtime.Value, error) {
	data := source.V
	
	// Strip optional $. prefix (common LLM hallucination)
	path = strings.TrimPrefix(path, "$.")
	
	parts := strings.Split(path, ".")
	curr := data
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		m, ok := curr.(map[string]interface{})
		if !ok {
			return runtime.Value{}, fmt.Errorf("JSON_GET failed: current value is not a map")
		}
		val, ok := m[part]
		if !ok {
			// Provide available keys hint
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			return runtime.Value{}, fmt.Errorf("JSON_GET failed: key %q not found. Available keys: %v", part, keys)
		}
		curr = val
	}

	// Wrap result in Value
	return runtime.Value{Kind: runtime.KindJSON, V: curr}, nil
}
