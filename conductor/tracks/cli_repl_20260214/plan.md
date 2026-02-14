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
- [ ] **Task: Implement Run Subcommand**
    - [ ] Logic to read file, compile, and execute.
    - [ ] Flag parsing for budgets.
- [ ] **Task: Implement Validate Subcommand**
    - [ ] Logic to parse and validate without execution.
- [ ] **Task: Test CLI**
    - [ ] Integration tests for CLI subcommands.
- [ ] **Task: Verify CLI Quality**
    - [ ] Ensure 80% coverage for CLI/glue code (as per workflow).
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: CLI Features' (Protocol in workflow.md)**

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
