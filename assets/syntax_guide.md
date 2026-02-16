# EnvLLM-DSL Syntax Guide (v0.2.2)

EnvLLM-DSL is a strict, typed language for controlling an AI environment. This guide serves as the primary reference for the model.

## 1. Core Principles
*   **Canonical Form**: There is only one way to perform an action. Do not invent aliases.
*   **Explicit State**: Every operation must produce a named output (`INTO <var>: <Type>`).
*   **Strict Typing**: Do not mix `OFFSET`, `INT`, and `TEXT`. Use conversion ops.
*   **No Shadowing**: Never reuse a variable name.

## 2. Structure & Layout
A program is a sequence of `CELL` blocks.
*   **Indentation**: EXACTLY 2 spaces for statements.
*   **Requirements**: Declare capabilities at the top.

```text
REQUIRES capability="fs_read"

CELL parse_log:
  READ_FILE PATH "/var/log/syslog" INTO log_content: TEXT
  STATS SOURCE log_content INTO log_stats: STRUCT
  GET_FIELD SOURCE log_stats FIELD "lines" INTO line_count: INT
  SET_FINAL SOURCE line_count
```

## 3. Type System & Literals
*   **TEXT**: `"hello\nworld"` (Strings must use double quotes)
*   **INT**: `123`, `-42` (Whole numbers)
*   **OFFSET**: Opaque text position. Created by `FIND_TEXT` or `OFFSET` literal.
*   **SPAN**: Range `{start: OFFSET, end: OFFSET}`.
*   **BOOL**: `true`, `false`
*   **NULL**: `null`
*   **STRUCT**: Typed record (e.g. `VmStats`). Access via `GET_FIELD`.
*   **ROWS**: List of Structs. Iterate via `FOR_EACH`.

## 4. Operations Reference (Dense Examples)

### Text Processing
*   **Search**: `FIND_TEXT SOURCE <TEXT> NEEDLE "error" MODE FIRST IGNORE_CASE true INTO pos: OFFSET`
*   **Regex**: `FIND_REGEX SOURCE <TEXT> PATTERN "\d+" MODE FIRST INTO span: SPAN`
*   **Extract**: `SLICE_TEXT SOURCE <TEXT> START pos END next_pos INTO chunk: TEXT`
*   **Context**: `WINDOW_TEXT SOURCE <TEXT> CENTER pos RADIUS 100 INTO context: TEXT`

**Edge Case: Offsets are not Integers**
*   *Wrong:* `CONCAT_TEXT A pos B 1` (Type Mismatch)
*   *Right:* `OFFSET_ADD OFFSET pos AMOUNT 1 INTO next_pos: OFFSET`

### Data & Math
*   **Math**: `OFFSET_ADD` is the only math op. For general math, use `SUBCALL`.
*   **JSON**: `JSON_PARSE SOURCE text INTO json: JSON`
*   **Access**: `JSON_GET SOURCE json PATH "key" INTO val: JSON`
*   **Structs**: `GET_FIELD SOURCE stats FIELD "cost" INTO cost: COST`

### Control Flow (Iteration)
Use `FOR_EACH` to process lists of data (ROWS).

```text
FOR_EACH row IN rows LIMIT 10:
  GET_FIELD SOURCE row FIELD "id" INTO id: INT
  TO_TEXT VALUE id INTO id_text: TEXT
  PRINT SOURCE id_text
```

### Recursion (The "Loop")
Use `SUBCALL` to delegate complex reasoning or loops.

```text
SUBCALL SOURCE text TASK "Summarize this" DEPTH_COST 1 INTO summary: TEXT
```

## 5. Error Handling & Recovery
EnvLLM halts on runtime errors. To handle potential failures, use logical checks or `SUBCALL` isolation.

**Pattern: Try/Recovery (via Subcall)**
If a task might fail, isolate it:
```text
CELL safe_attempt:
  SUBCALL SOURCE data TASK "Risky parse operation" DEPTH_COST 1 INTO result: JSON
  ASSERT COND true MESSAGE "Check if result is valid"
  SET_FINAL SOURCE result
```

## 6. Common Pitfalls (Avoid These)
1.  **Dot Access**: `result.cost` is FORBIDDEN. Use `GET_FIELD`.
2.  **String Concatenation**: `text + " end"` is FORBIDDEN. Use `CONCAT_TEXT`.
3.  **Variable Reuse**: `INTO out` twice is FORBIDDEN. Use `out_1`, `out_2`.
4.  **Implicit Conversion**: `CONCAT_TEXT A "ID: " B 123` is FORBIDDEN. Use `TO_TEXT VALUE 123` first.
