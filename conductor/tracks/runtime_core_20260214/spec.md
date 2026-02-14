# Track Spec: Core Runtime (VM) and TextStore

## Overview
Implement the core execution engine for RLM-Go. This includes managing program state (variables), efficient prompt/text storage, and the deterministic interpreter loop.

## Requirements

1. **TextStore (`internal/store`)**:
    - Manage long prompt content and generated snippets.
    - Support handle-based access to avoid copying large strings.
    - Implement basic "window" and "slice" operations on handles.

2. **Environment & Values (`internal/runtime`)**:
    - Define `Value` type (tagged union/struct for INT, BOOL, TEXT, JSON, SPAN).
    - Implement `Env` for variable storage and lookup.
    - Enforce single-assignment (reassignment check).

3. **Interpreter Loop (`internal/runtime`)**:
    - Implement `Session` to manage the execution lifecycle.
    - Execute `CELL` blocks one statement at a time.
    - Support basic control flow: `SET_FINAL`, `PRINT`, `ASSERT`.
    - Implement budget tracking (statements, bytes).

4. **Ops Dispatch (`internal/ops`)**:
    - Create a registry for operations.
    - Implement the `PRINT` and `SET_FINAL` operations as the first functional "ops".

## Success Criteria
- A simple program with `PRINT` and `SET_FINAL` executes and produces a valid `ExecResult` JSON.
- `TextStore` correctly handles "handles" vs "previews".
- Reassignment of variables in the same session triggers an error.
