# Project Workflow: RLM-Go

This document defines the development workflow, testing requirements, and communication standards for the RLM-Go project.

## 1. Development Cycle
- **Task-Based Development:** Work is organized into high-level "tracks," which are further decomposed into discrete tasks in a `plan.md`.
- **Commit Frequency:** Commit early and often. At a minimum, a commit MUST be made after the completion of each task.
- **Commit Quality:** Every commit must build successfully and pass all unit tests. Commit messages must be meaningful and descriptive.

## 2. Testing & Quality Gates
- **Test-Driven Development (TDD):** Highly recommended for the core runtime, parser, and validator.
- **Code Coverage Thresholds:**
    - **Overall Project:** ≥ 80%
    - **Core Packages (`internal/runtime`, `internal/validate`, `internal/parse`):** ≥ 90%
    - **Excluded:** CLI/glue code in `cmd/` is excluded from the mandatory threshold.
- **Mandatory Quality Checks:**
    - `go test ./...` (including fuzzing for inputs).
    - `golangci-lint run`.
    - JSON Schema validation for all protocol-related assets.

## 3. Communication & Documentation
- **Task Summaries:** Use **Pull Request (PR) descriptions** as the primary channel for task and feature summaries. Summaries should include "what," "why," and "how it was tested."
- **Architectural Decisions:** Significant architectural changes or decisions must be documented as **Architectural Decision Records (ADRs)** in `docs/decisions/`.
- **Git Notes:** Not used by default for task summaries to ensure maximum visibility across team members and CI/CD platforms.

## 4. Phase Completion Protocol
At the end of each development Phase (as defined in `plan.md`), the following "Phase Completion Verification and Checkpointing Protocol" must be followed:
1. **Automated Verification:** All tests and linting must pass at the required thresholds.
2. **User Manual Verification:** The developer must walk through the phase's deliverables to ensure they meet the `spec.md`.
3. **Checkpoint:** A summary of the phase's progress is added to the PR or a dedicated ADR if significant changes occurred.
