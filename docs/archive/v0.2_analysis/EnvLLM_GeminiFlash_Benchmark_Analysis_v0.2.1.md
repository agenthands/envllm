# EnvLLM Benchmark (Gemini 1.5 Flash): Results → Fix Plan (v0.2.1)

**Audience:** engineering team (compiler/runtime + extensions)
**Date:** 2026-02-16 (Europe/Kyiv)

---

## 1) What the benchmark proved (good news)

Your stabilization work is working as intended: **every LLM output was either executed safely or rejected with a structured error**. The runtime is now a “safety membrane” that blocks:

* **Malformed syntax** (STRICT indentation / missing type annotations)
* **Unauthorized operations** via **capability gating** (e.g., `READ_FILE`, `SUBCALL`)
* **Unsafe/ill-typed logic** via parse/lint/typecheck + runtime checks
* Lexer robustness improved (escaped quote termination bug fixed and committed)

This is the right foundation for scaling features through the **Extension Framework**.

---

## 2) Benchmark summary (Gemini 1.5 Flash, zero-shot)

| Suite           |  Total | Passed |  Success | Primary failure reason                  |
| --------------- | -----: | -----: | -------: | --------------------------------------- |
| A: Reliability  |      4 |      1 |      25% | Capability denials / JSON hallucination |
| B: VM & Budgets |      3 |      0 |       0% | Property-access hallucination (`.cost`) |
| C: RLM Tasks    |      4 |      1 |      25% | Type mismatch (INT vs SPAN)             |
| D: Recursion    |      1 |      0 |       0% | Missing JSON keys                       |
| **TOTAL**       | **12** |  **2** | **~17%** |                                         |

**Key observation:** Failures are not “hard reasoning failures”; they’re mostly **surface-form mismatches** between what the model expects (dot access, struct fields, arbitrary JSON keys) and what the DSL permits (getter ops + explicit schemas).

---

## 3) Failure taxonomy (what to fix first)

### F1 — Property access hallucination (highest impact)

The model repeatedly emits:

* `result.cost`
* `span.start`
* `obj.steps` (or similar)

But EnvLLM-DSL requires:

* `JSON_GET ...` (for JSON) or
* typed getter ops (e.g., `GET_SPAN_START span=...`)

**Impact:** Wipes out *Suite B entirely* and contributes to A/D failures.

---

### F2 — Type confusion (INT vs SPAN)

The model often passes:

* result of `FIND_TEXT` (**INT**) into `GET_SPAN_START` (**requires SPAN**)

Even though the correct workflow is:

* `off = FIND_TEXT(...)`
* `span = WINDOW_TEXT(offset=off, ...)` or `FIND_REGEX(...)` to obtain a SPAN
* then `GET_SPAN_START(span=span)`

**Impact:** Core contributor to Suite C failures.

---

### F3 — Capability probing (“explore the environment”)

Even when restricted, the model tries:

* `READ_FILE`
* `SUBCALL`
* other operations not in allowed capability set

**Impact:** Drives Suite A `capability_denied`.

---

### F4 — Over-engineering / meta-programming in DSL

Model tries to write test-like logic inside DSL:

* searching for a `steps` key in a non-existent JSON
* assuming ubiquitous objects exist (`ctx`, `response`, `result.schema`, etc.)

**Impact:** Drives D “missing JSON keys” and some A/B failures.

---

## 4) Engineering goal for v0.2.1

Increase zero-shot pass rate by eliminating avoidable surface mismatches **without weakening safety**.

**Principle:**
Keep STRICT traps for unsafe behavior, but add *mechanical, safe* affordances so the model’s default habits map to valid code.

---

## 5) Concrete fixes (v0.2.1) — minimal, high leverage

### 5.1 Introduce “typed field access” as FIRST-CLASS ops (not JSON by default)

**Problem:** the model thinks outputs are structs with fields.

**Fix:** make that true in a controlled way:

* define closed typed getters for the common cases
* forbid arbitrary key access unless the value is explicitly `Json`

#### Required getters

* `GET_COST result=<VmResult> INTO cost: Cost`
* `GET_SPAN_START span=<Span> INTO start: Offset`
* `GET_SPAN_END span=<Span> INTO end: Offset`

**Acceptance tests:** Suite B should stop failing on `.cost` once the DSL card clearly advertises `GET_COST`.

---

