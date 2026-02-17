package rewrite

import (
	"context"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
)

type MissingTypesRule struct {
	table *ops.Table
}

func NewMissingTypesRule(table *ops.Table) *MissingTypesRule {
	return &MissingTypesRule{table: table}
}

func (r *MissingTypesRule) ID() RuleID {
	return "RULE_MISSING_TYPES"
}

func (r *MissingTypesRule) Description() string {
	return "Add mandatory type annotations to INTO clauses based on operation definitions."
}

func (r *MissingTypesRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	var matches []MatchResult

	visitor := &typeVisitor{
		table: r.table,
		onMatch: func(node ast.Node, expectedType string) {
			matches = append(matches, MatchResult{Node: node, Data: expectedType})
		},
	}
	ast.Walk(visitor, prog)

	return matches, len(matches) > 0
}

type typeVisitor struct {
	table   *ops.Table
	onMatch func(node ast.Node, expectedType string)
}

func (v *typeVisitor) Visit(node ast.Node) ast.Visitor {
	if op, ok := node.(*ast.OpStmt); ok {
		if op.Into != "" && op.IntoType == "" {
			if def, found := v.table.Ops[op.OpName]; found && def.ResultType != "" {
				v.onMatch(op, string(def.ResultType))
			}
		}
	}
	return v
}

func (r *MissingTypesRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
	for _, m := range matches {
		if op, ok := m.Node.(*ast.OpStmt); ok {
			op.IntoType = m.Data.(string)
		}
	}
	return prog, nil
}
