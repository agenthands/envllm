package fmt

import (
	"fmt"
	"strings"

	"github.com/agenthands/envllm/internal/ast"
)

// Format produces a canonical string representation of an AST.
func Format(prog *ast.Program) string {
	var sb strings.Builder

	if prog.Version != "" {
		sb.WriteString("RLMDSL ")
		sb.WriteString(prog.Version)
		sb.WriteString("\n")
	}

	if prog.Dialect != "" {
		sb.WriteString("DIALECT ")
		sb.WriteString(prog.Dialect)
		sb.WriteString("\n")
	}

	for ext, ver := range prog.Extensions {
		sb.WriteString("EXT ")
		sb.WriteString(ext)
		sb.WriteString("=")
		sb.WriteString(ver)
		sb.WriteString("\n")
	}

	if prog.Task != nil {
		if prog.Version != "" || prog.Dialect != "" || len(prog.Extensions) > 0 {
			sb.WriteString("\n")
		}
		formatTask(&sb, prog.Task)
	}

	return sb.String()
}

func formatTask(sb *strings.Builder, t *ast.Task) {
	sb.WriteString("TASK ")
	sb.WriteString(t.Name)
	sb.WriteString(":\n")

	for _, in := range t.Inputs {
		sb.WriteString("  INPUT ")
		sb.WriteString(in.Name)
		sb.WriteString(": ")
		sb.WriteString(in.Type)
		sb.WriteString("\n")
	}

	formatBody(sb, t.Body, 2)

	sb.WriteString("  OUTPUT ")
	sb.WriteString(t.Output)
	sb.WriteString("\n")
}

func formatBody(sb *strings.Builder, body []ast.BodyItem, indent int) {
	indentStr := strings.Repeat(" ", indent)
	for _, item := range body {
		switch it := item.(type) {
		case *ast.Requirement:
			sb.WriteString(indentStr)
			sb.WriteString("REQUIRES capability=")
			sb.WriteString(quoteString(it.Capability))
			sb.WriteString("\n")
		case *ast.Cell:
			sb.WriteString(indentStr)
			sb.WriteString("CELL ")
			sb.WriteString(it.Name)
			sb.WriteString(":\n")
			for _, stmt := range it.Stmts {
				sb.WriteString(indentStr)
				sb.WriteString("  ")
				formatStmt(sb, stmt, indent+2)
				sb.WriteString("\n")
			}
		case *ast.IfStmt:
			sb.WriteString(indentStr)
			sb.WriteString("IF ")
			formatExpr(sb, it.Cond)
			sb.WriteString(":\n")
			formatBody(sb, it.ThenBody, indent)
			if it.ElseBody != nil {
				sb.WriteString(indentStr)
				sb.WriteString("ELSE:\n")
				formatBody(sb, it.ElseBody, indent)
			}
			sb.WriteString(indentStr)
			sb.WriteString("END\n")
		}
	}
}

func formatStmt(sb *strings.Builder, stmt ast.Stmt, indent int) {
	switch s := stmt.(type) {
	case *ast.OpStmt:
		sb.WriteString(s.OpName)
		for _, arg := range s.Args {
			sb.WriteString(" ")
			sb.WriteString(arg.Keyword)
			sb.WriteString(" ")
			formatExpr(sb, arg.Value)
		}
		sb.WriteString(" INTO ")
		sb.WriteString(s.Into)
		if s.IntoType != "" {
			sb.WriteString(": ")
			sb.WriteString(s.IntoType)
		}
	case *ast.SetFinalStmt:
		sb.WriteString("SET_FINAL SOURCE ")
		formatExpr(sb, s.Source)
	case *ast.AssertStmt:
		sb.WriteString("ASSERT COND ")
		formatExpr(sb, s.Cond)
		sb.WriteString(" MESSAGE ")
		sb.WriteString(quoteString(s.Message))
	case *ast.PrintStmt:
		sb.WriteString("PRINT SOURCE ")
		formatExpr(sb, s.Source)
	case *ast.ForEachStmt:
		sb.WriteString("FOR_EACH ")
		sb.WriteString(s.Iterator)
		sb.WriteString(" IN ")
		sb.WriteString(s.Collection)
		sb.WriteString(" LIMIT ")
		sb.WriteString(fmt.Sprintf("%d", s.Limit))
		sb.WriteString(":\n")
		indentStr := strings.Repeat(" ", indent+2)
		for _, bs := range s.Body {
			sb.WriteString(indentStr)
			formatStmt(sb, bs, indent+2)
			sb.WriteString("\n")
		}
	}
}

func formatExpr(sb *strings.Builder, expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		sb.WriteString(e.Name)
	case *ast.StringExpr:
		sb.WriteString(quoteString(e.Value))
	case *ast.IntExpr:
		sb.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.BoolExpr:
		if e.Value {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case *ast.NullExpr:
		sb.WriteString("null")
	}
}

func quoteString(s string) string {
	r := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"\n", "\\n",
		"\t", "\\t",
		"\r", "\\r",
	)
	return "\"" + r.Replace(s) + "\""
}
