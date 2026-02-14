# Tech Stack: RLM-Go

This document defines the official technology stack and engineering standards for the RLM-Go project.

## 1. Core Language & Runtime
- **Primary Language:** Go (Golang)
- **Version Constraint:** Go ≥ 1.18 (required for native fuzzing support and generics).
- **Architecture:** Embedded Interpreter/VM pattern.
- **External Dependencies:** Zero required for core execution; must run fully in-process without a database.

## 2. Data Protocols & Schemas
- **Schema Standard:** JSON Schema Draft 2020-12.
- **Validation (Go):** `github.com/santhosh-tekuri/jsonschema/v5` (chosen for thread-safety and full draft 2020-12 support).
- **Formats:**
  - **DSL Definition:** `assets/ops.json`
  - **Runtime Communication:** JSON-only observations and execution results.
  - **AST:** JSON-encoded tree structure.

## 3. Safety & Security
- **Regex Engine:** Standard Go `regexp` (RE2) to prevent catastrophic backtracking and ensure linear-time execution.
- **Recursion Control:** Managed via the `SUBCALL` op with explicit `DEPTH_COST` and host-enforced recursion limits.
- **Resource Management:** Explicit budgeting for wall-time, statements-per-cell, memory (total bytes), and subcall counts.

## 4. Quality Assurance & Tooling
- **Testing:** `go test` for unit and integration testing.
- **Fuzzing:** Native Go fuzzing targets are mandatory for the parser, lexer, and validator to ensure no panics on malformed input.
- **Linting:** `golangci-lint` for consistent code hygiene and static analysis.
- **Documentation:** Markdown for all specifications, guidelines, and dialect cards.
