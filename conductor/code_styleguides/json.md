# JSON Style & Protocol Guide: RLM-Go

For RLM-Go, JSON is a **protocol**, not just a configuration format. These rules are mandatory for all host-generated and runtime-consumed JSON.

## JSON Contract Rules

### 1. Mandatory Schema Validation
All runtime JSON MUST be validated against the official JSON Schemas (Draft 2020-12). This applies to:
- Observations
- Execution Results
- AST Dumps
- Host Input
No “best effort” parsing is allowed.

### 2. No Optional Structural Keys
Structural keys must always be present.
- **Allowed:** `{ "errors": [] }`
- **Forbidden:** `{}` (missing `errors` key)
- **Reason:** Missing keys cause unpredictable LLM parsing behavior and increase interpretation entropy.

### 3. Stable Ordering for Generated JSON
All host-generated JSON must be marshalled using:
- Deterministic struct encoding.
- Stable key ordering for maps (keys must be sorted alphabetically).
- **Reason:** Non-stable JSON causes diff noise and makes debugging RLM traces significantly harder.

### 4. Tagged Value Encoding
Never send raw values for dynamic data. Always use tagged encoding to ensure type safety for the LLM.
- **Forbidden:** `"pos": 450`
- **Required:** `"pos": {"kind":"INT","v":450}`
- **Reason:** Prevents LLM type inference errors and maintains a strict machine-to-machine contract.
