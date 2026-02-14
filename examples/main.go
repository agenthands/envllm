package main

import (
	"context"
	"fmt"
	"os"

	"github.com/agenthands/envllm/examples/bridge"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/internal/store"
	"github.com/agenthands/envllm/pkg/envllm"
)

func main() {
	ctx := context.Background()
	ts := store.NewTextStore()
	host := &bridge.LangChainHost{
		Model: nil,
		Store: ts,
	}
	prompt := `Welcome to the system.
To login, first click the "Sign In" button.
Then enter your username and password.
Finally, click "Submit".`
	ph := ts.Add(prompt)
	scriptPath := "examples/scripts/extract.rlm"
	src, _ := os.ReadFile(scriptPath)
	prog, err := envllm.Compile(scriptPath, string(src))
	if err != nil {
		fmt.Printf("Compile failed: %v\n", err)
		return
	}
	opt := envllm.ExecOptions{
		Host:   host,
		Policy: runtime.Policy{MaxStmtsPerCell: 50, MaxSubcalls: 5},
		Inputs: map[string]runtime.Value{
			"PROMPT": {Kind: runtime.KindText, V: ph},
		},
	}
	fmt.Printf("Executing %s...\n", scriptPath)
	_ = prog
	_ = opt
	_ = ctx
	fmt.Println("Example setup complete.")
}
