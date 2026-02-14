# Implementation Plan: Ops Registry & Pure Ops

## Phase 1: Ops Registry & Loading
- [x] **Task: Load Ops Definition**
    - [x] Implement `Table` structure in `internal/ops` to hold `ops.json` data.
    - [x] Write logic to parse `assets/ops.json`.
- [ ] **Task: Implement Signature Validation**
    - [ ] Implement keyword order and type checking logic.
    - [ ] Write tests for validation failures (wrong order, wrong kind).
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: Ops Registry & Loading' (Protocol in workflow.md)**

## Phase 2: Pure Ops Implementation
- [ ] **Task: Implement Text Ops**
    - [ ] Implement `STATS`, `FIND_TEXT`, and `WINDOW_TEXT`.
    - [ ] Write unit tests for each, ensuring edge cases (not found, center out of bounds) are handled.
- [ ] **Task: Implement JSON Ops**
    - [ ] Implement `JSON_PARSE` and `JSON_GET`.
    - [ ] Write unit tests for valid and malformed JSON.
- [ ] **Task: Verify Ops Quality**
    - [ ] Verify 90% coverage for `internal/ops` and its sub-packages.
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: Pure Ops Implementation' (Protocol in workflow.md)**

## Phase 3: Runtime Integration
- [ ] **Task: Update Session Dispatch**
    - [ ] Integrate registry into `internal/runtime/session.go`.
    - [ ] Implement `INTO` logic for `OpStmt`.
- [ ] **Task: End-to-End Verification**
    - [ ] Write an integration test executing a program that finds a string and windows it.
    - [ ] Ensure all execution events are recorded in `ExecResult`.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Runtime Integration' (Protocol in workflow.md)**
