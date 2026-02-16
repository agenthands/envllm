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
Produce a single CELL that solves this specific task and ends with SET_FINAL SOURCE <result>.
Use ONLY valid EnvLLM-DSL 0.1.
`, h.DialectCard, req.Task, h.resolveHandle(req.Source))

	completion, err := llms.GenerateFromSinglePrompt(ctx, h.Model, prompt)
	if err != nil {
		return runtime.SubcallResponse{}, err
	}

	dslCode := h.stripMarkdown(completion)
	
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

		completion, err := llms.GenerateFromSinglePrompt(ctx, h.Model, prompt)
		if err != nil {
			return obs, err
		}

		dslCode := h.stripMarkdown(completion)
		dslCode = h.cleanDSL(dslCode)
		fmt.Printf("--- Turn %d LLM Output ---\n%s\n------------------------\n", i, dslCode)

		prog, err := envllm.Compile(fmt.Sprintf("turn_%d.rlm", i), dslCode, envllm.ModeCompat)
		if err != nil {
			return obs, fmt.Errorf("turn %d: DSL compilation failed: %v\nCode:\n%s", i, err, dslCode)
		}

		for _, cell := range prog.AST.Cells {
			if err := s.ExecuteCell(ctx, cell); err != nil {
				return s.GenerateResult("error", []runtime.Error{{Code: "EXEC_ERROR", Message: err.Error()}}), nil
			}
		}
	}

	return s.GenerateResult("error", []runtime.Error{{Code: "TIMEOUT", Message: "Max turns reached"}}), nil
}

func (h *LangChainHost) cleanDSL(s string) string {
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

func (h *LangChainHost) stripMarkdown(s string) string {
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
- You communicate ONLY by emitting EnvLLM-DSL 0.2 code.
- Declare required capabilities using "REQUIRES capability=..." at the VERY TOP, BEFORE any CELL.
- Then define one "CELL <name>:" block.
- Every statement inside the CELL must be indented by EXACTLY 2 spaces.
- NO line numbers, NO preambles, NO conversational text, NO "NEXT CELL:".
- Use the available "PROMPT" variable to access initial content.
- Use "SET_FINAL SOURCE <expr>" when the task is complete.

EXAMPLE VALID OUTPUT:
REQUIRES capability="fs_read"

CELL find_data:
  READ_FILE PATH "data.txt" INTO content: TEXT
  SET_FINAL SOURCE content

OBSERVATION:
%s

NEXT CELL:`, dialectCard, task, obsJSON)
}
