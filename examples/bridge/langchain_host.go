package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/pkg/envllm"
	"github.com/tmc/langchaingo/llms"
)

// LangChainHost implements runtime.Host using LangChainGo.
type LangChainHost struct {
	Model       llms.Model
	Store       runtime.TextStore
	DialectCard string
}

func NewLangChainHost(model llms.Model, store runtime.TextStore) *LangChainHost {
	card, _ := os.ReadFile("assets/syntax_guide.md")
	// If syntax guide missing, fall back to dialect card or empty
	if len(card) == 0 {
		card, _ = os.ReadFile("assets/dialect_card.md")
	}
	return &LangChainHost{
		Model:       model,
		Store:       store,
		DialectCard: string(card),
	}
}

// Subcall implements the recursive call logic.
func (h *LangChainHost) Subcall(ctx context.Context, req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
	prompt := fmt.Sprintf(`%s

TASK: %s
CONTEXT: %s

You are in a recursive SUBCALL. 
Produce a TASK block that solves this specific task.
Example:
TASK sub_task:
  INPUT PROMPT: TEXT
  CELL main:
    ...
    SET_FINAL SOURCE result
  OUTPUT result
`, h.DialectCard, req.Task, h.resolveHandle(req.Source))

	completion, err := llms.GenerateFromSinglePrompt(ctx, h.Model, prompt)
	if err != nil {
		return runtime.SubcallResponse{}, err
	}

	dslCode := h.StripMarkdown(completion)
	
	// Execute the returned DSL in a nested session
	ts := envllm.NewTextStore()
	ph := ts.Add(h.resolveHandle(req.Source))
	
	prog, err := envllm.Compile("subcall.rlm", dslCode, envllm.ModeCompat)
	if err != nil {
		return runtime.SubcallResponse{}, fmt.Errorf("subcall DSL compilation failed: %v\nCode:\n%s", err, dslCode)
	}

	opt := envllm.ExecOptions{
		Policy:    runtime.Policy{MaxStmtsPerCell: 50},
		TextStore: ts,
		Host:      h,
		Inputs:    map[string]runtime.Value{"PROMPT": {Kind: runtime.KindText, V: ph}},
	}

	res, err := prog.Execute(ctx, opt)
	if err != nil {
		return runtime.SubcallResponse{}, err
	}

	if res.Final == nil {
		return runtime.SubcallResponse{}, fmt.Errorf("subcall did not produce a final result")
	}

	return runtime.SubcallResponse{
		Result: *res.Final,
	}, nil
}

func (h *LangChainHost) resolveHandle(handle runtime.TextHandle) string {
	text, ok := h.Store.Get(handle)
	if !ok {
		return "[MISSING CONTENT]"
	}
	return text
}

// RunSession executes the RLM loop until completion or error.
func (h *LangChainHost) RunSession(ctx context.Context, task string, ph runtime.TextHandle, policy runtime.Policy) (runtime.ExecResult, error) {
	s := runtime.NewSession(policy, h.Store)
	s.Host = h

	// Setup dispatcher
	tbl, err := ops.LoadTable("../../assets/ops.json")
	if err != nil {
		tbl, _ = ops.LoadTable("assets/ops.json")
	}
	s.Dispatcher = ops.NewRegistry(tbl)
	
	// Set initial prompt
	if err := s.Env.Define("PROMPT", runtime.Value{Kind: runtime.KindText, V: ph}); err != nil {
		return runtime.ExecResult{}, err
	}

	for i := 0; i < 5; i++ {
		obs := s.GenerateResult("ok", nil)
		if s.Final != nil {
			return obs, nil
		}

		obsJSON, _ := json.MarshalIndent(obs, "", "  ")

		prompt := FormatPrompt(h.DialectCard, task, string(obsJSON))

		// Request JSON structured output to avoid markdown pollution
		// Using langchaingo's options for JSON response format
		var completion string
		
		jsonPrompt := fmt.Sprintf("%s\n\nReturn your response as a JSON object with two fields:\n1. \"reasoning\": string (your internal thoughts)\n2. \"code\": string (the EnvLLM-DSL v0.2 code)", prompt)

		completion, err = llms.GenerateFromSinglePrompt(ctx, h.Model, jsonPrompt, llms.WithJSONMode())
		if err != nil {
			// Fallback if WithJSONMode fails
			completion, err = llms.GenerateFromSinglePrompt(ctx, h.Model, jsonPrompt)
			if err != nil {
				return obs, err
			}
		}

		// Parse JSON response
		var response struct {
			Reasoning string `json:"reasoning"`
			Code      string `json:"code"`
		}
		
		var prog *envllm.Program
		if err := json.Unmarshal([]byte(h.StripMarkdown(completion)), &response); err != nil {
			// Fallback: try to treat the whole completion as raw DSL if JSON parsing fails
			dslCode := h.StripMarkdown(completion)
			dslCode = h.CleanDSL(dslCode)
			fmt.Printf("--- Turn %d LLM Raw Output (JSON Parse Failed) ---\n%s\n------------------------\n", i, dslCode)
			prog, err = envllm.Compile(fmt.Sprintf("turn_%d.rlm", i), dslCode, envllm.ModeCompat)
		} else {
			fmt.Printf("--- Turn %d LLM Reasoning ---\n%s\n------------------------\n", i, response.Reasoning)
			fmt.Printf("--- Turn %d LLM DSL Code ---\n%s\n------------------------\n", i, response.Code)
			prog, err = envllm.Compile(fmt.Sprintf("turn_%d.rlm", i), response.Code, envllm.ModeCompat)
		}

		if err != nil {
			return obs, fmt.Errorf("turn %d: DSL compilation failed: %v", i, err)
		}

		if err := s.ExecuteTask(ctx, prog.AST.Task); err != nil {
			return s.GenerateResult("error", []runtime.Error{{Code: "EXEC_ERROR", Message: err.Error()}}), nil
		}
	}

	return s.GenerateResult("error", []runtime.Error{{Code: "TIMEOUT", Message: "Max turns reached"}}), nil
}

