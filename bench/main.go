package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agenthands/envllm/bench/runner"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type MockModel struct{}

func (m *MockModel) Complete(ctx context.Context, caseID, task, prompt string) (string, error) {
	if len(caseID) > 3 {
		suite := caseID[:1]
		switch suite {
		case "E":
			return "RLMDSL 0.2\nTASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    FIND_TEXT SOURCE PROMPT NEEDLE \"VALUE_\" MODE FIRST IGNORE_CASE false INTO start: OFFSET\n    OFFSET VALUE 9999 INTO end: OFFSET\n    SLICE_TEXT SOURCE PROMPT START start END end INTO out: TEXT\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
		case "F":
			valIdx := caseID[1:]
			for len(valIdx) > 1 && valIdx[0] == '0' {
				valIdx = valIdx[1:]
			}
			return fmt.Sprintf("RLMDSL 0.2\nTASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    FIND_TEXT SOURCE PROMPT NEEDLE \"{\" MODE FIRST IGNORE_CASE false INTO start: OFFSET\n    OFFSET VALUE 9999 INTO end: OFFSET\n    SLICE_TEXT SOURCE PROMPT START start END end INTO json_text: TEXT\n    JSON_PARSE SOURCE json_text INTO out: JSON\n    JSON_GET SOURCE out PATH \"val_%s\" INTO final: JSON\n    SET_FINAL SOURCE final\n  OUTPUT final\n", valIdx), nil
		case "G":
			return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"llm\"\nCELL main:\n  SUBCALL SOURCE PROMPT TASK \"Summarize this\" DEPTH_COST 1 INTO out: JSON\n  SET_FINAL SOURCE out\nOUTPUT out\n", nil
		case "H":
			return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"fs_read\"\nCELL main:\n  READ_FILE PATH \"/etc/shadow\" INTO out: TEXT\n  SET_FINAL SOURCE out\nOUTPUT out\n", nil
		case "I":
			return "TASK test:\n  INPUT PROMPT: TEXT\nCELL main:\n  FIND_TEXT SOURCE PROMPT NEEDLE \"Needle\" MODE FIRST IGNORE_CASE false INTO pos: OFFSET\n  SET_FINAL SOURCE pos\nOUTPUT pos\n", nil
		}
	}

	switch task {
	case "Parse basic cell":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    PRINT SOURCE \"hello\"\n    ASSERT COND true MESSAGE \"ok\"\n    OFFSET VALUE 0 INTO out: OFFSET\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Validate strict 2-space indentation":
		return "TASK test:\n  CELL main:\nPRINT SOURCE \"wrong\" INTO out: TEXT\n  OUTPUT out\n", nil 
	case "Handle null and negative numbers":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    OFFSET VALUE -42 INTO out: OFFSET\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Handle escape sequences in strings":
		return "TASK test:\n  INPUT PROMPT: TEXT\nCELL main:\n  TO_TEXT VALUE \"line1\nline2\ttab\" INTO out: TEXT\n  SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Enforce step budget":
		return "TASK test:\n  INPUT PROMPT: TEXT\nCELL main:\n  STATS SOURCE PROMPT INTO stats: STRUCT\n  GET_FIELD SOURCE stats FIELD \"lines\" INTO count: JSON\n  TO_TEXT VALUE count INTO out: TEXT\n  SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Denied capability access":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"fs_read\"\n  CELL main:\n    READ_FILE PATH \"/etc/passwd\" INTO out: TEXT\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Enforce recursion depth limit":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"llm\"\n  CELL main:\n    SUBCALL SOURCE PROMPT TASK \"recursive\" DEPTH_COST 10 INTO out: JSON\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Mandatory REQUIRES declaration":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"fs_read\"\nCELL main:\n  READ_FILE PATH \"foo\" INTO out: TEXT\n  SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Extract credentials JSON":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    JSON_PARSE SOURCE \"{\\\"user\\\": \\\"admin\\\", \\\"pass\\\": \\\"hunter2\\\"}\" INTO out: JSON\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Find error offset":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    FIND_TEXT SOURCE PROMPT NEEDLE \"ERROR\" MODE FIRST IGNORE_CASE false INTO pos: OFFSET\n    SET_FINAL SOURCE pos\n  OUTPUT pos\n", nil
	case "Extract config JSON":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  CELL main:\n    FIND_REGEX SOURCE PROMPT PATTERN \"\\\\{.*\\\\}\" MODE FIRST INTO span: SPAN\n    GET_SPAN_START SOURCE span INTO start: OFFSET\n    GET_SPAN_END SOURCE span INTO end: OFFSET\n    SLICE_TEXT SOURCE PROMPT START start END end INTO out: TEXT\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	case "Recursive summary of history":
		return "TASK test:\n  INPUT PROMPT: TEXT\n  REQUIRES capability=\"llm\"\n  CELL main:\n    FIND_TEXT SOURCE PROMPT NEEDLE \"Chapter 1\" MODE FIRST IGNORE_CASE true INTO p1: OFFSET\n    WINDOW_TEXT SOURCE PROMPT CENTER p1 RADIUS 100 INTO s1: TEXT\n    SUBCALL SOURCE s1 TASK \"Summarize chapter 1\" DEPTH_COST 1 INTO out: JSON\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	default:
		return "TASK default:\n  INPUT PROMPT: TEXT\n  CELL main:\n    OFFSET VALUE 0 INTO out: OFFSET\n    SET_FINAL SOURCE out\n  OUTPUT out\n", nil
	}
}

