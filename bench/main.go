package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agenthands/envllm/bench/runner"
	"github.com/agenthands/envllm/examples/bridge"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type MockModel struct{}

func (m *MockModel) Complete(ctx context.Context, task, prompt string) (string, error) {
	switch task {
	case "Parse basic cell":
		return "CELL test:\n  PRINT SOURCE \"hello\"\n  SET_FINAL SOURCE true\n", nil
	case "Validate strict 2-space indentation":
		return "CELL test:\nPRINT SOURCE \"wrong\" INTO out: TEXT\n", nil 
	case "Handle null and negative numbers":
		return `CELL test:
  PRINT SOURCE null
  SET_FINAL SOURCE -42
`, nil
	case "Handle escape sequences in strings":
		return `CELL test:
  SET_FINAL SOURCE "line1\nline2\ttab"
`, nil
	case "Enforce step budget":
		return "CELL test:\n  STATS SOURCE PROMPT INTO stats: STRUCT\n  GET_FIELD SOURCE stats FIELD \"cost\" INTO cost: INT\n  TO_TEXT VALUE cost INTO cost_text: TEXT\n  SET_FINAL SOURCE cost_text\n", nil
	case "Denied capability access":
		return "REQUIRES capability=\"fs_read\"\nCELL test:\n  READ_FILE PATH \"/etc/passwd\" INTO out: TEXT\n  SET_FINAL SOURCE out\n", nil
	case "Enforce recursion depth limit":
		return "REQUIRES capability=\"llm\"\nCELL test:\n  SUBCALL SOURCE PROMPT TASK \"recursive\" DEPTH_COST 10 INTO out: JSON\n  SET_FINAL SOURCE out\n", nil
	case "Mandatory REQUIRES declaration":
		return "CELL test:\n  READ_FILE PATH \"foo\" INTO out: TEXT\n  SET_FINAL SOURCE out\n", nil
	case "Extract credentials JSON":
		return "CELL test:\n  JSON_PARSE SOURCE \"{\\\"user\\\": \\\"admin\\\", \\\"pass\\\": \\\"hunter2\\\"}\" INTO out: JSON\n  SET_FINAL SOURCE out\n", nil
	case "Find error offset":
		return "CELL test:\n  FIND_TEXT SOURCE PROMPT NEEDLE \"ERROR\" MODE FIRST IGNORE_CASE false INTO pos: OFFSET\n  ASSERT COND true MESSAGE \"Found error\"\n  SET_FINAL SOURCE pos\n", nil
	case "Extract config JSON":
		return "CELL test:\n  FIND_REGEX SOURCE PROMPT PATTERN \"\\\\{.*\\\\}\" MODE FIRST INTO span: SPAN\n  GET_SPAN_START SOURCE span INTO start: OFFSET\n  GET_SPAN_END SOURCE span INTO end: OFFSET\n  SLICE_TEXT SOURCE PROMPT START start END end INTO snippet: TEXT\n  JSON_PARSE SOURCE snippet INTO cfg: JSON\n  SET_FINAL SOURCE cfg\n", nil
	case "Recursive summary of history":
		return "REQUIRES capability=\"llm\"\nCELL test:\n  FIND_TEXT SOURCE PROMPT NEEDLE \"Chapter 1\" MODE FIRST IGNORE_CASE true INTO p1: OFFSET\n  WINDOW_TEXT SOURCE PROMPT CENTER p1 RADIUS 100 INTO s1: TEXT\n  SUBCALL SOURCE s1 TASK \"Summarize chapter 1\" DEPTH_COST 1 INTO r1: JSON\n  SET_FINAL SOURCE r1\n", nil
	default:
		return "CELL default:\n  SET_FINAL SOURCE null\n", nil
	}
}

type RealLLMModel struct {
	model       llms.Model
	dialectCard string
}

func (m *RealLLMModel) Complete(ctx context.Context, task, prompt string) (string, error) {
	// For benchmarking, we use a single-turn completion driver
	obsJSON := "{}" 
	llmPrompt := bridge.FormatPrompt(m.dialectCard, task, obsJSON)
	return llms.GenerateFromSinglePrompt(ctx, m.model, llmPrompt)
}

func main() {
	useLLM := flag.Bool("llm", false, "Use a real LLM for benchmarks")
	flag.Parse()

	fmt.Println("EnvLLM Benchmark Runner v0.1")
	fmt.Println("----------------------------")

	baseDir := "./bench"
	casesDir := filepath.Join(baseDir, "cases")
	
	files, err := os.ReadDir(casesDir)
	if err != nil {
		fmt.Printf("Error reading cases: %v\n", err)
		return
	}

	var model runner.Model
	ctx := context.Background()

	if *useLLM {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		if apiKey == "" {
			fmt.Println("Error: GEMINI_API_KEY or GOOGLE_API_KEY not set")
			os.Exit(1)
		}
		// Fallback to default (likely gemini-pro or gemini-1.5-flash)
		m, err := googleai.New(ctx, googleai.WithAPIKey(apiKey))
		if err != nil {
			fmt.Printf("Error creating model: %v\n", err)
			os.Exit(1)
		}
		card, _ := os.ReadFile("assets/dialect_card.md")
		model = &RealLLMModel{model: m, dialectCard: string(card)}
		fmt.Println("Mode: Real LLM (Default)")
	} else {
		model = &MockModel{}
		fmt.Println("Mode: Mock Model")
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}

		fmt.Printf("\nRunning suite: %s\n", f.Name())
		runSuite(ctx, filepath.Join(casesDir, f.Name()), model, baseDir)
	}
}

func runSuite(ctx context.Context, path string, m runner.Model, baseDir string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("  Error opening suite: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var c runner.Case
		if err := json.Unmarshal(scanner.Bytes(), &c); err != nil {
			fmt.Printf("  Error parsing case: %v\n", err)
			continue
		}

		res, err := runner.RunCase(ctx, c, m, baseDir)
		if err != nil {
			fmt.Printf("  [FAILED] Case %s: %v\n", c.ID, err)
			continue
		}

		status := "PASSED"
		if !res.Passed {
			status = "FAILED"
		}
		fmt.Printf("  [%s] Case %s: %s (Status: %s)\n", status, c.ID, c.Task, res.Status)
		if !res.Passed && res.Code != "" {
			fmt.Printf("    Generated Code:\n---\n%s\n---\n", res.Code)
		}
		if res.Error != "" {
			fmt.Printf("    Error: %s\n", res.Error)
		}
		if res.Mismatch != "" {
			fmt.Printf("    Mismatch: %s\n", res.Mismatch)
		}
		for _, e := range res.Output.Errors {
			fmt.Printf("    DSL Error: [%s] %s\n", e.Code, e.Message)
		}
	}
}
