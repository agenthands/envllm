package migrate

import (
	"context"
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/parse"
	dfmt "github.com/agenthands/envllm/internal/fmt"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/rewrite"
)

// Report contains details about the migration results.
type Report struct {
	Changes []string
}

// Migrate is a high-level helper for CLI usage.
func Migrate(src string, table *ops.Table) (string, *Report, error) {
	l := lex.NewLexer("migrate.rlm", src)
	p := parse.NewParser(l, parse.ModeCompat)
	prog, err := p.Parse()
	if err != nil {
		return "", nil, err
	}

	m := NewMigrator(table)
	patched, applied, err := m.MigrateV01ToV02(context.Background(), prog)
	if err != nil {
		return "", nil, err
	}

	report := &Report{Changes: applied}
	return dfmt.Format(patched), report, nil
}

// Migrator handles version-to-version transformations.
type Migrator struct {
	rewriteEngine *rewrite.Engine
	opsTable      *ops.Table
}

func NewMigrator(table *ops.Table) *Migrator {
	registry := rewrite.DefaultRegistry(table)
	return &Migrator{
		rewriteEngine: rewrite.NewEngine(registry),
		opsTable:      table,
	}
}

// MigrateV01ToV02 upgrades a v0.1 program to v0.2.
func (m *Migrator) MigrateV01ToV02(ctx context.Context, prog *ast.Program) (*ast.Program, []string, error) {
	// v0.1 to v0.2 involves:
	// 1. Fixing missing REQUIRES (mandatory in v0.2)
	// 2. Fixing dot access (forbidden in v0.2)
	// 3. Ensuring type-safe concatenation
	
	// We can use the AutoRepair engine which applies these rules iteratively.
	patched, applied, err := m.rewriteEngine.AutoRepair(ctx, prog)
	if err != nil {
		return nil, nil, fmt.Errorf("migration failed during rewrite: %v", err)
	}
	
	// Update version header
	prog.Version = "0.2"
	
	return patched, applied, nil
}
