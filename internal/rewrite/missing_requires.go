package rewrite

import (
	"context"
	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
)

type MissingRequiresRule struct {
	table *ops.Table
}

func NewMissingRequiresRule(table *ops.Table) *MissingRequiresRule {
	return &MissingRequiresRule{table: table}
}

func (r *MissingRequiresRule) ID() RuleID {
	return "RULE_MISSING_REQUIRES"
}

func (r *MissingRequiresRule) Description() string {
	return "Add missing REQUIRES declarations for operations used in the program."
}

type opVisitor struct {
	usedOps map[string]bool
}

func (v *opVisitor) Visit(node ast.Node) ast.Visitor {
	if op, ok := node.(*ast.OpStmt); ok {
		v.usedOps[op.OpName] = true
	}
	return v
}

func (r *MissingRequiresRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	if prog.Task == nil {
		return nil, false
	}

	// 1. Find all used ops
	v := &opVisitor{usedOps: make(map[string]bool)}
	ast.Walk(v, prog)

	// 2. Find declared capabilities
	declared := make(map[string]bool)
	for _, item := range prog.Task.Body {
		if req, ok := item.(*ast.Requirement); ok {
			declared[req.Capability] = true
		}
	}

	// 3. Check for missing capabilities
	missing := make(map[string]bool)
	for opName := range v.usedOps {
		if opDef, ok := r.table.Ops[opName]; ok {
			for _, cap := range opDef.Capabilities {
				if cap == "pure" {
					continue
				}
				if !declared[cap] {
					missing[cap] = true
				}
			}
		}
	}

	if len(missing) == 0 {
		return nil, false
	}

	// Return the missing capabilities as metadata in a single MatchResult on the Program
	return []MatchResult{{Node: prog, Data: missing}}, true
}

func (r *MissingRequiresRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
	if len(matches) == 0 {
		return prog, nil
	}

	missing := matches[0].Data.(map[string]bool)
	
	// Create new requirements
	var newReqs []ast.BodyItem
	for cap := range missing {
		newReqs = append(newReqs, &ast.Requirement{
			Capability: cap,
		})
	}

	// Insert at the top of Task.Body
	prog.Task.Body = append(newReqs, prog.Task.Body...)

	return prog, nil
}
