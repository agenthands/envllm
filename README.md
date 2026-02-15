# EnvLLM

**EnvLLM** is a Go-native Recursive Language Model (RLM) runtime and DSL, inspired by the RLM research paper and Anka's reliability principles. It enables LLMs to symbolically interact with long prompts stored in an external environment and recursively invoke themselves for complex tasks.

## Key Features
- **Deterministic Runtime**: A Go interpreter for pure operations.
- **LLM-Friendly DSL**: Verbose, canonical syntax designed to minimize hallucination.
- **Controlled Recursion**: Managed `SUBCALL` with explicit depth-cost and resource budgets.
- **Prompt as Environment**: Treats long prompts as external variables, accessed via tools (search, slice, window).

## Installation
```bash
go get github.com/agenthands/envllm
```

## CLI Usage
Execute scripts or start an interactive REPL:
```bash
# Validate a script
envllm validate script.rlm

# Run a script
envllm run script.rlm --timeout 5s

# Start the REPL
envllm repl
```

## LangChainGo Integration
EnvLLM is designed to be easily embedded. The `examples/bridge` provides a `Host` implementation using [LangChainGo](https://github.com/tmc/langchaingo).

```go
host := &bridge.LangChainHost{
    Model: model, // Any llms.Model (OpenAI, Ollama, etc.)
    Store: ts,
}
```

See `examples/main.go` for a full setup.

## Documentation
- [Language Specification](docs/SPEC.md)
- [Protocol Contract](docs/protocol.md)
- [Product Guidelines](conductor/product-guidelines.md)
