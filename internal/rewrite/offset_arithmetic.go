package rewrite

import (
	"context"
	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
)

type OffsetArithmeticRule struct {
	table *ops.Table
}

func NewOffsetArithmeticRule(table *ops.Table) *OffsetArithmeticRule {
	return &OffsetArithmeticRule{table: table}
}

func (r *OffsetArithmeticRule) ID() RuleID {
	return "RULE_OFFSET_ARITHMETIC"
}

func (r *OffsetArithmeticRule) Description() string {
	return "Replace FIND_TEXT + OFFSET_ADD(len) with a single AFTER_TEXT operation."
}

func (r *OffsetArithmeticRule) Match(ctx context.Context, prog *ast.Program, err interface{}) ([]MatchResult, bool) {
	if prog.Task == nil {
		return nil, false
	}

	var matches []MatchResult

	// Track FIND_TEXT outputs and their needles
	offsets := make(map[string]*ast.OpStmt) // var -> find_text_op

	for _, item := range prog.Task.Body {
		cell, ok := item.(*ast.Cell)
		if !ok {
			continue
		}
		for _, stmt := range cell.Stmts {
			op, ok := stmt.(*ast.OpStmt)
			if !ok {
				continue
			}

			if op.OpName == "FIND_TEXT" && op.Into != "" {
				offsets[op.Into] = op
			}

			if op.OpName == "OFFSET_ADD" {
				var targetOffset string
				var amount int
				var hasAmount bool

				for _, arg := range op.Args {
					if arg.Keyword == "OFFSET" {
						if id, ok := arg.Value.(*ast.IdentExpr); ok {
							targetOffset = id.Name
						}
					}
					if arg.Keyword == "AMOUNT" {
						if lit, ok := arg.Value.(*ast.IntExpr); ok {
							amount = lit.Value
							hasAmount = true
						}
					}
				}

				if hasAmount && targetOffset != "" {
					var findMatch *ast.OpStmt
					if findOp, exists := offsets[targetOffset]; exists {
						// Check if amount matches needle length
						for _, arg := range findOp.Args {
							if arg.Keyword == "NEEDLE" {
								if s, ok := arg.Value.(*ast.StringExpr); ok {
									if len(s.Value) == amount {
										findMatch = findOp
									}
								}
							}
						}
					}
					matches = append(matches, MatchResult{
						Node: op,
						Data: findMatch,
					})
				}
			}
		}
	}

	return matches, len(matches) > 0
}

func (r *OffsetArithmeticRule) Apply(ctx context.Context, prog *ast.Program, matches []MatchResult) (*ast.Program, error) {
	for _, m := range matches {
		op, ok := m.Node.(*ast.OpStmt)
		if !ok || op == nil {
			continue
		}
		
		if m.Data != nil {
			findOp, ok := m.Data.(*ast.OpStmt)
			if ok && findOp != nil {
				// High-fidelity fix: Convert OFFSET_ADD to AFTER_TEXT
				op.OpName = "AFTER_TEXT"
				op.Args = findOp.Args
			}
		}
	}
	return prog, nil
}
