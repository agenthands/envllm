# Implementation Plan: File System & Whitelisting

## Phase 1: Path Security
- [x] **Task: Update Policy and Path Validation**
    - [x] Add `AllowedReadPaths` and `AllowedWritePaths` to `runtime.Policy`.
    - [x] Implement `ValidatePath(path, mode)` in `internal/runtime/session.go`.
- [x] **Task: Verify Path Security**
    - [x] Unit tests for absolute paths, relative paths, and traversal attempts (`../`).
    - [x] Ensure paths outside the whitelist are correctly identified.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Path Security' (Protocol in workflow.md)**

## Phase 2: File Operations
- [x] **Task: Update ops.json**
    - [x] Add `READ_FILE`, `WRITE_FILE`, and `LIST_DIR` definitions.
- [x] **Task: Implement File Ops**
    - [x] Implement the operations in a new `internal/ops/capability/fs.go` file.
    - [x] Operations must call `s.ValidatePath` before performing any I/O.
- [x] **Task: Verify File Ops Quality**
    - [x] Write unit tests for successful I/O and various failure modes (file not found, permission denied, out of whitelist).
    - [x] Verify 90% coverage for FS code.
- [x] **Task: Conductor - User Manual Verification 'Phase 2: File Operations' (Protocol in workflow.md)**

## Phase 3: Integration
- [x] **Task: E2E FS Test**
    - [x] Update `cmd/envllm/e2e_test.go` with a test case that writes a snippet to a file and then reads it back.
- [x] **Task: Conductor - User Manual Verification 'Phase 3: Integration' (Protocol in workflow.md)**
