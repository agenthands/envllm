# Track Spec: Capability Gating and Regex Operations

## Overview
Enhance the security and expressive power of EnvLLM. This track implements the capability enforcement layer to protect the host from unauthorized operation calls and adds a high-performance regex-based search operation.

## Requirements

1. **Capability Gating (`internal/runtime`)**:
    - Update `Policy` to include an `AllowedCapabilities` map.
    - Modify the `OpDispatcher` interface or the `Session.ExecuteStmt` logic to verify that the operation's required capabilities (from `ops.json`) are explicitly allowed in the session's policy.
    - If a capability is missing, the operation must fail with a `capability_denied` status.

2. **Regex Operation (`internal/ops/pure`)**:
    - **`FIND_REGEX`**: Implement a new operation that searches a `TEXT` handle for a regular expression pattern.
    - Use Go's standard `regexp` library (RE2) for linear-time safety.
    - Support returning the first/last match as a `SPAN` (start, end).
    - Support returning capture groups as a `JSON` object.

3. **Operations Update (`assets/ops.json`)**:
    - Add `FIND_REGEX` to the formal operations registry.
    - Define its signature: `SOURCE` (TEXT), `PATTERN` (TEXT), `MODE` (FIRST/LAST).

## Success Criteria
- A program calling `SUBCALL` fails if the "llm" capability is not enabled in the policy.
- `FIND_REGEX` correctly identifies patterns and returns accurate `SPAN` values.
- 90% test coverage for capability checks and regex matching logic.
