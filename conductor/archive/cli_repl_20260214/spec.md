# Track Spec: CLI and REPL Interface

## Overview
Implement the user-facing interface for RLM-Go. This includes a robust CLI for executing files and an interactive REPL for iterative session development.

## Requirements

1. **Public API (`pkg/rlmgo`)**:
    - Provide a stable entry point for host applications to use RLM-Go.
    - Export `Program`, `Compile`, `Execute`, and `Session` types.
    - Simplify setup by handling default registries and stores.

2. **CLI (`cmd/rlmgo`)**:
    - Support subcommands: `run`, `repl`, `validate`.
    - `run`: Executes a `.rlm` file.
    - `validate`: Checks syntax and operation signatures without executing.
    - Handle flags for budgets (e.g., `--max-depth`, `--timeout`).

3. **REPL (`internal/repl`)**:
    - Implement a Read-Eval-Print Loop.
    - Allow users to enter `CELL` blocks interactively.
    - Maintain a persistent `Session` across entries.
    - Format and print `ExecResult` observations clearly.

4. **User Experience**:
    - Clear error reporting with line/column pointers.
    - Helpful usage/help text.

## Success Criteria
- `rlmgo run script.rlm` correctly executes and outputs the result.
- `rlmgo repl` allows defining a variable in one turn and using it in the next.
- Syntax errors in the CLI provide meaningful feedback.
