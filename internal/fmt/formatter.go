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

	for i, cell := range prog.Cells {
		if i > 0 || prog.Version != "" {
			sb.WriteString("\n")
		}
		formatCell(&sb, cell)
	}

	return sb.String()
}

func formatCell(sb *strings.Builder, cell *ast.Cell) {
	sb.WriteString("CELL ")
	sb.WriteString(cell.Name)
	sb.WriteString(":\n")

	for _, stmt := range cell.Stmts {
		sb.WriteString("  ")
		formatStmt(sb, stmt)
		sb.WriteString("\n")
	}
}

func formatStmt(sb *strings.Builder, stmt ast.Stmt) {
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
