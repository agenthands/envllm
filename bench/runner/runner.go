package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/pkg/envllm"
)

type Model interface {
	Complete(ctx context.Context, task, prompt string) (string, error)
}

type Case struct {
	ID          string         `json:"id"`
	Suite       string         `json:"suite"`
	PromptRef   string         `json:"prompt_ref"`
	Task        string         `json:"task"`
	Policy      runtime.Policy `json:"policy"`
	ExpectedRef string         `json:"expected_ref,omitempty"`
	Scoring     ScoringConfig  `json:"scoring"`
	Host        runtime.Host   `json:"-"`
}

type dummyHost struct{}

func (h *dummyHost) Subcall(ctx context.Context, req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
	return runtime.SubcallResponse{
		Result: runtime.Value{Kind: runtime.KindJSON, V: map[string]interface{}{}},
	}, nil
}

type ScoringConfig struct {
	Mode string `json:"mode"` // "status_ok", "status_budget_exceeded", "status_capability_denied", "json_semantic"
}

type Result struct {
	CaseID string             `json:"case_id"`
	Passed bool               `json:"passed"`
	Status string             `json:"status"`
	Output runtime.ExecResult `json:"output"`
	Error  string             `json:"error,omitempty"`
}

func RunCase(ctx context.Context, c Case, m Model, baseDir string) (Result, error) {
	promptPath := filepath.Join(baseDir, c.PromptRef)
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read prompt: %v", err)
	}

	dslCode, err := m.Complete(ctx, c.Task, string(promptBytes))
	if err != nil {
		return Result{CaseID: c.ID, Error: fmt.Sprintf("model failed: %v", err)}, nil
	}

	prog, err := envllm.Compile(c.ID+".rlm", dslCode)
	if err != nil {
		return Result{CaseID: c.ID, Status: "compile_error", Error: fmt.Sprintf("compile failed: %v", err), Passed: c.Scoring.Mode == "status_compile_error"}, nil
	}

	// Setup store and PROMPT input
	ts := envllm.NewTextStore()
	ph := ts.Add(string(promptBytes))

	// Provide a dummy host for SUBCALL tests
	host := c.Host
	if host == nil {
		host = &dummyHost{}
	}

	opt := envllm.ExecOptions{
		Policy:    c.Policy,
		TextStore: ts,
		Host:      host,
		Inputs: map[string]runtime.Value{
			"PROMPT": {Kind: runtime.KindText, V: ph},
		},
	}
	
	execRes, err := prog.Execute(ctx, opt)
	if err != nil {
		// Execution errors (like panic or system error) are different from DSL status errors
		return Result{CaseID: c.ID, Error: fmt.Sprintf("execution crashed: %v", err), Passed: false}, nil
	}

	passed := scoreResult(c, execRes, baseDir)

	return Result{
		CaseID: c.ID,
		Passed: passed,
		Status: execRes.Status,
		Output: execRes,
	}, nil
}

func scoreResult(c Case, res runtime.ExecResult, baseDir string) bool {
	switch c.Scoring.Mode {
	case "status_ok":
		return res.Status == "ok"
	case "status_budget_exceeded":
		return res.Status == "budget_exceeded"
	case "status_capability_denied":
		return res.Status == "capability_denied"
	case "status_error":
		return res.Status == "error"
	case "status_compile_error":
		return res.Status == "compile_error"
	case "json_semantic":
		if res.Status != "ok" || res.Final == nil {
			return false
		}
		if c.ExpectedRef == "" {
			return true
		}
		expectedPath := filepath.Join(baseDir, c.ExpectedRef)
		expectedBytes, err := os.ReadFile(expectedPath)
		if err != nil {
			return false
		}
		var expectedVal interface{}
		if err := json.Unmarshal(expectedBytes, &expectedVal); err != nil {
			return false
		}
		return reflect.DeepEqual(res.Final.V, expectedVal)
	default:
		return res.Status == "ok"
	}
}
