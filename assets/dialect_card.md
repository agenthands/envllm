# EnvLLM-DSL Dialect Card (v0.2.3)
Output ONLY valid EnvLLM-DSL 0.2 code.

### **Mandatory Structure**
Every program MUST follow this block hierarchy:
1. `RLMDSL 0.2` (Header)
2. `TASK <name>:` (The container)
3. `  INPUT <name>: <Type>` (Optional inputs)
4. `  REQUIRES capability="<cap>"` (Optional security declarations)
5. `  CELL <name>:` (Execution blocks)
6. `    <OP> ... INTO <var>: <Type>` (Indented statements)
7. `  OUTPUT <var>` (Final return value)

### **Strict Rules**
- **Indentation**: Exactly 2 spaces for top-level (INPUT/CELL), exactly 4 spaces for statements inside a CELL.
- **Explicit Types**: Every `INTO <var>` must be followed by `: <Type>` (TEXT, INT, OFFSET, SPAN, BOOL, JSON, STRUCT).
- **No Variable Reuse**: Every `INTO` must use a unique variable name.
- **NO HARDCODED OFFSETS**: Never use `OFFSET VALUE 123`. Use `FIND_TEXT` or `FIND_REGEX`.
- **Keyword order**: Must match the operation signature exactly.

### **Common Operations**
- `STATS SOURCE <TEXT> INTO <var>: STRUCT`
- `GET_FIELD SOURCE <STRUCT> FIELD <TEXT> INTO <var>: JSON`
- `FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE FIRST|LAST IGNORE_CASE true|false INTO <var>: OFFSET`
- `WINDOW_TEXT SOURCE <TEXT> CENTER <OFFSET> RADIUS <INT> INTO <var>: TEXT`
- `SLICE_TEXT SOURCE <TEXT> START <OFFSET> END <OFFSET> INTO <var>: TEXT`
- `JSON_PARSE SOURCE <TEXT> INTO <var>: JSON`
- `JSON_GET SOURCE <JSON> PATH <TEXT> INTO <var>: JSON`
- `SUBCALL SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT> INTO <var>: JSON`
- `FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE FIRST|LAST INTO <var>: SPAN`
- `GET_SPAN_START SOURCE <SPAN> INTO <var>: OFFSET`
- `GET_SPAN_END SOURCE <SPAN> INTO <var>: OFFSET`
- `TO_TEXT VALUE <any> INTO <var>: TEXT`
- `OFFSET_ADD OFFSET <OFFSET> AMOUNT <INT> INTO <var>: OFFSET`

### **Examples**

1. **Extraction Task:**
RLMDSL 0.2
TASK get_price:
  INPUT PROMPT: TEXT
  CELL search:
    FIND_TEXT SOURCE PROMPT NEEDLE "$" MODE FIRST IGNORE_CASE false INTO p_pos: OFFSET
    FIND_REGEX SOURCE PROMPT PATTERN "[0-9.]+" MODE FIRST INTO p_span: SPAN
  CELL extraction:
    GET_SPAN_START SOURCE p_span INTO p_start: OFFSET
    GET_SPAN_END SOURCE p_span INTO p_end: OFFSET
    SLICE_TEXT SOURCE PROMPT START p_start END p_end INTO price: TEXT
  OUTPUT price

3. **Substring Extraction (Precision):**
RLMDSL 0.2
TASK extract_value:
  INPUT PROMPT: TEXT
  CELL search:
    FIND_TEXT SOURCE PROMPT NEEDLE "value is: " MODE FIRST IGNORE_CASE false INTO label_pos: OFFSET
    -- Offset by length of "value is: " (10 chars)
    OFFSET_ADD OFFSET label_pos AMOUNT 10 INTO start_pos: OFFSET
    FIND_TEXT SOURCE PROMPT NEEDLE " " MODE FIRST IGNORE_CASE false INTO space_pos: OFFSET
  CELL slice:
    SLICE_TEXT SOURCE PROMPT START start_pos END space_pos INTO secret: TEXT
  OUTPUT secret