func (h *LangChainHost) CleanDSL(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "NEXT CELL:" || trimmed == "CELL_END" || trimmed == "RLMDSL 0.1" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

func (h *LangChainHost) StripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "```") {
		// Find first occurrence of ``` and last occurrence of ```
		start := strings.Index(s, "```")
		end := strings.LastIndex(s, "```")
		if start != -1 && end != -1 && start != end {
			content := s[start+3 : end]
			// Strip language identifier if present (e.g. ```rlm)
			if strings.HasPrefix(content, "rlm") {
				content = content[3:]
			} else if strings.HasPrefix(content, "text") {
				content = content[4:]
			} else if strings.HasPrefix(content, "json") {
				content = content[4:]
			} else if strings.HasPrefix(content, "envllm") {
				content = content[6:]
			}
			return strings.TrimSpace(content)
		}
	}
	return s
}

func FormatPrompt(dialectCard, task, obsJSON string) string {
	return fmt.Sprintf(`%s

SYSTEM INSTRUCTIONS:
- You are an EnvLLM Agent.
- Your task is: %s
- You communicate ONLY by emitting EnvLLM-DSL 0.2.3 code.
- Your code MUST follow this structure:
  TASK <name>:
    [INPUT <name>: <Type>]
    [REQUIRES capability="..."]
    [CELL <name>: ...]
    [IF <expr>: ...]
    OUTPUT <name>
- Declare required capabilities using "REQUIRES capability=..." at the top of the TASK body.
- One capability per REQUIRES line. DO NOT use commas to separate capabilities.
- Use "SET_FINAL SOURCE <expr>" to set the final result.
- NO line numbers, NO preambles, NO conversational text.
- Use the available "PROMPT" variable to access initial content.

RECOVERY TIPS:
- If you need to do math on an OFFSET (e.g. +1), use OFFSET_ADD. Do NOT use CONCAT.
- If you need to read a struct field (e.g. stats.lines), use GET_FIELD. Do NOT use dot notation.
- If you need to loop, use FOR_EACH with a LIMIT.
- Use EXTRACT_JSON SOURCE PROMPT to automatically find and parse JSON data.
- If the PROMPT contains text and JSON, EXTRACT_JSON is safer than JSON_PARSE.
- CAPABILITY MAPPING:
  - SUBCALL -> capability="llm"
  - READ_FILE -> capability="fs_read"
  - WRITE_FILE -> capability="fs_write"
  - LIST_DIR -> capability="fs_read"

EXAMPLE VALID OUTPUT:
TASK find_data:
  INPUT PROMPT: TEXT
  REQUIRES capability="fs_read"

  CELL read:
    READ_FILE PATH "data.txt" INTO content: TEXT
    SET_FINAL SOURCE content
  
  OUTPUT content

OBSERVATION:
%s

NEXT CELL:`, dialectCard, task, obsJSON)
}
