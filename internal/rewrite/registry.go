package rewrite

import (
	"github.com/agenthands/envllm/internal/ops"
)

// DefaultRegistry returns a registry populated with all core rewrite rules.
func DefaultRegistry(table *ops.Table) *Registry {
	r := NewRegistry()
	r.Register(NewMissingRequiresRule(table))
	r.Register(NewDotAccessRule())
	r.Register(NewNumericConcatRule())
	r.Register(NewMissingTypesRule(table))
	r.Register(&MissingOutputRule{})
	r.Register(&VarRenameRule{})
	r.Register(NewOffsetArithmeticRule(table))
	r.Register(NewUndefinedToLiteralRule(table))
	return r
}
