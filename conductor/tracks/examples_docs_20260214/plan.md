# Implementation Plan: Examples & Docs (LangChainGo)

## Phase 1: LangChainGo Bridge
- [x] **Task: Setup LangChainGo Dependencies**
    - [x] Run `go get github.com/tmc/langchaingo`.
- [x] **Task: Implement LangChainGo Host**
    - [x] Create `examples/bridge/langchain_host.go`.
    - [x] Implement `Subcall` using a LangChainGo `llms.Model`.
    - [x] Provide a basic implementation of the observation loop.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: LangChainGo Bridge' (Protocol in workflow.md)**

## Phase 2: RLM Examples
- [x] **Task: Create Example Scripts**
    - [x] Write `examples/scripts/summarize.rlm`.
    - [x] Write `examples/scripts/extract.rlm`.
- [x] **Task: Implement Main Example Entry Point**
    - [x] Create `examples/main.go` that ties the bridge and scripts together.
    - [x] Verify execution using a local provider (e.g., mock or local LLM).
- [x] **Task: Conductor - User Manual Verification 'Phase 2: RLM Examples' (Protocol in workflow.md)**

## Phase 3: Comprehensive Documentation
- [ ] **Task: Update README.md**
    - [ ] Add "Getting Started" section.
    - [ ] Add "LangChainGo Integration" section.
- [ ] **Task: Formalize docs/**
    - [ ] Move and clean up `SPEC.md` and `conductor/protocol.md` into `docs/`.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Comprehensive Documentation' (Protocol in workflow.md)**
