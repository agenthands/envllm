# Protocol Specification: RLM-Go

This document defines the lifecycle and contract between the RLM Host and the RLM Runtime.

## 1. Lifecycle
1. **Initialize:** Host provides long prompt and sets initial policy/budgets.
2. **Loop:**
    - Host provides Observation JSON to LLM.
    - LLM emits DSL Cell.
    - Runtime validates and executes Cell.
    - Host captures Results/Traces/Budgets.
3. **Termination:** Loop ends when `SET_FINAL` is called, an error occurs, or a budget is exhausted.

## 2. Observation Schema Rules
- **Constant-size Metadata:** Observations should not grow linearly with the prompt size. Use handles and previews.
- **Identity:** Every observation MUST include a `cell` object identifying the executed cell by `name` and `index`.
- **Delta-only:** `vars_delta` should only contain variables from the current execution turn.
- **Truncation Flags:** Always set boolean flags if any content (previews, prints) was truncated.

## 3. DSL Emission Contract
- The LLM must output RLM-DSL 0.1 exclusively.
- No conversational preambles or postambles.
- Strict adherence to the `ops.json` signatures.

## 4. Resource Accounting
- **Recursion Depth:** Tracked via `SUBCALL`. Each call increments the depth counter.
- **Statement Budget:** Every statement in a CELL consumes 1 unit.
- **Byte Budget:** Total memory used by all `vars` in the environment.

## 5. Truncation Logic
- **Previews:** Host must truncate `TEXT` previews to the `preview_bytes` limit defined in the policy.
- **Obs JSON:** If the total Observation JSON exceeds its limit, the host must drop older events or larger previews and set `truncated.obs=true`.
