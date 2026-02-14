# product-guidelines.md

## Product Guidelines: RLM-Go

These guidelines ensure RLM-Go remains aligned with:

* **Recursive Language Models (RLM):** long prompt treated as an external environment, with the model interacting via a REPL and recursive calls over snippets ([arXiv][1])
* **Anka-inspired reliability design:** constrained, explicit, canonical syntax to reduce LLM generation errors ([arXiv][2])

---

## 1) LLM Interaction Principles

RLM-Go is a **machine-to-machine interface**: the LLM is a controller; the host is a deterministic execution environment for **pure ops** plus a budgeted `SUBCALL`.

### 1.1 Observation Format (Host → LLM)

**Format:** strict JSON-only (no prose outside JSON).
**Tone:** minimalist and data-driven.

**Rationale:** low-entropy, canonical feedback reduces interpretation surface area and drift across many recursive iterations ([arXiv][2])

#### 1.1.1 Observation JSON contract (required keys)

Every observation MUST be a single JSON object with these required keys:

```json
{
  "schema_version": "obs-0.1",
  "cell": { "name": "plan", "index": 0 },
  "status": "ok",
  "vars_delta": { },
  "result": null,
  "final": null,
  "budgets": { },
  "events": [ ],
  "errors": [ ],
  "truncated": { "obs": false, "prints": false, "previews": false }
}
```

* `schema_version`: string, pinned (e.g., `obs-0.1`). **Never** omit.
* `cell`: `{name,index}` identifies executed cell.
* `status`: enum: `"ok" | "error" | "budget_exceeded" | "capability_denied"`.
* `vars_delta`: only variables **created/updated** in this cell (see §1.1.2).
* `result`: optional primary value for the cell (often `null`); do not duplicate large values already in `vars_delta`.
* `final`: `null` or the final typed value if `SET_FINAL` occurred.
* `budgets`: remaining + consumed counters (see §1.1.3).
* `events`: structured trace events (see §1.1.4).
* `errors`: list of structured errors (empty on success).
* `truncated`: booleans indicating truncation happened anywhere.

> JSON Schema MUST be provided in-repo and validated in tests. Use JSON Schema draft 2020-12 for schema versioning and tooling compatibility. ([json-schema.org][3])

#### 1.1.2 Typed value encoding (canonical)

All values in `vars_delta`, `result`, `final` MUST use a tagged encoding:

```json
{ "kind": "INT", "v": 450 }
{ "kind": "BOOL", "v": true }
{ "kind": "JSON", "v": { "a": 1 } }
{ "kind": "SPAN", "v": { "start": 10, "end": 20 } }
```

**TEXT values MUST be handle-based** to avoid copying long prompts:

```json
{
  "kind": "TEXT",
  "v": {
    "id": "t:12345",
    "bytes": 81234,
    "preview": "first <=N bytes…",
    "preview_bytes": 256
  }
}
```

Rules:

* `preview` is OPTIONAL; if present, it MUST be truncated to `preview_bytes` and set `truncated.previews=true` if shortened.
* Never inline full prompt content in observations.

#### 1.1.3 Budget reporting (mandatory)

`budgets` MUST include at minimum:

```json
{
  "wall_time_ms": { "used": 12, "limit": 2000 },
  "stmts": { "used": 3, "limit": 50 },
  "total_bytes": { "used": 1048576, "limit": 8388608 },
  "subcalls": { "used": 1, "limit": 8 },
  "recursion_depth": { "used": 1, "limit": 4 }
}
```

* All counters are integers.
* Always report `used` and `limit` for each budget dimension.

#### 1.1.4 Events (structured, no prose)

`events[]` is optional but recommended for debugging and training:

```json
{ "t": "op", "op": "FIND_TEXT", "into": "pos", "ms": 1 }
{ "t": "subcall", "ms": 140, "bytes_in": 2048, "depth_cost": 1 }
```

#### 1.1.5 Error objects

On failure, `status!="ok"` and `errors[]` MUST contain at least one entry:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Keyword order mismatch for FIND_TEXT",
  "loc": { "file": "prog.rlm", "line": 12, "col": 3 },
  "hint": "Expected: SOURCE NEEDLE MODE IGNORE_CASE"
}
```

No stack traces in observations; stack traces go to host logs only.

---

### 1.2 DSL Emission (LLM → Host)

* The LLM MUST emit valid **RLM-DSL 0.1** only (no markdown fences unless explicitly required by the host protocol).
* Style MUST follow Anka-like constraints: **explicit intermediates**, **verbose keywords**, **single canonical form per op** ([arXiv][2])
* Reassignment is forbidden unless `Policy.AllowReassign=true` (default false).

---

## 2) RLM Paradigm Adherence

* **Prompt Isolation:** long prompt content lives in host storage; the LLM receives only compact metadata and small previews ([arXiv][1])
* **Symbolic Interaction:** access prompt only via ops (search/slice/window/regex/parse).
* **Managed Recursion:** use `SUBCALL` for decomposition. All recursion MUST be depth-limited and budgeted ([arXiv][1])

---

## 3) Reliability and Safety Standards

* **Canonical Forms:** every op has exactly one signature and keyword order (defined in `ops.json`).
* **Schema Enforcement:** both observations and (optionally) DSL AST dumps MUST validate against JSON Schemas (draft 2020-12) ([json-schema.org][3])
* **Capability Gating:** any non-pure op (LLM/FS/NET) is disabled by default and requires explicit policy enablement.
* **Budgets Everywhere:** time/steps/bytes/recursion/subcalls/parallelism.
* **No raw prompt leakage:** observations must not echo large prompt chunks; only previews/handles.

---

## 4) Engineering Standards (Go)

### 4.1 Separation of concerns

* `lex/parse/ast`: syntax only
* `ops/validate`: signature + type checking only
* `runtime`: execution + budgets + stores
* `host`: capability boundary (Subcall/FS/NET)
* `cmd`: CLI glue only

### 4.2 Determinism

* All **pure ops** MUST be deterministic and side-effect free.
* Non-determinism is limited to explicit capability ops (`SUBCALL`, optional IO).

### 4.3 Testing (TDD)

Must-have test categories:

* Lexer/parser golden tests (valid + invalid programs)
* Validator tests: keyword order, unknown ops, type mismatch, reassignment
* Runtime tests: bounds, truncation flags, budget exhaustion behavior
* Fake Host tests: recursion depth/count/parallelism enforcement
* Fuzz tests: parse/validate no panic

---

## 5) Documentation & Versioning

* `RLMDSL <version>` and `schema_version` MUST be bumped on breaking changes.
* Backwards compatibility policy must be documented (at minimum: “N-1 supported” or “best-effort parse with strict mode flag”).
* Ship:

  * `assets/ops.json`
  * `assets/dialect_card.md`
  * `schemas/obs.schema.json`
  * `schemas/ast.schema.json`
  * `schemas/exec_result.schema.json`

---

[1]: https://arxiv.org/html/2512.24601v2?utm_source=chatgpt.com "Recursive Language Models"
[2]: https://arxiv.org/html/2512.23214v1?utm_source=chatgpt.com "A Domain-Specific Language for Reliable LLM Code Generation"
[3]: https://json-schema.org/draft/2020-12?utm_source=chatgpt.com "Draft 2020-12"
