package main

import (
	"context"
	"fmt"
	"os"

	"github.com/agenthands/rlm-go/examples/bridge"
	"github.com/agenthands/rlm-go/internal/runtime"
	"github.com/agenthands/rlm-go/internal/store"
	"github.com/agenthands/rlm-go/pkg/rlmgo"
)

func main() {
	// 1. Initialize dependencies
	ctx := context.Background()
	ts := store.NewTextStore()
	
	// Note: In a real run, you would use a real LangChainGo model provider.
	// For this example, we show the setup logic.
	// model, _ := openai.New() 
	
	host := &bridge.LangChainHost{
		Model: nil, // Placeholder: requires API key / setup
		Store: ts,
	}

	// 2. Prepare sample input
	prompt := `Welcome to the system. 
To login, first click the "Sign In" button. 
Then enter your username and password.
Finally, click "Submit".`
	ph := ts.Add(prompt)

	// 3. Compile and Execute script
	scriptPath := "examples/scripts/extract.rlm"
	src, _ := os.ReadFile(scriptPath)
	prog, err := rlmgo.Compile(scriptPath, string(src))
	if err != nil {
		fmt.Printf("Compile failed: %v
", err)
		return
	}

	opt := rlmgo.ExecOptions{
		Host:   host,
		Policy: runtime.Policy{MaxStmtsPerCell: 50, MaxSubcalls: 5},
		Inputs: map[string]runtime.Value{
			"PROMPT": {Kind: runtime.KindText, V: ph},
		},
	}

	fmt.Printf("Executing %s...
", scriptPath)
	// Note: This will fail currently because host.Model is nil.
	// This main file serves as documentation of how to use the bridge.
	_ = prog
	_ = opt
	_ = ctx
	
	fmt.Println("Example setup complete. (Set up an LLM provider in main.go to run for real)")
}
