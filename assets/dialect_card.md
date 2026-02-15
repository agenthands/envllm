# EnvLLM-DSL Dialect Card (prepend to the LLM)
Output ONLY valid EnvLLM-DSL 0.1 code.

Rules:
- Use CELL blocks.
- Every OP line ends with INTO <var>.
- Keyword order must match ops.json.
- Never reuse a variable name.
- Literals: "string", 123, true, false, null.
- Indentation: exactly 2 spaces per statement.

Ops:
- STATS SOURCE <TEXT> INTO <JSON>
- FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE FIRST|LAST IGNORE_CASE true|false INTO <INT>
- WINDOW_TEXT SOURCE <TEXT> CENTER <INT> RADIUS <INT> INTO <TEXT>
- JSON_PARSE SOURCE <TEXT> INTO <JSON>
- JSON_GET SOURCE <JSON> PATH <TEXT> INTO <JSON>
- SUBCALL SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT> INTO <JSON>
- FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE FIRST|LAST INTO <SPAN>
- READ_FILE PATH <TEXT> INTO <TEXT>
- WRITE_FILE PATH <TEXT> SOURCE <TEXT> INTO <BOOL>
- LIST_DIR PATH <TEXT> INTO <JSON>
- CONCAT A <TEXT> B <TEXT> INTO <TEXT>
- GET_SPAN_START SOURCE <SPAN> INTO <INT>
- GET_SPAN_END SOURCE <SPAN> INTO <INT>
- SLICE_TEXT SOURCE <TEXT> START <INT> END <INT> INTO <TEXT>

Notes:
- NO string concatenation with '+'. Use CONCAT.
- NO property access with '.'. Use JSON_GET or GET_SPAN_START.
- NO templates with '{{}}' or '${}'.
- If you need to transform text to JSON, use SUBCALL or JSON_PARSE.
