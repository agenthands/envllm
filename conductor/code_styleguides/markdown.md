# Markdown Style Guide: EnvLLM

These rules ensure that project documentation remains consistent and actionable for both human developers and machine interfaces.

## Required Project Markdown Rules

### 1. Mandatory Header Fields
Every specification document MUST contain:
```markdown
Version:
Status:
Breaking Changes:
Schema Version:
```

### 2. DSL Example Tagging
All RLM-DSL code examples must be fenced with the `text` language tag to distinguish them from other code blocks:
````text
```text
CELL plan:
  STATS SOURCE PROMPT INTO stats
```
````

### 3. Concrete Specifications
All behavior descriptions must be grounded in concrete examples. A specification is incomplete unless it includes:
- Example Input
- Example Output JSON
- Failure/Error Example
Avoid narrative-only descriptions for machine-interface behaviors.
