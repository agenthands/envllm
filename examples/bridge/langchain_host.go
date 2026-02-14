package bridge

import (
	"context"
	"fmt"

	"github.com/agenthands/envllm/internal/runtime"
	"github.com/tmc/langchaingo/llms"
)

type LangChainHost struct {
	Model llms.Model
	Store runtime.TextStore
}

func (h *LangChainHost) Subcall(ctx context.Context, req runtime.SubcallRequest) (runtime.SubcallResponse, error) {
	prompt := fmt.Sprintf(`Task: %s
Source Content: %s

Output only valid RLMDSL code.`, req.Task, h.resolveHandle(req.Source))
	completion, err := llms.GenerateFromSinglePrompt(ctx, h.Model, prompt)
	if err != nil {
		return runtime.SubcallResponse{}, err
	}
	return runtime.SubcallResponse{
		Result: runtime.Value{Kind: runtime.KindJSON, V: completion},
	}, nil
}

func (h *LangChainHost) resolveHandle(handle runtime.TextHandle) string {
	text, ok := h.Store.Get(handle)
	if !ok {
		return "[MISSING CONTENT]"
	}
	return text
}

func (h *LangChainHost) RunRLMLoop(ctx context.Context, session *runtime.Session, initialCell string) error {
	return nil
}
