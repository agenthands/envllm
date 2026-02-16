package lint

import (
	"fmt"
	"strings"

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
	requiredCaps := make(map[string]bool)

	// 1. Process Requirements
	for _, req := range prog.Requirements {
		requiredCaps[req.Capability] = true
	}

	for _, cell := range prog.Cells {
		for _, stmt := range cell.Stmts {
			switch s := stmt.(type) {
			case *ast.OpStmt:
				opErrs, outType := l.lintOpStmt(s, symbols, requiredCaps)
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

func (l *Linter) lintOpStmt(s *ast.OpStmt, symbols map[string]string, caps map[string]bool) ([]Error, string) {
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

	// Check Capability requirements
	for _, c := range opDef.Capabilities {
		if c == "pure" {
			continue
		}
		if !caps[c] {
			errs = append(errs, Error{
				Code:    "LINT_MISSING_REQUIRES",
				Message: fmt.Sprintf("operation %s requires capability %q but it was not declared with REQUIRES", s.OpName, c),
				Loc:     s.Loc,
				Hint:    fmt.Sprintf("Add 'REQUIRES capability=%q' to the top of your program.", c),
			})
		}
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
			
			// Enum check
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

	// 3. Check INTO type annotation
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
		if strings.Contains(e.Name, ".") {
			parts := strings.Split(e.Name, ".")
			obj := parts[0]
			prop := parts[1] // Simplification for 1 level
			
			// Auto-fix hint generation
			hint := fmt.Sprintf("Dot access (%s) is not allowed in STRICT mode.", e.Name)
			if prop == "cost" {
				hint = fmt.Sprintf("Use GET_COST result=%s INTO cost: COST", obj)
			} else if prop == "start" {
				hint = fmt.Sprintf("Use GET_SPAN_START span=%s INTO start: OFFSET", obj)
			} else if prop == "end" {
				hint = fmt.Sprintf("Use GET_SPAN_END span=%s INTO end: OFFSET", obj)
			} else {
				hint = fmt.Sprintf("Use JSON_GET SOURCE %s PATH \"%s\" INTO val: JSON", obj, prop)
			}

			errs = append(errs, Error{
				Code:             "LINT_DOT_ACCESS_FORBIDDEN",
				Message:          fmt.Sprintf("Dot access (%s) is not allowed in STRICT mode.", e.Name),
				Loc:              e.Pos(),
				Hint:             hint,
				ExpectedTemplate: hint, // Reuse hint as template for now
			})
			return errs
		}

		var ok bool
		actualType, ok = symbols[e.Name]
		if !ok {
			if e.Name == "PROMPT" {
				actualType = "TEXT"
			} else {
				// NO_META rule: catch common hallucinations
				if e.Name == "steps" || e.Name == "json" || e.Name == "response" {
					errs = append(errs, Error{
						Code:    "LINT_NO_META",
						Message: fmt.Sprintf("forbidden reference to meta-object: %s", e.Name),
						Loc:     e.Pos(),
						Hint:    "Do not attempt to inspect internal DSL state. Only use variables you defined via INTO.",
					})
				} else {
					errs = append(errs, Error{
						Code:    "LINT_UNDEFINED_VAR",
						Message: fmt.Sprintf("undefined variable: %s", e.Name),
						Loc:     e.Pos(),
					})
				}
				return errs
			}
		}
	case *ast.StringExpr:
		actualType = "TEXT"
	case *ast.IntExpr:
		actualType = "INT"
	case *ast.BoolExpr:
		actualType = "BOOL"
	case *ast.NullExpr:
		actualType = "NULL"
	}

	if expectedType != "" && actualType != "" && expectedType != actualType {
		if actualType != "NULL" {
			hint := ""
			if expectedType == "TEXT" {
				hint = fmt.Sprintf("Use TO_TEXT VALUE=%s to convert.", l.getExprName(expr))
			} else if expectedType == "OFFSET" && actualType == "INT" {
				hint = fmt.Sprintf("Use OFFSET VALUE=%s to create a position.", l.getExprName(expr))
			}

			errs = append(errs, Error{
				Code:    "LINT_TYPE_MISMATCH",
				Message: fmt.Sprintf("type mismatch: expected %s, got %s", expectedType, actualType),
				Loc:     expr.Pos(),
				Hint:    hint,
			})
		}
	}

	return errs
}

func (l *Linter) getExprName(e ast.Expr) string {
	if id, ok := e.(*ast.IdentExpr); ok {
		return id.Name
	}
	return "value"
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
