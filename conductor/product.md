# Initial Concept
Creating a **Go-native Recursive Language Model (RLM) runtime** and **LLM-friendly DSL** for recursive decomposition and long-context interaction, inspired by the RLM paradigm (prompt-as-environment + recursive self-calls) and Anka’s DSL design principles for reliable LLM code generation (canonical forms, explicit intermediates, step scaffolding, verbose keywords).

# Product Definition
EnvLLM is an embedded execution environment and constrained DSL that enables an LLM to **symbolically interact with a prompt stored in an external environment** (search/slice/window/parse), and to **recursively invoke itself** via a controlled `SUBCALL`, while keeping execution predictable and resource-bounded. Pure operations are deterministic; LLM subcalls are explicitly budgeted and audited.

## Core Goals
- **High LLM Legibility:** Strict canonical forms, verbose keywords, explicit intermediate naming (`INTO`), and explicit cell structure to reduce generation errors.
- **Go-Native Runtime:** Deterministic interpreter/VM for pure ops, implemented in Go, embeddable in applications.
- **Safe Recursion:** Managed recursion via `SUBCALL` with explicit depth cost and budgets (subcalls, depth, time, bytes).
- **No External Database Required:** Runs fully in-process; no DB/state dependency is required to execute programs.

## Key Features
- **CELL-based Structure:** Programs execute as discrete cells (REPL-like), enabling granular observation and controlled progress.
- **Mandatory Named Outputs:** Every operation produces `INTO <ident>` to keep data flow explicit and debuggable.
- **Prompt as Environment:** The prompt is stored outside the model and accessed via ops (slice/search/window), rather than being loaded directly into the model context.
- **Strict Budgets:** Built-in limits for steps, wall-time, memory/bytes, and recursion/subcalls; capability gating for any non-pure operations.

## Non-goals
- **Not a new neural model:** It is a runtime and scaffolding layer, not a weights/training project.
- **No general-purpose language features:** No arbitrary loops or user-defined functions unless explicitly added as controlled ops.
- **No FS/Network by default:** All non-pure capabilities are capability-gated and disabled by default.
