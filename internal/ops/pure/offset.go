package pure

import (
	"github.com/agenthands/envllm/internal/runtime"
)

// OffsetAdd implements the OFFSET_ADD operation.
func OffsetAdd(s *runtime.Session, offset runtime.Value, amount int) (runtime.Value, error) {
	val := offset.V.(int)
	return runtime.Value{Kind: runtime.KindOffset, V: val + amount}, nil
}
