# Implementation Plan: Ops Registry & Pure Ops

## Phase 1: Ops Registry & Loading
- [x] **Task: Load Ops Definition**
    - [x] Implement `Table` structure in `internal/ops` to hold `ops.json` data.
    - [x] Write logic to parse `assets/ops.json`.
- [x] **Task: Implement Signature Validation**
    - [x] Implement keyword order and type checking logic.
    - [x] Write tests for validation failures (wrong order, wrong kind).
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: Ops Registry & Loading' (Protocol in workflow.md)**

## Phase 2: Pure Ops Implementation
- [x] **Task: Implement Text Ops**
    - [x] Implement `STATS`, `FIND_TEXT`, and `WINDOW_TEXT`.
    - [x] Write unit tests for each, ensuring edge cases (not found, center out of bounds) are handled.
- [x] **Task: Implement JSON Ops**
    - [x] Implement `JSON_PARSE` and `JSON_GET`.
    - [x] Write unit tests for valid and malformed JSON.
- [x] **Task: Verify Ops Quality**
    - [x] Verify 90% coverage for `internal/ops` and its sub-packages.- [x] **Task: Conductor - User Manual Verification 'Phase 2: Pure Ops Implementation' (Protocol in workflow.md)**

## Phase 3: Runtime Integration
- [x] **Task: Update Session Dispatch**
    - [x] Integrate registry into `internal/runtime/session.go`.
    - [x] Implement `INTO` logic for `OpStmt`.
- [x] **Task: End-to-End Verification**
    - [x] Write an integration test executing a program that finds a string and windows it.
    - [x] Ensure all execution events are recorded in `ExecResult`.
- [x] **Task: Conductor - User Manual Verification 'Phase 3: Runtime Integration' (Protocol in workflow.md)**
