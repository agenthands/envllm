package harness

import (
	"context"
	"fmt"
	"strings"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/fmt"
	"github.com/agenthands/envllm/internal/lint"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/pkg/envllm"
)

type Model interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Harness struct {
	model  Model
	table  *ops.Table
	linter *lint.Linter
}

func NewHarness(m Model, table *ops.Table) *Harness {
	return &Harness{
		model:  m,
		table:  table,
		linter: lint.NewLinter(table),
	}
}

// GenerateStepByStep implements the two-pass generation protocol.
func (h *Harness) GenerateStepByStep(ctx context.Context, task string, initialPrompt string) (*ast.Program, error) {
	// Pass 1: Skeleton
	skeleton, err := h.generateSkeleton(ctx, task, initialPrompt)
	if err != nil {
		return nil, err
	}

	// Pass 2: Fill (step by step)
	// For each step in the skeleton, we prompt the model to fill it.
	// This is a bit complex for a single turn DSL, but the plan says "step by step".
	// In the context of RLM, it means generating one CELL at a time.
	
	return skeleton, nil // Placeholder for now, Pass 1 is the main skeleton
}

func (h *Harness) generateSkeleton(ctx context.Context, task string, initialPrompt string) (*ast.Program, error) {
	prompt := fmt.Sprintf("Task: %s
Context: %s

Write the skeleton of an EnvLLM-DSL program (CELL names and ops only).", task, initialPrompt)
	
	resp, err := h.model.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Basic cleaning
	resp = h.cleanDSL(resp)

	prog, err := envllm.Compile("skeleton.rlm", resp, envllm.ModeStrict)
	if err != nil {
		return nil, fmt.Errorf("skeleton failed validation: %v", err)
	}

	return prog.AST, nil
}

func (h *Harness) cleanDSL(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "```") {
		start := strings.Index(s, "```")
		end := strings.LastIndex(s, "```")
		if start != -1 && end != -1 && start != end {
			content := s[start+3 : end]
			if strings.HasPrefix(content, "rlm") { content = content[3:] }
			return strings.TrimSpace(content)
		}
	}
	return s
}
