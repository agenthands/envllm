# Track Spec: Core Parser (RLM-DSL 0.1)

## Overview
Implement the foundational syntax processing layer for EnvLLM, converting raw DSL text into a validated AST.

## Requirements
1. **Lexer (`internal/lex`)**:
    - Tokenize `CELL`, `INTO`, `RLMDSL`, and operation names.
    - Support verbose keywords (e.g., `SOURCE`, `NEEDLE`, `MODE`).
    - Handle literals: TEXT (quotes), INT, BOOL, JSON (basic object support).
    - Track line and column locations for every token.

2. **Parser (`internal/parse`)**:
    - Implement a deterministic parser based on the EBNF in `SPEC.md`.
    - Support multiple `CELL` blocks.
    - Enforce mandatory `INTO` for ops.
    - Generate an AST compliant with `schemas/ast.schema.json`.

3. **AST Validation (`internal/validate`)**:
    - Validate that generated ASTs match the JSON schema.
    - Check for basic semantic errors (e.g., duplicate cell names).

4. **Testing**:
    - Golden tests for valid DSL snippets.
    - Error recovery/reporting tests for malformed DSL.
