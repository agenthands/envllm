package harness

import (
	"context"
	"fmt"
	"strings"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lint"
	"github.com/agenthands/envllm/internal/metrics"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/pkg/envllm"
)

type Model interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Harness struct {
	model   Model
	table   *ops.Table
	linter  *lint.Linter
	Metrics *metrics.SessionMetrics
}

func NewHarness(m Model, table *ops.Table) *Harness {
	return &Harness{
		model:   m,
		table:   table,
		linter:  lint.NewLinter(table),
		Metrics: &metrics.SessionMetrics{},
	}
}

// GenerateStepByStep implements the two-pass generation protocol with a repair loop.
func (h *Harness) GenerateStepByStep(ctx context.Context, task string, initialPrompt string) (*ast.Program, error) {
	// Pass 1: Skeleton
	skeleton, err := h.generateWithRepair(ctx, task, initialPrompt, "skeleton")
	if err != nil {
		return nil, err
	}

	return skeleton, nil
}

func (h *Harness) generateWithRepair(ctx context.Context, task string, contextStr string, mode string) (*ast.Program, error) {
	prompt := fmt.Sprintf("Task: %s\nContext: %s\n\nWrite an EnvLLM-DSL program.", task, contextStr)
	if mode == "skeleton" {
		prompt += " Focus on the skeleton (CELL names and ops only)."
	}

	maxRetries := 3
	var lastErrors string
	var lastProg *ast.Program

	for i := 0; i < maxRetries; i++ {
		currentPrompt := prompt
		if lastErrors != "" {
			currentPrompt += "\n\nYour previous output had errors:\n" + lastErrors + "\n\nPlease fix them. Only output the corrected DSL."
		}

		resp, err := h.model.Complete(ctx, currentPrompt)
		if err != nil {
			return nil, err
		}

		dslCode := h.cleanDSL(resp)
		prog, parseErr := envllm.Compile("harness.rlm", dslCode, envllm.ModeStrict)
		
		h.Metrics.RecordParse(parseErr == nil)
		if parseErr != nil {
			h.Metrics.RecordRepair()
			lastErrors = fmt.Sprintf("Parse Error: %v", parseErr)
			continue
		}

		// Drift Guard (EPIC F3)
		if lastProg != nil {
			if driftErr := h.checkDrift(lastProg, prog.AST); driftErr != nil {
				h.Metrics.RecordRepair()
				lastErrors = fmt.Sprintf("Drift Violation: %v", driftErr)
				continue
			}
		}

		lintErrs := h.linter.Lint(prog.AST)
		h.Metrics.RecordLint(len(lintErrs) == 0)
		if len(lintErrs) == 0 {
			return prog.AST, nil
		}

		h.Metrics.RecordRepair()
		lastProg = prog.AST

		// Format linter errors for feedback
		var sb strings.Builder
		for _, le := range lintErrs {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", le.Code, le.Message))
			if le.Hint != "" {
				sb.WriteString(fmt.Sprintf("  Hint: %s\n", le.Hint))
			} else if le.ExpectedTemplate != "" {
				sb.WriteString(fmt.Sprintf("  Expected: %s\n", le.ExpectedTemplate))
			}
		}
		lastErrors = sb.String()
	}

	return nil, fmt.Errorf("failed to generate valid DSL after %d retries. Last errors:\n%s", maxRetries, lastErrors)
}

func (h *Harness) checkDrift(old, new *ast.Program) error {
	if len(old.Cells) != len(new.Cells) {
		return fmt.Errorf("number of cells changed from %d to %d", len(old.Cells), len(new.Cells))
	}
	for i := range old.Cells {
		if old.Cells[i].Name != new.Cells[i].Name {
			return fmt.Errorf("cell name changed from %q to %q at index %d", old.Cells[i].Name, new.Cells[i].Name, i)
		}
	}
	return nil
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
