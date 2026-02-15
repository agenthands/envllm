# Track Spec: Ops Registry and Pure Symbolic Operations

## Overview
Implement the dynamic operation system for EnvLLM. This includes loading operation definitions from `assets/ops.json`, validating execution signatures at runtime, and implementing the first set of "pure" operations used for prompt analysis and data manipulation.

## Requirements

1. **Ops Registry (`internal/ops`)**:
    - Load and parse `assets/ops.json`.
    - Provide a lookup mechanism for operation metadata (signature, keywords, result type).
    - Implement runtime signature validation:
        - Exact keyword order.
        - Argument type checking against `Value.Kind`.
        - Result type enforcement.

2. **Pure Operation Implementations (`internal/ops/pure`)**:
    - **`STATS`**: Returns JSON with prompt metadata (length, lines, token estimate).
    - **`FIND_TEXT`**: Implements searching for a "needle" string with `FIRST`/`LAST` modes and `IGNORE_CASE` support.
    - **`WINDOW_TEXT`**: Extracts a snippet from a `TEXT` handle given a center position and radius.
    - **`JSON_PARSE`**: Converts a `TEXT` handle content into a `JSON` value.
    - **`JSON_GET`**: Extracts a sub-value from a `JSON` value given a path string.

3. **Runtime Integration (`internal/runtime`)**:
    - Update `Session.ExecuteStmt` to dispatch `OpStmt` calls through the registry.
    - Handle `INTO` assignment for operation results.
    - Capture execution events for each operation (time taken, keywords used).

## Success Criteria
- The runtime successfully executes a multi-step program using `FIND_TEXT` and `WINDOW_TEXT`.
- Keyword order mismatches or type errors trigger structured `Error` objects in the `ExecResult`.
- 90% test coverage for all new operation logic.
