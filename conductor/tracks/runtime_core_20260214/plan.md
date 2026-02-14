# Implementation Plan: Core Runtime & TextStore

## Phase 1: Values and TextStore
- [x] **Task: Define Typed Values**
    - [x] Implement `Value` and `Kind` in `internal/runtime`.
    - [x] Add JSON marshalling logic for tagged value encoding (`{"kind": "...", "v": ...}`).
- [ ] **Task: Implement TextStore**
    - [ ] Implement `TextStore` in `internal/store`.
    - [ ] Write tests for handle-based access and snippet creation.
    - [ ] Verify 90% coverage for `internal/store`.
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: Values and TextStore' (Protocol in workflow.md)**

## Phase 2: Session and Environment
- [ ] **Task: Implement Environment**
    - [ ] Implement `Env` for variable storage and single-assignment enforcement.
    - [ ] Write tests for variable scope and reassignment errors.
- [ ] **Task: Implement Session & Interpreter**
    - [ ] Implement `Session` and the `Step()` loop in `internal/runtime`.
    - [ ] Implement `SET_FINAL`, `PRINT`, and `ASSERT` logic.
    - [ ] Implement statement and memory budget tracking.
- [ ] **Task: Verify Runtime Quality**
    - [ ] Verify 90% coverage for `internal/runtime`.
    - [ ] Run fuzz tests for basic execution patterns (random stmt sequences).
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: Session and Environment' (Protocol in workflow.md)**

## Phase 3: Result Encoding
- [ ] **Task: Implement ExecResult Generation**
    - [ ] Implement result encoding compliant with `schemas/exec_result.schema.json`.
    - [ ] Implement `vars_delta` tracking per cell.
- [ ] **Task: Verify Protocol Compliance**
    - [ ] Write tests to ensure `ExecResult` validates against the JSON schema.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Result Encoding' (Protocol in workflow.md)**
