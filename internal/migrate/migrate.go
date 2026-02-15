package migrate

import (
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	dfmt "github.com/agenthands/envllm/internal/fmt"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/pkg/envllm"
)

type MigrationReport struct {
	Changes []string
}

func Migrate(src string, table *ops.Table) (string, *MigrationReport, error) {
	report := &MigrationReport{}

	// 1. Parse in COMPAT mode
	prog, err := envllm.Compile("migration.rlm", src, envllm.ModeCompat)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse v0.1 source: %v", err)
	}

	// 2. Normalize and infer types
	migrateProgram(prog.AST, table, report)

	// 3. Format in STRICT mode (canonical)
	canonical := dfmt.Format(prog.AST)

	return canonical, report, nil
}

func migrateProgram(prog *ast.Program, table *ops.Table, report *MigrationReport) {
	for _, cell := range prog.Cells {
		for _, stmt := range cell.Stmts {
			if op, ok := stmt.(*ast.OpStmt); ok {
				if opDef, ok := table.Ops[op.OpName]; ok {
					// Infer INTO type if missing
					if op.IntoType == "" && opDef.ResultType != "" {
						op.IntoType = string(opDef.ResultType)
						report.Changes = append(report.Changes, fmt.Sprintf("Inferred type %s for INTO %s in op %s", op.IntoType, op.Into, op.OpName))
					}
					
					// Reorder clauses if wrong? 
					// The current parser preserved the order from source.
					// If we want to strictly reorder, we'd need to match keywords to signature.
					// For v0.2, the formatter just prints what's in op.Args.
					// A full migration would reorder op.Args based on opDef.Signature.
					reorderArgs(op, opDef, report)
				}
			}
		}
	}
}

func reorderArgs(op *ast.OpStmt, def *ops.Op, report *MigrationReport) {
	if len(op.Args) != len(def.Signature) { return }
	
	newArgs := make([]ast.KwArg, len(def.Signature))
	changed := false
	for i, param := range def.Signature {
		found := false
		for _, arg := range op.Args {
			if arg.Keyword == param.Kw {
				newArgs[i] = arg
				found = true
				break
			}
		}
		if !found { return } // Something is wrong, skip reorder
		if op.Args[i].Keyword != param.Kw { changed = true }
	}
	
	if changed {
		op.Args = newArgs
		report.Changes = append(report.Changes, fmt.Sprintf("Reordered clauses for op %s", op.OpName))
	}
}
