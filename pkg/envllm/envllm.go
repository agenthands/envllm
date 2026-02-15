package envllm

import (
	"context"
	"errors"
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/parse"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/internal/store"
)

// Program represents a compiled RLM-DSL program.
type Program struct {
	AST *ast.Program
}

// Compile compiles source code into a Program.
func Compile(filename string, src string) (*Program, error) {
	l := lex.NewLexer(filename, src)
	p := parse.NewParser(l)
	astProg, err := p.Parse()
	if err != nil {
		return nil, err
	}
	return &Program{AST: astProg}, nil
}

// ExecOptions defines the options for program execution.
type ExecOptions struct {
	Host      runtime.Host
	Policy    runtime.Policy
	Inputs    map[string]runtime.Value
	TextStore runtime.TextStore
}

// NewTextStore creates a new TextStore.
func NewTextStore() runtime.TextStore {
	return store.NewTextStore()
}

// Execute executes the program using the provided options.
func (p *Program) Execute(ctx context.Context, opt ExecOptions) (runtime.ExecResult, error) {
	ts := opt.TextStore
	if ts == nil {
		ts = store.NewTextStore()
	}

	// Load ops table (default path for now)
	tbl, err := ops.LoadTable("assets/ops.json")
	if err != nil {
		// Fallback for tests or relative paths
		tbl, err = ops.LoadTable("../../assets/ops.json")
		if err != nil {
			return runtime.ExecResult{}, fmt.Errorf("failed to load ops table: %v", err)
		}
	}
	reg := ops.NewRegistry(tbl)

	s := runtime.NewSession(opt.Policy, ts)
	s.Dispatcher = reg
	s.Host = opt.Host

	// Set inputs
	for k, v := range opt.Inputs {
		if err := s.Env.Define(k, v); err != nil {
			return runtime.ExecResult{}, err
		}
	}

	var lastErr error
	for _, cell := range p.AST.Cells {
		if err := s.ExecuteCell(ctx, cell); err != nil {
			lastErr = err
			break
		}
	}

	status := "ok"
	var errs []runtime.Error
	if lastErr != nil {
		status = "error"
		var bErr *runtime.BudgetExceededError
		var cErr *runtime.CapabilityDeniedError
		if errors.As(lastErr, &bErr) {
			status = "budget_exceeded"
		} else if errors.As(lastErr, &cErr) {
			status = "capability_denied"
		}
		
		errs = append(errs, runtime.Error{
			Code:    "EXEC_ERROR",
			Message: lastErr.Error(),
		})
	}

	return s.GenerateResult(status, errs), nil
}
