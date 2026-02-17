package rewrite

import (
	"context"
	"github.com/agenthands/envllm/internal/ast"
)

// RuleID is a unique identifier for a rewrite rule.
type RuleID string

// MatchResult contains information about where a rule matched.
type MatchResult struct {
	Node ast.Node
	Data interface{} // Optional rule-specific metadata
}

// RewriteRule defines the interface for an AST transformation.
type RewriteRule interface {
	ID() RuleID
	Description() string
	
	// Match checks if the rule applies to the given program or a specific error.
	// It returns matched nodes and true if applicable.
	Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool)
	
	// Apply transforms the program based on the match results.
	Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error)
}

// Registry stores and manages rewrite rules.
type Registry struct {
	rules map[RuleID]RewriteRule
}

func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[RuleID]RewriteRule),
	}
}

func (r *Registry) Register(rule RewriteRule) {
	r.rules[rule.ID()] = rule
}

func (r *Registry) Find(id RuleID) (RewriteRule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

func (r *Registry) List() []RewriteRule {
	var list []RewriteRule
	for _, rule := range r.rules {
		list = append(list, rule)
	}
	return list
}
