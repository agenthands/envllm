package rewrite

import (
	"context"

	"github.com/agenthands/envllm/internal/ast"
)

type MissingOutputRule struct{}

func (r *MissingOutputRule) ID() RuleID {
	return "RULE_MISSING_OUTPUT"
}

func (r *MissingOutputRule) Description() string {
	return "Ensure the TASK has a valid OUTPUT variable."
}

func (r *MissingOutputRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	if prog.Task == nil || prog.Task.Output != "" {
		return nil, false
	}

	// Find the last variable defined via INTO
	lastVar := ""
	ast.Walk(&outputVisitor{onVar: func(v string) { lastVar = v }}, prog)

	if lastVar == "" {
		return nil, false
	}

	return []MatchResult{{Node: prog.Task, Data: lastVar}}, true
}

type outputVisitor struct {
	onVar func(string)
}

func (v *outputVisitor) Visit(node ast.Node) ast.Visitor {
	if op, ok := node.(*ast.OpStmt); ok && op.Into != "" {
		v.onVar(op.Into)
	}
	return v
}

func (r *MissingOutputRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
	if len(matches) > 0 {
		prog.Task.Output = matches[0].Data.(string)
	}
	return prog, nil
}
