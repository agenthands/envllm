package main

import (
	"context"
	"os"
	"testing"

	"github.com/agenthands/envllm/examples/bridge"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/internal/store"
	"github.com/tmc/langchaingo/llms/googleai"
)

func TestLLMIntegration_Gemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY or GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	// Attempt to use a stronger model if available, or default.
	// langchaingo/llms/googleai defaults to gemini-pro.
	// We can specify model via option.
	model, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-2.0-flash"))
	if err != nil {
		t.Fatalf("failed to create googleai model: %v", err)
	}

	ts := store.NewTextStore()
	host := bridge.NewLangChainHost(model, ts)
	card, err := os.ReadFile("../../assets/syntax_guide.md")
	if err != nil {
		t.Fatalf("failed to read dialect card: %v", err)
	}
	host.DialectCard = string(card)

	// Real Task: Extract a specific piece of information from a structured log.
	prompt := `
SYSTEM LOG - 2026-02-15
[ID: 1001] START
[ID: 1002] USER_LOGIN: user="admin", status="success"
[ID: 1003] ACTION: type="upload", target="config.json", size="45KB"
[ID: 1004] ERROR: database="timeout", retry="3"
[ID: 1005] END
`
	ph := ts.Add(prompt)

	policy := runtime.Policy{
		MaxStmtsPerCell:     50,
		MaxSubcalls:         5,
		MaxRecursionDepth:   3,
		AllowedCapabilities: map[string]bool{"llm": true},
	}

	task := "Extract the ID associated with the ERROR in the log."

	t.Logf("Running real LLM session with Gemini Flash...")
	res, err := host.RunSession(ctx, task, ph, policy)
	if err != nil {
		t.Fatalf("Session failed: %v", err)
	}

	if res.Status != "ok" {
		t.Fatalf("Expected status ok, got %s. Errors: %+v", res.Status, res.Errors)
	}

	if res.Final == nil {
		t.Fatalf("Expected final result, got nil")
	}

	t.Logf("Final Result Kind: %s", res.Final.Kind)
	t.Logf("Final Result Value: %+v", res.Final.V)

	if res.Final.Kind == runtime.KindText {
		finalText, _ := ts.Get(res.Final.V.(runtime.TextHandle))
		t.Logf("Final Text: %q", finalText)
	}

	// Verification: The model should have extracted {"database": "timeout", "retry": "3"}
	// (or similar semantic structure)
}
