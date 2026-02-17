package harness

import (
	"context"
	"fmt"
	"strings"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/lint"
	"github.com/agenthands/envllm/internal/metrics"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/rewrite"
	"github.com/agenthands/envllm/pkg/envllm"
)

type Model interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Harness struct {
	model         Model
	table         *ops.Table
	linter        *lint.Linter
	Metrics       *metrics.SessionMetrics
	rewriteEngine *rewrite.Engine
}

func NewHarness(m Model, table *ops.Table) *Harness {
	return &Harness{
		model:         m,
		table:         table,
		linter:        lint.NewLinter(table),
		Metrics:       &metrics.SessionMetrics{},
		rewriteEngine: rewrite.NewEngine(rewrite.DefaultRegistry(table)),
	}
}

// GenerateStepByStep implements a high-reliability generation protocol with a repair loop.
func (h *Harness) GenerateStepByStep(ctx context.Context, task string, initialPrompt string) (*ast.Program, error) {
	// We use a single-pass with aggressive repair to ensure full logic is delivered.
	prog, err := h.generateWithRepair(ctx, task, initialPrompt, "full")
	if err != nil {
		return nil, err
	}

	return prog, nil
}

func (h *Harness) generateWithRepair(ctx context.Context, task string, contextStr string, mode string) (*ast.Program, error) {
	prompt := fmt.Sprintf("Task: %s\nPrompt Content: %s\n\nWrite an EnvLLM-DSL v0.2.3 program.", task, contextStr)
	if mode == "full" {
		prompt += `
Follow this hierarchy:
RLMDSL 0.2
TASK <name>:
  INPUT PROMPT: TEXT
  [REQUIRES capability="..."]
  CELL <step_name>:
    <OP> ... INTO <var>: <Type>
  OUTPUT <var>

You MUST use the 'PROMPT' variable to access the 'Prompt Content'.
CRITICAL: Solve the Task LITERALLY and PRECISELY. 
- If the task asks for a specific value, write logic to extract it.
- Use dynamic search (FIND_TEXT/FIND_REGEX) to find data.
- Handle all offsets with 100% precision.`
	}

	maxRetries := 3
	var lastErrors string
	// var lastProg *ast.Program

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
		// fmt.Printf("[DEBUG] Raw Model Response:\n%s\n", resp)
		// fmt.Printf("[DEBUG] Cleaned DSL:\n%s\n", dslCode)

		prog, parseErr := envllm.Compile("harness.rlm", dslCode, envllm.ModeStrict)
		
		h.Metrics.RecordParse(parseErr == nil)
		if parseErr != nil {
			h.Metrics.RecordRepair()
			lastErrors = fmt.Sprintf("Parse Error: %v", parseErr)
			continue
		}

		// Apply Heuristic Recovery (Anka-style)
		patchedProg, appliedRules, err := h.rewriteEngine.AutoRepair(ctx, prog.AST)
		if err == nil && len(appliedRules) > 0 {
			fmt.Printf("[DEBUG] Applied AutoRepair rules: %v\n", appliedRules)
			prog.AST = patchedProg
		}

		// Ensure LLM capability if SUBCALL is used (Suite G fix)
		hasSubcall := false
		ast.Walk(&subcallVisitor{onFound: func() { hasSubcall = true }}, prog.AST)
		if hasSubcall {
			hasCap := false
			for _, item := range prog.AST.Task.Body {
				if req, ok := item.(*ast.Requirement); ok && req.Capability == "llm" {
					hasCap = true
					break
				}
			}
			if !hasCap {
				fmt.Println("[DEBUG] Auto-injecting 'llm' capability for SUBCALL")
				prog.AST.Task.Body = append([]ast.BodyItem{&ast.Requirement{Capability: "llm"}}, prog.AST.Task.Body...)
			}
		}

		lintErrs := h.linter.Lint(prog.AST)
		h.Metrics.RecordLint(len(lintErrs) == 0)
		if len(lintErrs) == 0 {
			return prog.AST, nil
		}

		h.Metrics.RecordRepair()
		// lastProg = prog.AST

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

type subcallVisitor struct {
	onFound func()
}

func (v *subcallVisitor) Visit(node ast.Node) ast.Visitor {
	if op, ok := node.(*ast.OpStmt); ok && op.OpName == "SUBCALL" {
		v.onFound()
	}
	return v
}

func (h *Harness) checkDrift(old, new *ast.Program) error {
	oldCells := h.getCells(old)
	newCells := h.getCells(new)

	if len(oldCells) != len(newCells) {
		return fmt.Errorf("number of cells changed from %d to %d", len(oldCells), len(newCells))
	}
	for i := range oldCells {
		if oldCells[i].Name != newCells[i].Name {
			return fmt.Errorf("cell name changed from %q to %q at index %d", oldCells[i].Name, newCells[i].Name, i)
		}
	}
	return nil
}

func (h *Harness) getCells(p *ast.Program) []*ast.Cell {
	var cells []*ast.Cell
	if p.Task == nil {
		return cells
	}
	for _, item := range p.Task.Body {
		if c, ok := item.(*ast.Cell); ok {
			cells = append(cells, c)
		}
	}
	return cells
}

func (h *Harness) cleanDSL(s string) string {
	s = strings.TrimSpace(s)
	// 1. Look for markdown code blocks
	if strings.Contains(s, "```") {
		start := strings.Index(s, "```")
		// Find end of the first line (e.g. ```rlm)
		lineEnd := strings.Index(s[start:], "\n")
		if lineEnd != -1 {
			end := strings.LastIndex(s, "```")
			if end > start+lineEnd {
				return strings.TrimSpace(s[start+lineEnd : end])
			}
		}
	}
	// 2. If no code blocks, look for TASK or RLMDSL keywords at start of line
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "RLMDSL") {
			return strings.Join(lines[i:], "\n")
		}
		if strings.HasPrefix(trimmed, "TASK ") {
			return "RLMDSL 0.2\n" + strings.Join(lines[i:], "\n")
		}
	}
	return s
}
