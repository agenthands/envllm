# EnvLM (Anka-inspired) — Go-native Recursive Language Model DSL + Runtime
Version: 0.1 (spec)
Date: 2026-02-14

## Why this exists
This project implements a **Go-native RLM runtime** plus an **LLM-friendly DSL** that the model writes during an RLM session. It ports **Anka's design principles** (constrained, explicit syntax that minimizes LLM degrees of freedom) while targeting **RLM's REPL + recursion paradigm** (treat the user prompt as an external environment, and let the model programmatically inspect/decompose it and invoke itself on slices).

## Core goals
- **No DB dependency**. Execution is via a small, deterministic **interpreter/VM in Go**.
- **RLM-aligned**: iterative REPL loop; prompt stored as an external variable; code “peeks/slices/searches” the prompt; recursion happens only through a controlled `SUBCALL`.
- **LLM-friendly**: one canonical form per operation, mandatory `INTO` named outputs, explicit `CELL` blocks, verbose keywords, no significant whitespace.
- **Safe by default**: default-deny non-pure capabilities; strict budgets (time/steps/bytes/recursion/concurrency).

## Non-goals (v0.1)
- User-defined functions/macros.
- General loops in the DSL (the host controls iteration by running multiple cells/turns).
- Arbitrary filesystem/network access (optional op packs can be added later, off by default).

---

# 1. RLM session model (what we execute)
An **RLM Session** runs a loop:
1) Host provides the model constant-size metadata about the prompt/environment + brief observation summary.
2) Model emits a DSL cell (code).
3) Host validates, executes it inside a persistent environment (variables + stores).
4) Host returns a compact observation; loop continues until `FINAL` is set or a budget is exhausted.

---

# 2. Language: RLM-DSL (Anka-inspired)
## 2.1 File form
A program is a sequence of CELL blocks. Each CELL contains statements.

Example:
```text
RLMDSL 0.1

CELL plan:
  STATS SOURCE PROMPT INTO stats
  FIND_TEXT SOURCE PROMPT NEEDLE "login" MODE FIRST IGNORE_CASE true INTO pos
  WINDOW_TEXT SOURCE PROMPT CENTER pos RADIUS 1800 INTO snippet

CELL solve:
  SUBCALL SOURCE snippet TASK "Extract login flow as JSON." DEPTH_COST 1 INTO out
  SET_FINAL SOURCE out
```

## 2.2 Values and types
Supported value kinds:
- TEXT, INT, BOOL, JSON, BYTES
- SPAN (start,end)
- LIST[T] where T ∈ {TEXT, INT, BOOL, JSON, SPAN}

TEXT values are handles into a TextStore (rope/chunk store). Ops should avoid copying.

## 2.3 Statements (canonical)
All executable lines have the same shape:

`OP KW1 <expr> KW2 <expr> ... INTO <ident>`

Keyword order is exact per op signature (from ops.json).
Every op produces a named value via INTO except SET_FINAL / PRINT / ASSERT.
Re-assignment is forbidden (default). A name can be written once.

## 2.4 Special statements
- SET_FINAL SOURCE <expr>
- ASSERT COND <expr_bool> MESSAGE "<string>"
- PRINT SOURCE <expr>

---

# 3. Ops system (ops.json)
assets/ops.json defines allowed ops, exact keyword signatures + types, capabilities, and per-op limits.

---

# 4. Runtime architecture (Go)
## 4.1 Packages
- internal/lex: tokenization + locs
- internal/parse: parser -> AST
- internal/ast: node types
- internal/ops: ops.json loader
- internal/validate: signature/type/capability validation
- internal/runtime: session, VM, stores, budgets, trace
- pkg/envllm: public API
- cmd/envllm: CLI (validate/run/repl)

## 4.2 Public API
```go
type Program struct { /* compiled cells + validated */ }

type CompileOptions struct {
    Ops    *ops.Table
    Policy runtime.Policy
}

func Compile(filename string, src []byte, opt CompileOptions) (*Program, error)

type ExecOptions struct {
    Host   runtime.Host
    Inputs map[string]runtime.Value // includes PROMPT
    Policy runtime.Policy
}

func (p *Program) Execute(ctx context.Context, opt ExecOptions) (*runtime.ExecResult, error)
```

## 4.3 Core runtime types (sketch)
```go
type Session struct {
    Env    *Env
    Stores *Stores
    Policy Policy
    Host   Host
    Final  *Value
}

type Host interface {
    Subcall(ctx context.Context, req SubcallRequest) (SubcallResponse, error)
    ReadFile(ctx context.Context, path string) ([]byte, error)   // optional
    Fetch(ctx context.Context, req FetchRequest) (FetchResponse, error) // optional
}

type Policy struct {
    AllowCaps map[string]bool
    MaxCells, MaxStmtsPerCell int
    MaxWallTime time.Duration
    MaxTotalBytes, MaxValueBytes int
    MaxSubcalls, MaxRecursionDepth, MaxParallelSubcalls int
    MaxPrintBytes, MaxObsBytes int
}
```

---

# 5. EBNF (complete v0.1)
```ebnf
program      = header? { cell } ;
header       = "RLMDSL" version NL ;

cell         = "CELL" ident ":" NL { stmt } ;

stmt         = op_stmt NL | set_final NL | assert_stmt NL | print_stmt NL ;

op_stmt      = op_name { kw_arg } "INTO" ident ;
kw_arg       = UIDENT expr ;

set_final    = "SET_FINAL" "SOURCE" expr ;
assert_stmt  = "ASSERT" "COND" expr "MESSAGE" string ;
print_stmt   = "PRINT" "SOURCE" expr ;

expr         = ident | string | int | bool ;
```
