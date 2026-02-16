package lint

import (
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
)

type Linter struct {
	table *ops.Table
}

func NewLinter(table *ops.Table) *Linter {
	return &Linter{table: table}
}

func (l *Linter) Lint(prog *ast.Program) []Error {
	var errs []Error
	symbols := make(map[string]string) // name -> type

	for _, cell := range prog.Cells {
		for _, stmt := range cell.Stmts {
			switch s := stmt.(type) {
			case *ast.OpStmt:
				opErrs, outType := l.lintOpStmt(s, symbols)
				errs = append(errs, opErrs...)
				if s.Into != "" && outType != "" {
					symbols[s.Into] = outType
				}
			case *ast.SetFinalStmt:
				errs = append(errs, l.lintExpr(s.Source, "", symbols)...)
			case *ast.PrintStmt:
				errs = append(errs, l.lintExpr(s.Source, "", symbols)...)
			case *ast.AssertStmt:
				errs = append(errs, l.lintExpr(s.Cond, "BOOL", symbols)...)
			}
		}
	}

	return errs
}

func (l *Linter) lintOpStmt(s *ast.OpStmt, symbols map[string]string) ([]Error, string) {
	var errs []Error

	opDef, ok := l.table.Ops[s.OpName]
	if !ok {
		errs = append(errs, Error{
			Code:    "LINT_UNKNOWN_OP",
			Message: fmt.Sprintf("unknown operation: %s", s.OpName),
			Loc:     s.Loc,
		})
		return errs, ""
	}

	// 1. Enforce clause order and type check arguments
	if len(s.Args) != len(opDef.Signature) {
		errs = append(errs, Error{
			Code:    "LINT_ARG_COUNT",
			Message: fmt.Sprintf("%s: expected %d arguments, got %d", s.OpName, len(opDef.Signature), len(s.Args)),
			Loc:     s.Loc,
		})
	} else {
		for i, arg := range s.Args {
			param := opDef.Signature[i]
			if arg.Keyword != param.Kw {
				template := l.getCanonicalTemplate(opDef)
				errs = append(errs, Error{
					Code:             "LINT_CLAUSE_ORDER",
					Message:          fmt.Sprintf("%s: clause %d must be %s, got %s", s.OpName, i+1, param.Kw, arg.Keyword),
					Loc:              s.Loc,
					Hint:             fmt.Sprintf("Reorder clauses to match canonical form: %s", template),
					ExpectedTemplate: template,
				})
			}
			
			// Enum check: allow raw identifiers if they match allowed enum values
			isEnumVal := false
			if len(param.Enum) > 0 {
				if ident, ok := arg.Value.(*ast.IdentExpr); ok {
					for _, ev := range param.Enum {
						if ev == ident.Name {
							isEnumVal = true
							break
						}
					}
				}
			}

			if !isEnumVal {
				errs = append(errs, l.lintExpr(arg.Value, string(param.Type), symbols)...)
			}
		}
	}

	// 2. Enforce INTO presence
	if opDef.Into && s.Into == "" {
		errs = append(errs, Error{
			Code:    "LINT_MISSING_INTO",
			Message: fmt.Sprintf("%s: mandatory INTO clause missing", s.OpName),
			Loc:     s.Loc,
		})
	}

	// 3. Check INTO type annotation if present
	if s.IntoType != "" && opDef.ResultType != "" && s.IntoType != string(opDef.ResultType) {
		errs = append(errs, Error{
			Code:    "LINT_TYPE_MISMATCH",
			Message: fmt.Sprintf("%s: INTO type annotation mismatch: expected %s, got %s", s.OpName, opDef.ResultType, s.IntoType),
			Loc:     s.Loc,
		})
	}

	return errs, string(opDef.ResultType)
}

func (l *Linter) lintExpr(expr ast.Expr, expectedType string, symbols map[string]string) []Error {
	var errs []Error
	actualType := ""

	switch e := expr.(type) {
	case *ast.IdentExpr:
		var ok bool
		actualType, ok = symbols[e.Name]
		if !ok {
			// PROMPT is a special global
			if e.Name == "PROMPT" {
				actualType = "TEXT"
			} else {
				errs = append(errs, Error{
					Code:    "LINT_UNDEFINED_VAR",
					Message: fmt.Sprintf("undefined variable: %s", e.Name),
					Loc:     e.Pos(),
				})
				return errs
			}
		}
	case *ast.StringExpr:
		actualType = "TEXT" // We treat quoted strings as TEXT in v0.1/v0.2
	case *ast.IntExpr:
		actualType = "INT"
	case *ast.BoolExpr:
		actualType = "BOOL"
	case *ast.NullExpr:
		actualType = "NULL"
	}

	if expectedType != "" && actualType != "" && expectedType != actualType {
		// Small exception: NULL matches anything? Usually not in strict DSLs.
		// v0.1/v0.2 strictness says they must match.
		if actualType != "NULL" {
			errs = append(errs, Error{
				Code:    "LINT_TYPE_MISMATCH",
				Message: fmt.Sprintf("type mismatch: expected %s, got %s", expectedType, actualType),
				Loc:     expr.Pos(),
			})
		}
	}

	return errs
}

func (l *Linter) getCanonicalTemplate(op *ops.Op) string {
	res := op.Name
	for _, p := range op.Signature {
		res += " " + p.Kw + " <expr>"
	}
	if op.Into {
		res += " INTO <ident>"
		if op.ResultType != "" {
			res += ": " + string(op.ResultType)
		}
	}
	return res
}

type Error struct {
	Code             string
	Message          string
	Loc              interface{}
	Hint             string
	ExpectedTemplate string
}
