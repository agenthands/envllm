package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/pkg/envllm"
)

type Model interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Case struct {
	ID               string         `json:"id"`
	Suite            string         `json:"suite"`
	PromptRef        string         `json:"prompt_ref"`
	Task             string         `json:"task"`
	Policy           runtime.Policy `json:"policy"`
}

type Result struct {
	CaseID string             `json:"case_id"`
	Passed bool               `json:"passed"`
	Output runtime.ExecResult `json:"output"`
	Error  string             `json:"error,omitempty"`
}

func RunCase(ctx context.Context, c Case, m Model, baseDir string) (Result, error) {
	promptPath := filepath.Join(baseDir, c.PromptRef)
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read prompt: %v", err)
	}

	modelPrompt := fmt.Sprintf("Task: %s\\nPrompt Content: %s\\n\\nWrite EnvLM DSL code.", c.Task, string(promptBytes))
	dslCode, err := m.Complete(ctx, modelPrompt)
	if err != nil {
		return Result{CaseID: c.ID, Error: fmt.Sprintf("model failed: %v", err)}, nil
	}

	prog, err := envllm.Compile(c.ID+".rlm", dslCode)
	if err != nil {
		return Result{CaseID: c.ID, Error: fmt.Sprintf("compile failed: %v", err)}, nil
	}

	opt := envllm.ExecOptions{
		Policy: c.Policy,
		Inputs: make(map[string]runtime.Value),
	}
	
	execRes, err := prog.Execute(ctx, opt)
	if err != nil {
		return Result{CaseID: c.ID, Error: fmt.Sprintf("execution failed: %v", err)}, nil
	}

	return Result{
		CaseID: c.ID,
		Passed: execRes.Status == "ok",
		Output: execRes,
	}, nil
}
