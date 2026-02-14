# RLM-Go (Anka-inspired RLM DSL + Runtime Specification)

## Project Overview
This project defines the specification for **RLM-Go**, a Go-native Recursive Language Model (RLM) runtime and its associated Domain Specific Language (DSL). It is inspired by Anka's design principles, focusing on a constrained, explicit syntax that minimizes the degrees of freedom for Large Language Models (LLMs) while enabling them to programmatically inspect, decompose, and recursively invoke themselves on data.

### Core Goals
- **Deterministic Runtime**: A Go-native interpreter/VM with no external database dependencies.
- **LLM-Friendly DSL**: Verbose, canonical syntax with mandatory named outputs (`INTO`) and explicit blocks (`CELL`).
- **Recursive Paradigm**: Prompt-as-environment, allowing the model to peek, slice, and search the prompt, with recursion controlled via `SUBCALL`.
- **Safety**: Default-deny capabilities and strict resource budgets (time, steps, memory, recursion depth).

## Directory Overview
This repository contains the formal specifications, operation definitions, and data schemas for the RLM-Go system.

- **`/assets`**: Contains the core definitions for the DSL's operations and guidelines for LLM interaction.
- **`/schemas`**: Contains JSON schemas for validating the AST and execution outputs of the runtime.
- **Root**: Contains the primary specification document and project metadata.

## Key Files
- **`SPEC.md`**: The primary specification document (v0.1). It outlines the session model, language syntax (EBNF), values/types, and the proposed Go runtime architecture.
- **`assets/ops.json`**: The formal registry of all supported operations (e.g., `STATS`, `FIND_TEXT`, `WINDOW_TEXT`, `SUBCALL`), defining their signatures, required keywords, and capabilities.
- **`assets/dialect_card.md`**: A concise reference designed to be prepended to LLM prompts to ensure the model generates valid RLM-DSL code.
- **`schemas/ast.schema.json`**: A JSON schema for the Abstract Syntax Tree (AST) produced by the parser.
- **`schemas/exec_result.schema.json`**: A JSON schema for the results returned by the runtime after executing a program.

## Usage
The contents of this directory serve as the source of truth for:
1.  **Implementing the Runtime**: Providing the EBNF, operation signatures, and expected AST structures for developers building the Go interpreter.
2.  **Prompt Engineering**: Using the `dialect_card.md` and `ops.json` to guide LLMs in writing valid RLM-DSL programs.
3.  **Validation**: Using the provided JSON schemas to ensure compatibility between different components of the RLM-Go ecosystem.
