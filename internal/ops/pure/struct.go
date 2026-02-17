package pure

import (
	"fmt"

	"github.com/agenthands/envllm/internal/runtime"
)

// GetField implements GET_FIELD for Structs and Spans.
func GetField(s *runtime.Session, source runtime.Value, field string) (runtime.Value, error) {
	if source.Kind == runtime.KindSpan {
		span := source.V.(runtime.Span)
		if field == "start" {
			return runtime.Value{Kind: runtime.KindOffset, V: span.Start}, nil
		}
		if field == "end" {
			return runtime.Value{Kind: runtime.KindOffset, V: span.End}, nil
		}
		return runtime.Value{}, fmt.Errorf("GET_FIELD: unknown field %q for SPAN (use start or end)", field)
	}

	if source.Kind != runtime.KindStruct {
		return runtime.Value{}, fmt.Errorf("GET_FIELD source must be STRUCT or SPAN, got %s", source.Kind)
	}

	m := source.V.(map[string]interface{})
	val, ok := m[field]
	if !ok {
		// Provide available keys hint
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return runtime.Value{}, fmt.Errorf("GET_FIELD failed: field %q not found. Available fields: %v", field, keys)
	}

	// Auto-box common types
	switch v := val.(type) {
	case int:
		return runtime.Value{Kind: runtime.KindJSON, V: v}, nil
	case float64:
		return runtime.Value{Kind: runtime.KindJSON, V: int(v)}, nil
	case string:
		return runtime.Value{Kind: runtime.KindText, V: s.Stores.Text.Add(v)}, nil
	case bool:
		return runtime.Value{Kind: runtime.KindBool, V: v}, nil
	case runtime.Value:
		return v, nil
	default:
		// Fallback for nested structs or unhandled types
		return runtime.Value{Kind: runtime.KindJSON, V: v}, nil
	}
}
