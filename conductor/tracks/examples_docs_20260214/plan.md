# Implementation Plan: Examples & Docs (LangChainGo)

## Phase 1: LangChainGo Bridge
- [ ] **Task: Setup LangChainGo Dependencies**
    - [ ] Run `go get github.com/tmc/langchaingo`.
- [ ] **Task: Implement LangChainGo Host**
    - [ ] Create `examples/bridge/langchain_host.go`.
    - [ ] Implement `Subcall` using a LangChainGo `llms.Model`.
    - [ ] Provide a basic implementation of the observation loop.
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: LangChainGo Bridge' (Protocol in workflow.md)**

## Phase 2: RLM Examples
- [ ] **Task: Create Example Scripts**
    - [ ] Write `examples/scripts/summarize.rlm`.
    - [ ] Write `examples/scripts/extract.rlm`.
- [ ] **Task: Implement Main Example Entry Point**
    - [ ] Create `examples/main.go` that ties the bridge and scripts together.
    - [ ] Verify execution using a local provider (e.g., mock or local LLM).
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: RLM Examples' (Protocol in workflow.md)**

## Phase 3: Comprehensive Documentation
- [ ] **Task: Update README.md**
    - [ ] Add "Getting Started" section.
    - [ ] Add "LangChainGo Integration" section.
- [ ] **Task: Formalize docs/**
    - [ ] Move and clean up `SPEC.md` and `conductor/protocol.md` into `docs/`.
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Comprehensive Documentation' (Protocol in workflow.md)**
