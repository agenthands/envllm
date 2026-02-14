# Track Spec: SUBCALL and Host Interface for Controlled Recursion

## Overview
Implement the recursive capabilities of RLM-Go. This involves defining how the runtime communicates back to the LLM (the Host) via `SUBCALL` and enforcing strict limits on recursion depth and resource consumption during those calls.

## Requirements

1. **Host Interface (`internal/runtime`)**:
    - Define the `Host` interface with a `Subcall` method.
    - `Subcall` must take a `SubcallRequest` (Source text, Task description, Budgets) and return a `SubcallResponse` (ExecResult/Value).

2. **Recursion Accounting (`internal/runtime`)**:
    - Add `RecursionDepth` and `SubcallCount` to the `Session` state.
    - Implement depth-cost logic: each `SUBCALL` consumes a portion of the `recursion_depth` budget.
    - Enforce `MaxRecursionDepth` and `MaxSubcalls` limits from the `Policy`.

3. **SUBCALL Operation (`internal/ops`)**:
    - Implement the `SUBCALL` logic in the operation registry.
    - Map `SOURCE`, `TASK`, and `DEPTH_COST` keywords to the `Host.Subcall` invocation.
    - Ensure `DEPTH_COST` is subtracted from the current session's remaining depth before initiating the call.

4. **Safety & Budgeting**:
    - Prevent infinite recursion by validating depth limits before every subcall.
    - Capture "subcall" events in the `ExecResult` trace, including time and depth cost.

## Success Criteria
- A test using a `MockHost` can execute a `SUBCALL` and receive a result.
- Reaching the `MaxRecursionDepth` or `MaxSubcalls` limit triggers a `budget_exceeded` error.
- All recursive metadata is correctly reported in the final `ExecResult` JSON.
