# EnvLLM-DSL Dialect Card (prepend to the LLM)
Output ONLY valid EnvLLM-DSL 0.2 code.

Rules:
- Use CELL blocks.
- Every OP line ends with INTO <var>: <Type>.
- Keyword order must match ops.json.
- Never reuse a variable name.
- Literals: "string", 123, true, false, null.
- Indentation: exactly 2 spaces per statement.
- Types: TEXT, INT, OFFSET, SPAN, BOOL, JSON.

Ops:
- STATS SOURCE <TEXT> INTO <JSON>
  - Returns: {"bytes": INT, "lines": INT}
- FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE FIRST|LAST IGNORE_CASE true|false INTO <OFFSET>
- WINDOW_TEXT SOURCE <TEXT> CENTER <OFFSET> RADIUS <INT> INTO <TEXT>
- SLICE_TEXT SOURCE <TEXT> START <OFFSET> END <OFFSET> INTO <TEXT>
- JSON_PARSE SOURCE <TEXT> INTO <JSON>
- JSON_GET SOURCE <JSON> PATH <TEXT> INTO <JSON>
- SUBCALL SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT> INTO <JSON>
- FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE FIRST|LAST INTO <SPAN>
- READ_FILE PATH <TEXT> INTO <TEXT>
- WRITE_FILE PATH <TEXT> SOURCE <TEXT> INTO <BOOL>
- LIST_DIR PATH <TEXT> INTO <JSON>
- CONCAT_TEXT A <TEXT> B <TEXT> INTO <TEXT>
- TO_TEXT VALUE <any> INTO <TEXT>
- OFFSET VALUE <INT> INTO <OFFSET>
- OFFSET_ADD OFFSET <OFFSET> AMOUNT <INT> INTO <OFFSET>
- SPAN START <OFFSET> END <OFFSET> INTO <SPAN>
- GET_SPAN_START SOURCE <SPAN> INTO <OFFSET>
- GET_SPAN_END SOURCE <SPAN> INTO <OFFSET>

Notes:
- NO string concatenation with '+'. Use CONCAT_TEXT.
- NO property access with '.'. Use JSON_GET or GET_SPAN_START.
- NO implicit conversion. Use TO_TEXT to convert OFFSET/INT to TEXT before CONCAT.
- OFFSET is an opaque position handle. Do not use as a number.
- If you need to transform text to JSON, use SUBCALL or JSON_PARSE.

Examples:

1. Declaring Capabilities:
REQUIRES capability="fs_read"
CELL read:
  READ_FILE PATH "log.txt" INTO content: TEXT
  SET_FINAL SOURCE content

2. Text Slicing with Math:
CELL slice:
  FIND_TEXT SOURCE PROMPT NEEDLE "start" MODE FIRST IGNORE_CASE false INTO pos: OFFSET
  OFFSET_ADD OFFSET pos AMOUNT 5 INTO start: OFFSET
  SLICE_TEXT SOURCE PROMPT START pos END start INTO snippet: TEXT
  SET_FINAL SOURCE snippet
