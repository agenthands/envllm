package rewrite

import (
	"context"
	"github.com/agenthands/envllm/internal/ast"
)

// Engine handles the application of multiple rewrite rules.
type Engine struct {
	registry *Registry
}

func NewEngine(registry *Registry) *Engine {
	return &Engine{registry: registry}
}

// ApplyRules applies a set of rules identified by their IDs.
func (e *Engine) ApplyRules(ctx context.Context, prog *ast.Program, ruleIDs []string) (*ast.Program, error) {
	currentProg := prog
	for _, id := range ruleIDs {
		rule, ok := e.registry.Find(RuleID(id))
		if !ok {
			continue // Skip unknown rules
		}
		
		matches, ok := rule.Match(ctx, currentProg, nil)
		if ok {
			var err error
			currentProg, err = rule.Apply(ctx, currentProg, matches)
			if err != nil {
				return nil, err
			}
		}
	}
	return currentProg, nil
}

// AutoRepair attempts to match and apply any applicable rules until no more matches are found (or max iterations).
func (e *Engine) AutoRepair(ctx context.Context, prog *ast.Program) (*ast.Program, []string, error) {
	currentProg := prog
	appliedIDs := []string{}
	
	for i := 0; i < 10; i++ { // Limit iterations to prevent loops
		matched := false
		for _, rule := range e.registry.List() {
			matches, ok := rule.Match(ctx, currentProg, nil)
			if ok {
				var err error
				currentProg, err = rule.Apply(ctx, currentProg, matches)
				if err != nil {
					return nil, nil, err
				}
				appliedIDs = append(appliedIDs, string(rule.ID()))
				matched = true
				break // Restart loop to check for new matches in modified AST
			}
		}
		if !matched {
			break
		}
	}
	
	return currentProg, appliedIDs, nil
}
