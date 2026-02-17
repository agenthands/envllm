package main

import (
	"context"
	"os"
	"testing"

	"github.com/agenthands/envllm/bench/runner"
	"github.com/agenthands/envllm/examples/bridge"
	"github.com/agenthands/envllm/internal/store"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type LLMModel struct {
	host *bridge.LangChainHost
}

func (m *LLMModel) Complete(ctx context.Context, caseID, task, prompt string) (string, error) {
	_ = prompt // Unused in this simple turn driver
	_ = caseID
	obsJSON := "{}" // Simplified for single turn completion test
	dialectCard := m.host.DialectCard
	
	llmPrompt := bridge.FormatPrompt(dialectCard, task, obsJSON)
	return llms.GenerateFromSinglePrompt(ctx, m.host.Model, llmPrompt)
}

func TestBenchmarks_RealLLM(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		t.Skip("LLM API key not set")
	}

	ctx := context.Background()
	model, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-2.0-flash"))
	if err != nil {
		t.Fatalf("failed to create model: %v", err)
	}

	ts := store.NewTextStore()
	host := bridge.NewLangChainHost(model, ts)
	card, _ := os.ReadFile("../assets/dialect_card.md")
	host.DialectCard = string(card)

	m := &LLMModel{host: host}
	
	c := runner.Case{
		ID:        "C1-real",
		Suite:     "C",
		Task:      "Extract credentials from the prompt. Specifically, find the user and pass.",
		PromptRef: "prompts/C1-needle-001.txt",
		Scoring:   runner.ScoringConfig{Mode: "status_ok"},
	}

	t.Logf("Running real LLM benchmark case...")
	res, err := runner.RunCase(ctx, c, m, ".")
	if err != nil {
		t.Fatalf("Case failed: %v", err)
	}

	if !res.Passed {
		t.Errorf("LLM failed the task. Status: %s, Error: %s", res.Status, res.Error)
		for _, e := range res.Output.Errors {
			t.Logf("DSL Error: [%s] %s", e.Code, e.Message)
		}
	}
}