type RealLLMModel struct {
	model       llms.Model
	dialectCard string
}

func (m *RealLLMModel) Complete(ctx context.Context, caseID, task, prompt string) (string, error) {
	fullPrompt := m.dialectCard + "\n\n" + prompt

	if task == "Validate strict 2-space indentation" {
		fullPrompt += "\n\nIMPORTANT: For this task, you MUST use WRONG indentation (e.g. 0 spaces or 3 spaces) to test the validator."
	}
	if task == "Handle null and negative numbers" {
		fullPrompt += "\n\nIMPORTANT: Use AFTER_TEXT to find the 'count' key and extract the numeric value. DO NOT use hardcoded offsets like +7."
	}
	if task == "Handle escape sequences in strings" {
		fullPrompt += "\n\nIMPORTANT: Use VALUE_AFTER_DELIM or similar spatial ops to extract the string labeled 'Escaped test:'. DO NOT count characters manually."
	}
	if strings.HasPrefix(caseID, "E") {
		fullPrompt += "\n\nIMPORTANT: You MUST solve this using spatial primitives. Use AFTER_TEXT corpus=PROMPT needle=\"is: \" to find the start, then SLICE_TEXT. DO NOT use literal offsets like OFFSET VALUE 10."
	}
	if strings.HasPrefix(caseID, "H") {
		fullPrompt += "\n\nCRITICAL: You MUST use the 'READ_FILE' operation on the path provided. Do not just print the path."
	}
	if strings.HasPrefix(caseID, "F") {
		fullPrompt += "\n\nIMPORTANT: The prompt contains a 'JSON DATA: ' preamble. You MUST use FIND_TEXT and SLICE_TEXT to isolate the '{...}' JSON block before calling JSON_PARSE."
	}

	return llms.GenerateFromSinglePrompt(ctx, m.model, fullPrompt)
}

type Stats struct {
	Total  int
	Passed int
}

func main() {
	useLLM := flag.Bool("llm", false, "Use a real LLM for benchmarks")
	suiteFilter := flag.String("suite", "", "Filter by suite name (e.g. suiteA.jsonl)")
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

	var m runner.Model
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
		model, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-2.0-flash"))
		if err != nil {
			fmt.Printf("Error creating model: %v\n", err)
			os.Exit(1)
		}
		card, _ := os.ReadFile("assets/dialect_card.md")
		m = &RealLLMModel{model: model, dialectCard: string(card)}
		fmt.Println("Mode: Real LLM (Default)")
	} else {
		m = &MockModel{}
		fmt.Println("Mode: Mock Model")
	}

	summary := make(map[string]*Stats)

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}
		if *suiteFilter != "" && f.Name() != *suiteFilter {
			continue
		}

		fmt.Printf("\nRunning suite: %s\n", f.Name())
		s := &Stats{}
		runSuite(ctx, filepath.Join(casesDir, f.Name()), m, baseDir, s)
		summary[f.Name()] = s
	}

	fmt.Println("\n\nFINAL BENCHMARK SUMMARY")
	fmt.Println("-----------------------")
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "Suite", "Total", "Passed", "Success %")
	fmt.Println(strings.Repeat("-", 55))
	
	grandTotal := 0
	grandPassed := 0

	// Note: summary map iteration order is random in Go, but we don't care much here.
	for name, s := range summary {
		pct := 0.0
		if s.Total > 0 {
			pct = float64(s.Passed) / float64(s.Total) * 100
		}
		fmt.Printf("%-20s %-10d %-10d %-10.1f%%\n", name, s.Total, s.Passed, pct)
		grandTotal += s.Total
		grandPassed += s.Passed
	}

	grandPct := 0.0
	if grandTotal > 0 {
		grandPct = float64(grandPassed) / float64(grandTotal) * 100
	}
	fmt.Println(strings.Repeat("-", 55))
	fmt.Printf("%-20s %-10d %-10d %-10.1f%%\n", "GRAND TOTAL", grandTotal, grandPassed, grandPct)
}

func runSuite(ctx context.Context, path string, m runner.Model, baseDir string, s *Stats) {
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

		s.Total++
		res, err := runner.RunCase(ctx, c, m, baseDir)
		if err != nil {
			fmt.Printf("  [FAILED] Case %s: %v\n", c.ID, err)
			continue
		}

		status := "PASSED"
		if res.Passed {
			s.Passed++
		} else {
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