### 5.2 Split ambiguous types: `Int` ≠ `Offset` ≠ `Cost`

Right now “INT” is overloaded, causing confusion.

**Fix:** Add distinct scalar types with no implicit coercion:

* `Int` — generic integer
* `Offset` — an index into a corpus/text buffer
* `Cost` — budget units / cost units

Then update op signatures so the model can’t pass the wrong thing:

* `FIND_TEXT … INTO off: Offset` (NOT Int)
* `FIND_REGEX … INTO sp: Span`
* `WINDOW_TEXT offset=<Offset> … INTO sp: Span` *(or returns {Span, Text} explicitly)*

**Typecheck rule:** No implicit `Offset <-> Int` coercion in STRICT.

---

### 5.3 Add an explicit bridge op: `AS_SPAN` (Offset → Span)

This is the canonical way to recover when the model only has an offset.

```text
AS_SPAN offset=<Offset> len=<Int>
INTO sp: Span
```

Recommended canonical defaults:

* `len=0` means “point span” where `start=end=offset`

This makes “offset used where span expected” repairable in **one step**.

---

### 5.4 Add a STRICT lint rule + repair hint for dot access

**Do not accept dot access as valid syntax** in STRICT (keep trap), but make rejection highly repairable.

Lint code:

* `LINT_DOT_ACCESS_FORBIDDEN`

Structured error should include an **auto-fix template**:

```json
{
  "code": "LINT_DOT_ACCESS_FORBIDDEN",
  "step": "budget_check",
  "message": "Dot access (result.cost) is not allowed in STRICT mode.",
  "hint_template": "GET_COST result=result INTO cost: Cost"
}
```

Same for `span.start`:

```json
{
  "hint_template": "GET_SPAN_START span=sp INTO start: Offset"
}
```

---

### 5.5 Add a STRICT lint rule: NO_META (ban “program-inspecting” patterns)

Reject references to undeclared identifiers / invented universal objects:

* `steps`, `ctx`, `response`, `schema`, etc. unless produced by a prior op.

Lint code:

* `LINT_UNKNOWN_IDENTIFIER`
* `LINT_META_PROGRAMMING_FORBIDDEN` (optional umbrella)

Add repair guidance:

* “Use ASSERT / typed getters / documented outputs only.”

---

### 5.6 Capability probing: improve the error payload and task contract

Capability gating is correct; the missing piece is “how to recover”.

When capability is denied:

* include the attempted op
* include allowed capability list
* include 1–3 suggested alternatives (module ops) if applicable

```json
{
  "code": "ERR_CAPABILITY_DENIED",
  "op": "READ_FILE",
  "allowed": ["env.query", "env.list", "web.navigate", "web.dom.query"],
  "hint": "Use ENV_QUERY corpus=<...> query=<...> top_k=<...> INTO snips: Snippets"
}
```

**Also:** in STRICT mode, require tasks to declare capabilities up front:

```text
REQUIRES capability="env.query"
REQUIRES capability="web.dom.query"
```

This reduces “probing” because the model sees what’s allowed.

---

## 6) Canonical templates to add to the dialect card (v0.2.1)

> These are the minimum additions to address the benchmark failures.

### Scalars

```text
Offset
Cost
Span = {start: Offset, end: Offset}
```

### Text search

```text
FIND_TEXT corpus=<CorpusRef> query=<Text> INTO off: Offset
FIND_REGEX corpus=<CorpusRef> pattern=<Regex> INTO sp: Span
WINDOW_TEXT corpus=<CorpusRef> offset=<Offset> before=<Int> after=<Int> INTO sp: Span
AS_SPAN offset=<Offset> len=<Int> INTO sp: Span
```

### Span getters

```text
GET_SPAN_START span=<Span> INTO start: Offset
GET_SPAN_END span=<Span> INTO end: Offset
```

### VM/Budget getters

```text
GET_COST result=<VmResult> INTO cost: Cost
```

### JSON (only when explicitly typed as Json)

```text
JSON_GET obj=<Json> path=<JsonPath> INTO out: Json
```

---

## 7) Type-directed repair rules (compiler/runtime feature)

Add a dedicated phase: `ExplainTypeError(err) -> RepairHint`.

### Rule R1 — Offset passed where Span expected

If a step calls `GET_SPAN_START(span=<Offset>)`:

* Suggest inserting `AS_SPAN offset=<Offset> len=0 INTO tmp: Span`
* Replace call with `GET_SPAN_START span=tmp …`

