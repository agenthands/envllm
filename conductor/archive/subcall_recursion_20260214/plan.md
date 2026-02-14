# Implementation Plan: SUBCALL & Host Interface

## Phase 1: Host Interface and Accounting
- [x] **Task: Define Host Interface**
    - [x] Add `Host` interface to `internal/runtime/session.go`.
    - [x] Define `SubcallRequest` and `SubcallResponse` types.
- [x] **Task: Implement Recursion State**
    - [x] Update `Session` struct with `RecursionDepth` and `SubcallCount`.
    - [x] Update `Policy` with `MaxRecursionDepth` and `MaxSubcalls`.
    - [x] Update `GenerateResult` to include recursion stats in `budgets`.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Host Interface and Accounting' (Protocol in workflow.md)**

## Phase 2: SUBCALL Implementation
- [x] **Task: Implement SUBCALL Logic**
    - [x] Implement `SUBCALL` operation in `internal/ops/registry.go` (or a dedicated file).
    - [x] Logic must:
        - Validate `DEPTH_COST` vs remaining budget.
        - Call `s.Host.Subcall`.
        - Increment subcall counters and depth.
- [x] **Task: Test with Mock Host**
    - [x] Create a `MockHost` that returns predefined responses.
    - [x] Write unit tests for successful subcalls and budget exhaustion.
- [x] **Task: Verify Quality**
    - [x] Verify 90% coverage for new subcall and accounting logic.
- [x] **Task: Conductor - User Manual Verification 'Phase 2: SUBCALL Implementation' (Protocol in workflow.md)**

## Phase 3: Integration & E2E
- [x] **Task: E2E Recursive Test**
    - [x] Update `cmd/rlmgo/e2e_test.go` with a recursive program (e.g., a cell that calls itself or another task).
    - [x] Verify that events and budgets are correctly propagated through the result.
- [x] **Task: Conductor - User Manual Verification 'Phase 3: Integration & E2E' (Protocol in workflow.md)**
