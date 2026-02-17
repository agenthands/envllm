package lint

import (
	"fmt"
	"strings"
	"time"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/rewrite"
	"github.com/agenthands/envllm/internal/trace"
)

type Linter struct {
	table    *ops.Table
	sink     trace.Sink
	registry *rewrite.Registry
	mode     Mode
}

type Mode int

const (
	ModeCompat Mode = iota
	ModeStrict
)

func NewLinter(table *ops.Table) *Linter {
	return &Linter{table: table, mode: ModeCompat}
}

func (l *Linter) WithMode(mode Mode) *Linter {
	l.mode = mode
	return l
}

func (l *Linter) WithSink(sink trace.Sink) *Linter {
	l.sink = sink
	return l
}

func (l *Linter) WithRegistry(registry *rewrite.Registry) *Linter {
	l.registry = registry
	return l
}

func (l *Linter) emitTrace(step trace.TraceStep) {
	if l.sink != nil {
		if step.Timestamp.IsZero() {
			step.Timestamp = time.Now()
		}
		if step.Phase == "" {
			step.Phase = trace.PhaseLint
		}
		l.sink.Emit(step)
	}
}

func (l *Linter) Lint(prog *ast.Program) []Error {
	var errs []Error
	symbols := make(map[string]string) // name -> type
	requiredCaps := make(map[string]bool)

	if prog.Task == nil {
		return errs
	}

	// 1. Process Inputs
	for _, in := range prog.Task.Inputs {
		symbols[in.Name] = in.Type
	}

	// 2. Process Body
	errs = append(errs, l.lintBody(prog.Task.Body, symbols, requiredCaps)...)

	// 3. Process Output
	if prog.Task.Output != "" {
		if _, ok := symbols[prog.Task.Output]; !ok {
			err := Error{
				Code:    "LINT_UNDEFINED_VAR",
				Message: fmt.Sprintf("task output variable %q not defined", prog.Task.Output),
				Loc:     prog.Task.Loc,
			}
			errs = append(errs, err)
			l.emitTrace(trace.TraceStep{
				Decision: trace.DecisionReject,
				Error:    &trace.TraceError{Code: err.Code, Message: err.Message},
			})
		}
	}

	if len(errs) == 0 {
		l.emitTrace(trace.TraceStep{
			Decision: trace.DecisionAccept,
		})
	}

	return errs
}

func (l *Linter) lintBody(body []ast.BodyItem, symbols map[string]string, requiredCaps map[string]bool) []Error {
	var errs []Error
	for _, item := range body {
		errs = append(errs, l.lintBodyItem(item, symbols, requiredCaps)...)
	}
	return errs
}

func (l *Linter) lintBodyItem(item ast.BodyItem, symbols map[string]string, requiredCaps map[string]bool) []Error {
	switch it := item.(type) {
	case *ast.Requirement:
		requiredCaps[it.Capability] = true
		return nil
	case *ast.Cell:
		return l.lintStmts(it.Stmts, symbols, requiredCaps)
	case *ast.IfStmt:
		var errs []Error
		errs = append(errs, l.lintExpr(it.Cond, "BOOL", symbols)...)
		errs = append(errs, l.lintBody(it.ThenBody, symbols, requiredCaps)...)
		if it.ElseBody != nil {
			errs = append(errs, l.lintBody(it.ElseBody, symbols, requiredCaps)...)
		}
		return errs
	case ast.Stmt:
		return l.lintStmts([]ast.Stmt{it}, symbols, requiredCaps)
	default:
		return []Error{{Code: "LINT_INTERNAL_ERROR", Message: fmt.Sprintf("unhandled body item: %T", item)}}
	}
}

