# EnvLLM Platform Guide

EnvLLM is a robust, modular platform for reliable LLM code generation. It uses a specialized Domain Specific Language (DSL) called **EnvLLM-DSL** (v0.2) to ensure deterministic execution, safety, and verifiability of AI-generated workflows.

## Core Concepts

### 1. The Language (EnvLLM-DSL)
EnvLLM-DSL is designed to be "one way to do it." It minimizes the freedom an LLM has to make mistakes by enforcing a strict, canonical structure.

- **CELLs**: The unit of execution. Like a Jupyter cell or a function body.
- **Canonical Ops**: Every operation has exactly one spelling, one keyword order, and mandatory named outputs.
- **Types**: Strongly typed variables (`TEXT`, `INT`, `JSON`, `SPAN`, `BOOL`).
- **Strict Indentation**: Mandatory 2-space indentation.

### 2. The Runtime
A deterministic Go interpreter that executes the DSL safely. It features:
- **Capability Gating**: Operations (like file access or web navigation) must be explicitly allowed by a `Policy`.
- **Resource Budgets**: Strict limits on steps, memory, and wall-time.
- **Recursion Control**: Managed `SUBCALL` logic to prevent infinite AI loops.

### 3. The Extension Framework
EnvLLM is domain-agnostic. Features are added via **Modules**:
- **Core Module**: Pure logic (Search, Regex, JSON parsing).
- **FS Module**: Filesystem operations.
- **Web Module**: Browser automation (Navigate, Click, Type).

## Getting Started

### CLI Usage

```bash
# Run a script
envllm run script.rlm

# Format a script to canonical v0.2
envllm fmt script.rlm

# Check for errors without running
envllm check script.rlm

# Migrate an older v0.1 script
envllm migrate old.rlm > new.rlm
```

### Writing a Module

To add a new domain (e.g., SQL Database access), implement the `Module` interface:

```go
type SQLModule struct{}

func (m *SQLModule) ID() string { return "sql" }

func (m *SQLModule) Operations() []ops.Op {
    return []ops.Op{
        {Name: "QUERY", Capabilities: []string{"sql.read"}, ...},
    }
}

func (m *SQLModule) Handlers() map[string]ops.OpImplementation {
    // Implementation logic
}
```

Then register it in your application:
```go
reg := ops.NewRegistry(table)
reg.RegisterModule(&SQLModule{})
```

## Reliability Features

### Repair Loop
The `Harness` component wraps the LLM interaction. If the LLM generates invalid code (syntax error, wrong type), the Harness:
1. Catches the error.
2. Formats a structured error message.
3. Feeds it back to the LLM for a "Step-Local Repair".
4. Prevents "Drift" (renaming variables or changing past steps).

### Migration
The system includes a migration tool to automatically upgrade loose v0.1 scripts to the strict v0.2 standard, ensuring backward compatibility while moving towards higher reliability.
