package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/agenthands/envllm/bench/runner"
)

type MockModel struct{}

func (m *MockModel) Complete(ctx context.Context, task, prompt string) (string, error) {
	// Simple mapping for testing the runner
	switch task {
	case "Parse basic cell":
		return "CELL test:\n  PRINT SOURCE \"hello\"\n  SET_FINAL SOURCE true\n", nil
	case "Validate strict 2-space indentation":
		return "CELL test:\nPRINT SOURCE \"wrong\"\n", nil 
	case "Enforce step budget":
		return "CELL test:\n  PRINT SOURCE 1\n  PRINT SOURCE 2\n  PRINT SOURCE 3\n", nil
	case "Denied capability access":
		return "CELL test:\n  READ_FILE PATH \"/etc/passwd\" INTO out\n  SET_FINAL SOURCE out\n", nil
	case "Enforce recursion depth limit":
		return "CELL test:\n  SUBCALL SOURCE PROMPT TASK \"recursive\" DEPTH_COST 10 INTO out\n  SET_FINAL SOURCE out\n", nil
	case "Extract credentials JSON":
		return "CELL test:\n  JSON_PARSE SOURCE \"{\\\"user\\\": \\\"admin\\\", \\\"pass\\\": \\\"hunter2\\\"}\" INTO out\n  SET_FINAL SOURCE out\n", nil
	case "Recursive summary of history":
		return "CELL test:\n  FIND_TEXT SOURCE PROMPT NEEDLE \"Chapter 1\" MODE FIRST IGNORE_CASE true INTO p1\n  WINDOW_TEXT SOURCE PROMPT CENTER p1 RADIUS 100 INTO s1\n  SUBCALL SOURCE s1 TASK \"Summarize chapter 1\" DEPTH_COST 1 INTO r1\n  SET_FINAL SOURCE r1\n", nil
	default:
		return "CELL default:\n  SET_FINAL SOURCE null\n", nil
	}
}

func main() {
	fmt.Println("EnvLLM Benchmark Runner v0.1")
	fmt.Println("----------------------------")

	baseDir := "./bench"
	casesDir := filepath.Join(baseDir, "cases")
	
	files, err := os.ReadDir(casesDir)
	if err != nil {
		fmt.Printf("Error reading cases: %v\n", err)
		return
	}

	model := &MockModel{}
	ctx := context.Background()

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}

		fmt.Printf("Running suite: %s\n", f.Name())
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
		if res.Error != "" {
			fmt.Printf("    Error: %s\n", res.Error)
		}
		for _, e := range res.Output.Errors {
			fmt.Printf("    DSL Error: [%s] %s\n", e.Code, e.Message)
		}
	}
}
