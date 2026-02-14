# Implementation Plan: File System & Whitelisting

## Phase 1: Path Security
- [ ] **Task: Update Policy and Path Validation**
    - [ ] Add `AllowedReadPaths` and `AllowedWritePaths` to `runtime.Policy`.
    - [ ] Implement `ValidatePath(path, mode)` in `internal/runtime/session.go`.
- [ ] **Task: Verify Path Security**
    - [ ] Unit tests for absolute paths, relative paths, and traversal attempts (`../`).
    - [ ] Ensure paths outside the whitelist are correctly identified.
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: Path Security' (Protocol in workflow.md)**

## Phase 2: File Operations
- [ ] **Task: Update ops.json**
    - [ ] Add `READ_FILE`, `WRITE_FILE`, and `LIST_DIR` definitions.
- [ ] **Task: Implement File Ops**
    - [ ] Implement the operations in a new `internal/ops/capability/fs.go` file.
    - [ ] Operations must call `s.ValidatePath` before performing any I/O.
- [ ] **Task: Verify File Ops Quality**
    - [ ] Write unit tests for successful I/O and various failure modes (file not found, permission denied, out of whitelist).
    - [ ] Verify 90% coverage for FS code.
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: File Operations' (Protocol in workflow.md)**

## Phase 3: Integration
- [ ] **Task: E2E FS Test**
    - [ ] Update `cmd/rlmgo/e2e_test.go` with a test case that writes a snippet to a file and then reads it back.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Integration' (Protocol in workflow.md)**