### Rule R2 — CONCAT over non-Text

If `CONCAT` is used with numeric types:

* Suggest `TO_TEXT` conversions (if you support them), or forbid `CONCAT` entirely and provide `CONCAT_TEXT` only.

### Rule R3 — Dot access

If parse/lint sees `a.b`:

* if `a` is `Span` and `b in {start,end}` → `GET_SPAN_*`
* if `a` is `VmResult` and `b == cost` → `GET_COST`
* else → `JSON_GET` (only if `a` is Json), otherwise reject with “no such field” + docs pointer.

---

## 8) Benchmark suite updates (so progress is measurable)

### 8.1 Add variants that measure “repairability”

For each failing class, add two cases:

* raw (should fail with structured error)
* repair loop enabled (should converge)

Examples:

* Dot access:

  * raw uses `result.cost` → should fail with `LINT_DOT_ACCESS_FORBIDDEN`
  * repaired uses `GET_COST` → should pass
* Type mismatch:

  * raw passes Offset to Span → fail with `TYPE_MISMATCH_FIELD`
  * repair inserts `AS_SPAN` → pass

### 8.2 Track new metrics

Per suite:

* `reject_parse`, `reject_lint`, `reject_type`, `capability_denied`
* `avg_repairs_per_case`
* `p95_repairs_per_case`

---

## 9) Implementation checklist (v0.2.1)

### Compiler / DSL

* [ ] Add types: `Offset`, `Cost`, `Span`
* [ ] Update op signatures: `FIND_TEXT -> Offset` (not Int)
* [ ] Add ops: `AS_SPAN`, `GET_SPAN_START`, `GET_SPAN_END`, `GET_COST`
* [ ] Add lint: `LINT_DOT_ACCESS_FORBIDDEN`
* [ ] Add lint: `LINT_UNKNOWN_IDENTIFIER` (NO_META)
* [ ] Add type-directed repair hint generator

### Runtime / Policy

* [ ] Improve `capability_denied` payload with “allowed + suggested alternatives”
* [ ] Require `REQUIRES` in STRICT mode (capability contract)

### Bench harness

* [ ] Add repairability variants + new metrics

---

## 10) Expected effect (why this should move the needle)

* **Suite B** stops failing on `.cost` because there is a single canonical `GET_COST`.
* **Suite C** stops failing on INT vs SPAN because:

  * `FIND_TEXT` returns `Offset` (not generic Int)
  * `AS_SPAN` provides an explicit bridge
  * repair hints fix mismatches deterministically
* **Suites A/D** improve because NO_META blocks invented objects, errors become repairable, and capability errors become guided instead of dead-ends.

---

## 11) References (design justification)

### Anka (canonical DSL for LLM reliability)

* Paper: [https://arxiv.org/abs/2512.23214](https://arxiv.org/abs/2512.23214)
* Syntax guide (“one way to do things”): [https://raw.githubusercontent.com/BleBlo/Anka/main/docs/syntax-guide.md](https://raw.githubusercontent.com/BleBlo/Anka/main/docs/syntax-guide.md)

### RLM (prompt as environment + query/subcalls)

* Paper: [https://arxiv.org/abs/2512.24601](https://arxiv.org/abs/2512.24601)
* HTML: [https://arxiv.org/html/2512.24601v1](https://arxiv.org/html/2512.24601v1)

---

## Appendix A — Example: fixing the “INT vs SPAN” failure

**LLM bad output (common):**

```text
STEP s1:
  FIND_TEXT corpus=logs query="ERROR"
  INTO off: Offset

STEP s2:
  GET_SPAN_START span=off
  INTO start: Offset
```

**Type-directed repair (canonical):**

```text
STEP s1:
  FIND_TEXT corpus=logs query="ERROR"
  INTO off: Offset

STEP s1b:
  AS_SPAN offset=off len=0
  INTO sp: Span

STEP s2:
  GET_SPAN_START span=sp
  INTO start: Offset
```

---

## Appendix B — Example: fixing `.cost`

**LLM bad output:**

```text
STEP budget:
  vm = RUN_VM ...
  cost = vm.cost
```

**STRICT canonical:**

```text
STEP budget:
  RUN_VM ...
  INTO vm: VmResult

STEP budget2:
  GET_COST result=vm
  INTO cost: Cost
```

---
