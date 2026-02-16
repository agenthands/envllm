package pure

import (
	"fmt"

	"github.com/agenthands/envllm/internal/runtime"
)

// GetCost implements the GET_COST operation.
// In v0.2, this extracts cost information from a previous result or global stats.
// Currently, it's a stub to satisfy the type system.
func GetCost(s *runtime.Session, source runtime.Value) (runtime.Value, error) {
	// source is expected to be JSON (e.g. from STATS or SUBCALL result)
	// For now, we return a dummy cost.
	// In a real implementation, this would look up "cost" or "budget" fields.
	
	// Check if source is a map
	if m, ok := source.V.(map[string]interface{}); ok {
		if c, ok := m["cost"]; ok {
			// Try float first (JSON unmarshal)
			if f, ok := c.(float64); ok {
				return runtime.Value{Kind: runtime.KindCost, V: int(f)}, nil
			}
			if i, ok := c.(int); ok {
				return runtime.Value{Kind: runtime.KindCost, V: i}, nil
			}
		}
	}
	
	return runtime.Value{Kind: runtime.KindCost, V: 0}, nil
}