func (l *Linter) lintStmts(stmts []ast.Stmt, symbols map[string]string, requiredCaps map[string]bool) []Error {
	var errs []Error
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.OpStmt:
			opErrs, outType := l.lintOpStmt(s, symbols, requiredCaps)
			errs = append(errs, opErrs...)
			if s.Into != "" {
				if _, exists := symbols[s.Into]; exists {
					errs = append(errs, Error{
						Code:    "LINT_VAR_REUSE_FORBIDDEN",
						Message: fmt.Sprintf("variable %q already defined", s.Into),
						Loc:     s.Loc,
						Hint:    fmt.Sprintf("Rename to %s_2 or %s_step%d", s.Into, s.Into, len(symbols)),
					})
				} else {
					if outType != "" {
						symbols[s.Into] = outType
					} else {
						symbols[s.Into] = "UNKNOWN"
					}
				}
			}
		case *ast.SetFinalStmt:
			errs = append(errs, l.lintExpr(s.Source, "", symbols)...)
		case *ast.PrintStmt:
			errs = append(errs, l.lintExpr(s.Source, "", symbols)...)
		case *ast.AssertStmt:
			errs = append(errs, l.lintExpr(s.Cond, "BOOL", symbols)...)
		case *ast.ForEachStmt:
			// No-op for now
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
			err := Error{
				Code:    "LINT_MISSING_REQUIRES",
				Message: fmt.Sprintf("operation %s requires capability %q but it was not declared with REQUIRES", s.OpName, c),
				Loc:     s.Loc,
				Hint:    fmt.Sprintf("Add 'REQUIRES capability=%q' to the top of your program.", c),
			}
			errs = append(errs, err)
			l.emitTrace(trace.TraceStep{
				Decision: trace.DecisionReject,
				Op:       s.OpName,
				Error:    &trace.TraceError{Code: err.Code, Message: err.Message},
				RuleID:   "RULE_MISSING_REQUIRES",
				Hint:     err.Hint,
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
			// ... (existing order check)

			// JSON Usage Check
			if s.OpName == "JSON_GET" && param.Kw == "SOURCE" {
				if id, ok := arg.Value.(*ast.IdentExpr); ok {
					if typ, known := symbols[id.Name]; known && typ == "STRUCT" {
						errs = append(errs, Error{
							Code:    "LINT_JSON_USED_FOR_STRUCTURAL_DATA",
							Message: fmt.Sprintf("JSON_GET used on STRUCT variable %q", id.Name),
							Loc:     s.Loc,
							Hint:    "Use GET_FIELD or a specialized getter (e.g. GET_COST) instead.",
						})
					}
				}
			}

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

			// EPIC: Forbid literal offset arithmetic in STRICT mode
			if l.mode == ModeStrict && s.OpName == "OFFSET_ADD" && param.Kw == "AMOUNT" {
				if _, ok := arg.Value.(*ast.IntExpr); ok {
					errs = append(errs, Error{
						Code:    "LINT_OFFSET_ARITHMETIC_FORBIDDEN",
						Message: "Literal offset arithmetic (+7, -3) is forbidden in STRICT mode.",
						Loc:     s.Loc,
						Hint:    "Use AFTER_TEXT or FIND_REGEX groups instead of hardcoding character counts.",
					})
				}
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

			err := Error{
				Code:             "LINT_DOT_ACCESS_FORBIDDEN",
				Message:          fmt.Sprintf("Dot access (%s) is not allowed in STRICT mode.", e.Name),
				Loc:              e.Pos(),
				Hint:             hint,
				ExpectedTemplate: hint, // Reuse hint as template for now
			}
			errs = append(errs, err)
			l.emitTrace(trace.TraceStep{
				Decision: trace.DecisionReject,
				Op:       e.Name,
				Error:    &trace.TraceError{Code: err.Code, Message: err.Message},
				RuleID:   "RULE_DOT_ACCESS_TO_GETTER",
				Hint:     err.Hint,
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
			} else if expectedType == "STRUCT" && actualType == "SPAN" {
				hint = "SPAN is not a STRUCT. Use GET_SPAN_START or GET_SPAN_END to get offsets from a SPAN."
			} else if expectedType == "OFFSET" && actualType == "SPAN" {
				hint = "You passed a SPAN where an OFFSET was expected. Use GET_SPAN_START or GET_SPAN_END."
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
