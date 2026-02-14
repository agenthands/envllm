# Implementation Plan: Core Parser

This plan follows the TDD workflow: write tests first, then implement, then verify.

## Phase 1: Lexer Implementation
- [x] **Task: Setup Lexer Scaffolding**
    - [x] Define Token types and `Loc` structure in `internal/lex`.
- [x] **Task: Implement Lexer Logic**
    - [x] Write tests for basic tokenization of keywords, idents, and literals.
    - [x] Implement the lexer to pass tests.
    - [x] Verify 90% coverage for `internal/lex`.
- [x] **Task: Conductor - User Manual Verification 'Phase 1: Lexer Implementation' (Protocol in workflow.md)**

## Phase 2: Parser & AST Generation
- [x] **Task: Define AST Structures**
    - [x] Implement Go structs in `internal/ast` matching `schemas/ast.schema.json`.
- [x] **Task: Implement Parser**
    - [x] Write parser tests for `CELL` blocks and `RLMDSL` header.
    - [x] Implement the parser logic using the lexer.
    - [x] Write tests for operation statements with keywords and `INTO`.
    - [x] Implement op statement parsing.
- [x] **Task: Verify Parser Quality**
    - [x] Verify 90% coverage for `internal/parse`.
    - [x] Run native Go fuzzing on parser entry point.
- [x] **Task: Conductor - User Manual Verification 'Phase 2: Parser & AST Generation' (Protocol in workflow.md)**

## Phase 3: Validation & Schemas
- [x] **Task: Implement Schema Validation**
    - [x] Integrate `github.com/santhosh-tekuri/jsonschema/v5`.
    - [x] Write tests to validate generated ASTs against `schemas/ast.schema.json`.
    - [x] Implement validator service in `internal/validate`.
- [x] **Task: Verify Validation Quality**
    - [x] Verify 90% coverage for `internal/validate`.
- [x] **Task: Conductor - User Manual Verification 'Phase 3: Validation & Schemas' (Protocol in workflow.md)**
