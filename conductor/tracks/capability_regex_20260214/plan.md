# Implementation Plan: Capability Gating & Regex

## Phase 1: Capability Enforcement
- [x] **Task: Update Policy and Dispatcher**
    - [x] Add `AllowedCapabilities map[string]bool` to `runtime.Policy`.
    - [x] Update `ops.Registry.Dispatch` to check `Op.Capabilities` against the policy.
- [x] **Task: Verify Gating**
    - [x] Write unit tests for denied capabilities (e.g., calling `SUBCALL` without "llm" access).
    - [x] Ensure the error status is `capability_denied`.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Capability Enforcement' (Protocol in workflow.md)**

## Phase 2: Regex Operations
- [ ] **Task: Update ops.json**
    - [ ] Add `FIND_REGEX` definition to `assets/ops.json`.
- [ ] **Task: Implement Regex Logic**
    - [ ] Implement `FindRegex` in `internal/ops/pure/text.go`.
    - [ ] Add support for `SPAN` result type.
- [ ] **Task: Verify Regex Quality**
    - [ ] Write tests for various patterns, including capture groups and complex matches.
    - [ ] Verify 90% coverage for regex-related code.
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: Regex Operations' (Protocol in workflow.md)**

## Phase 3: Integration
- [ ] **Task: E2E Security & Search Test**
    - [ ] Update `cmd/rlmgo/e2e_test.go` with a test case that mixes capability denials and regex searches.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Integration' (Protocol in workflow.md)**
