# Implementation Plan: CLI & REPL Interface

## Phase 1: Public API & Scaffolding
- [x] **Task: Implement pkg/rlmgo**
    - [x] Create `pkg/rlmgo/rlmgo.go`.
    - [x] Wrap internal components (parse, ops, runtime) into a clean API.
    - [x] Write unit tests for the public API.
- [x] **Task: Setup CLI Structure**
    - [x] Create `cmd/rlmgo/main.go` with basic subcommand routing (run, repl, validate).
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Public API & Scaffolding' (Protocol in workflow.md)**

## Phase 2: CLI Features
- [x] **Task: Implement Run Subcommand**
    - [x] Logic to read file, compile, and execute.
    - [x] Flag parsing for budgets.
- [x] **Task: Implement Validate Subcommand**
    - [x] Logic to parse and validate without execution.
- [x] **Task: Test CLI**
    - [x] Integration tests for CLI subcommands.
- [x] **Task: Verify CLI Quality**
    - [x] Ensure 80% coverage for CLI/glue code (as per workflow).
- [x] **Task: Conductor - User Manual Verification 'Phase 2: CLI Features' (Protocol in workflow.md)**

## Phase 3: REPL Implementation
- [ ] **Task: Implement internal/repl**
    - [ ] Interactive loop using `bufio.Scanner` or similar.
    - [ ] Session persistence logic.
- [ ] **Task: REPL UX & Formatting**
    - [ ] Colorized or formatted output for `ExecResult`.
    - [ ] Multi-line cell support.
- [ ] **Task: Final Verification**
    - [ ] Manual walkthrough of REPL features.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: REPL Implementation' (Protocol in workflow.md)**
