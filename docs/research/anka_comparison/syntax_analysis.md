# Anka Syntax Guide (Fetched for Analysis)

This file contains the content fetched from https://github.com/BleBlo/Anka/docs/syntax-guide.md.

## 1. Design Principles (Anka)
- **One Canonical Form:** One way to do things.
- **Named Intermediates:** `INTO` clause mandatory.
- **Explicit Step Structure:** `STEP` blocks.
- **Verbose Keywords:** English over symbols.

## 2. Key Differences in Prompting (Inferred)
Anka's syntax guide serves as its primary prompt material. It explicitly lists every operation with concrete examples.

### Example Anka Prompt Structure:
1.  **Language Definition**: "You are generating code in Anka..."
2.  **Schema**: "Input schema: TABLE[...]"
3.  **Task**: "Filter users where age > 30..."
4.  **Syntax Guide**: (Injected content from syntax-guide.md)

### Edge Case Handling:
- **Missing Keys:** Anka uses schema-aware `FILTER` conditions (e.g. `WHERE age > 30`). EnvLLM uses `JSON_GET` which is looser.
- **Type Coercion:** Anka has typed columns (INT, DECIMAL). EnvLLM has stricter `OFFSET` vs `INT` separation but relies on `TO_TEXT` for explicit conversion.
- **Control Flow:** Anka has `FOR_EACH` for iteration. EnvLLM v0.2.2 adds this.

## 3. Implementation Gaps & Action Items
1.  **Prompt Density**: Anka's guide is dense with examples. Our `dialect_card.md` is minimal. We should expand it to be a full "Syntax Guide" like Anka's.
2.  **Schema Awareness**: Anka prompts include table schemas. We should encourage including `STRUCT` definitions in our system prompt.
3.  **Error Recovery**: Anka relies on "Constrained Decoding" (grammar enforcement during generation). We use a "Repair Loop" (post-generation fix). This is a fundamental architectural difference. Our approach is more flexible but requires better error messages (which we added).
