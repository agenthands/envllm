# Implementation Plan: Core Runtime & TextStore

## Phase 1: Values and TextStore
- [x] **Task: Define Typed Values**
    - [x] Implement `Value` and `Kind` in `internal/runtime`.
    - [x] Add JSON marshalling logic for tagged value encoding (`{"kind": "...", "v": ...}`).
- [x] **Task: Implement TextStore**
    - [x] Implement `TextStore` in `internal/store`.
    - [x] Write tests for handle-based access and snippet creation.
    - [x] Verify 90% coverage for `internal/store`.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Values and TextStore' (Protocol in workflow.md)**

## Phase 2: Session and Environment
- [x] **Task: Implement Environment**
    - [x] Implement `Env` for variable storage and single-assignment enforcement.
    - [x] Write tests for variable scope and reassignment errors.
- [x] **Task: Implement Session & Interpreter**
    - [x] Implement `Session` and the `Step()` loop in `internal/runtime`.
    - [x] Implement `SET_FINAL`, `PRINT`, and `ASSERT` logic.
    - [x] Implement statement and memory budget tracking.
- [x] **Task: Verify Runtime Quality**
    - [x] Verify 90% coverage for `internal/runtime`.
    - [x] Run fuzz tests for basic execution patterns (random stmt sequences).
- [x] **Task: Conductor - User Manual Verification 'Phase 2: Session and Environment' (Protocol in workflow.md)**

## Phase 3: Result Encoding
- [x] **Task: Implement ExecResult Generation**
    - [x] Implement result encoding compliant with `schemas/exec_result.schema.json`.
    - [x] Implement `vars_delta` tracking per cell.
- [ ] **Task: Verify Protocol Compliance**
    - [ ] Write tests to ensure `ExecResult` validates against the JSON schema.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Result Encoding' (Protocol in workflow.md)**
