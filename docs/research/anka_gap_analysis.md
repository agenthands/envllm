# Gap Analysis: EnvLLM-DSL vs. Anka

This document analyzes the gaps between our EnvLLM-DSL implementation and the Anka language described in the reference paper ("Anka: A Domain-Specific Language for Reliable LLM Code Generation").

## Core Design Principles (Anka)

Anka's success is attributed to four key principles designed to reduce decision space for LLMs:

1.  **One Canonical Form (No Flexibility):** Every operation has exactly one way to be expressed. No aliases, no optional clauses, no alternative syntax.
2.  **Named Intermediate Results (Explicit State):** Every operation *must* produce a named output via `INTO`. No chaining, no implicit return values.
3.  **Explicit Step Structure (Scaffolding):** Operations are grouped into named `STEP` blocks. This acts as "thought scaffolding" for the model.
4.  **Verbose Keywords (Natural Language Alignment):** Uses English keywords (FILTER, MAP, WHERE) instead of symbols or operators.

## Current Implementation Status (EnvLLM-DSL v0.2.1)

Our implementation has successfully adopted many of these principles, but significant gaps remain in strictness and scaffolding.

### 1. Canonical Form
*   **Status:** Partially Implemented.
*   **Gap:** We still allow some flexibility. For example, `CONCAT_TEXT` vs `OFFSET_ADD` was a recent fix. We need to ensure *every* logical intent maps to exactly one op.
*   **Anka Comparison:** Anka has no equivalent to our `ModeCompat`. It is strict by default. Our migration to `ModeStrict` is the right path, but we must ensure the *linter* enforces the "one way" rule even for logic (e.g., forbidding `CONCAT` for offsets is a good start).

### 2. Named Intermediates (`INTO`)
*   **Status:** **Strong Match.**
*   **Alignment:** We enforce `INTO <var>: <Type>` in v0.2 strict mode. This is exactly aligned with Anka's design.
*   **Improvement:** Anka's paper highlights that this prevents "variable shadowing" (42% of Python errors). We should verify our linter actively checks for shadowing and scope pollution.

### 3. Structural Scaffolding (`STEP` vs `CELL`)
*   **Status:** **Gap Identified.**
*   **Alignment:** We use `CELL <name>:`. Anka uses `STEP <name>:`.
*   **Gap:** Anka's `STEP` structure is designed to guide *sequential* operations within a pipeline (e.g., `STEP filter_large`, `STEP summarize`). Our `CELL` model is derived from the RLM (Recursive Language Model) REPL concept, where a cell might be a single turn.
*   **Action:** We need to confirm if `CELL` provides enough scaffolding. If models struggle with multi-op logic inside a cell (which we saw in Suite C failures), we might need to enforce finer-grained `STEP` blocks or require comments/reasoning lines before ops.

### 4. Verbose Keywords
*   **Status:** **Strong Match.**
*   **Alignment:** `FIND_TEXT`, `WINDOW_TEXT` are verbose.
*   **Gap:** We use `OFFSET_ADD` which is technical. Anka prefers "natural" phrasing like `FILTER ... WHERE ...`.
*   **Action:** Review op names. Is `SLICE_TEXT` clear enough? Maybe `EXTRACT_TEXT FROM ...` is better? (Low priority, but worth noting).

## Missing Features / Divergences

### A. Type System Rigor
*   **Anka:** Uses explicit schema declarations (`TABLE[field: TYPE]`) and types in prompts.
*   **EnvLLM:** We added `OFFSET`, `COST`, `SPAN` recently.
*   **Gap:** Our `JSON` type is too loose. Anka validates schema matches. We just say "it's JSON".
*   **Recommendation:** We need **Typed JSON Schemas** or structured objects. `JSON_GET` failing on "key not found" is a symptom of this. We should allow declaring `STRUCT <Name> { field: Type }` or similar.

### B. Control Flow
*   **Anka:** Supports `IF/ELSE`, `FOR_EACH`, `TRY/ON_ERROR`.
*   **EnvLLM:** We rely on `SUBCALL` (recursion) for control flow. We have no loop or conditional in the DSL itself.
*   **Analysis:** This is a fundamental architectural choice (RLM vs Pipeline). Anka is a *pipeline* language. EnvLLM is a *control* language.
*   **Risk:** If a task requires simple iteration (e.g. "extract all emails"), EnvLLM forces a complex recursive call or a Python-style `FIND_ALL` (which we don't have).
*   **Recommendation:** Consider adding `FOREACH <List> INTO <Item> ... END` if we want to solve Suite C/D tasks more efficiently without full recursion overhead.

### C. Error Handling
*   **Anka:** `TRY/ON_ERROR` blocks.
*   **EnvLLM:** Fail-fast (runtime error stops execution).
*   **Gap:** The LLM cannot "recover" within a cell. It relies on the *next* turn (Repair Loop) to fix things.
*   **Verdict:** This is acceptable for our design (Interactive REPL), whereas Anka targets batch pipelines.

## Conclusion & Next Steps

EnvLLM-DSL v0.2 is very close to Anka's core value proposition (Reliability via Constraints). The biggest gap is **Type Safety for Data Structures**.

**Immediate Action Plan:**
1.  **Strict Shadowing Check:** Ensure linter forbids reusing variable names (Anka Principle 2).
2.  **Schema Validation:** Move beyond generic `JSON` type. Implement `STRUCT` or documented schema expectations for ops like `STATS` and `SUBCALL`.
3.  **Scaffolding:** Experiment with requiring a `# Plan:` comment before statements in a CELL to mimic Anka's "reasoning" benefits.
