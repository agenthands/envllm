# EnvLLM-DSL Syntax Guide (v0.2.3)

EnvLLM-DSL is a strict, typed language for controlling an AI environment. This guide serves as the primary reference for the model.

## 1. Core Principles
*   **Canonical Form**: There is only one way to perform an action. Do not invent aliases.
*   **Explicit State**: Every operation must produce a named output (`INTO <var>: <Type>`).
*   **Strict Typing**: Do not mix `OFFSET`, `INT`, and `TEXT`. Use conversion ops.
*   **No Shadowing**: Never reuse a variable name.

## 2. Structure & Layout
A program is a `TASK` containing `INPUT` declarations, a body of `CELL` or `IF` blocks, and an `OUTPUT`.

```text
TASK process_log:
  INPUT logs: TEXT
  
  REQUIRES capability="fs_read"

  IF true:
    CELL read:
      READ_FILE PATH "/var/log/syslog" INTO log_content: TEXT
      SET_FINAL SOURCE log_content
  END

  OUTPUT log_content
```

## 3. Control Flow

### IF/ELSE Branching
Use `IF` to branch logic based on boolean conditions.
```text
IF is_valid:
  CELL ok:
    PRINT SOURCE "valid"
ELSE:
  CELL fail:
    PRINT SOURCE "invalid"
END
```

### FOR_EACH Iteration
Use `FOR_EACH` to process lists of data (ROWS).
```text
FOR_EACH row IN rows LIMIT 10:
  CELL process:
    GET_FIELD SOURCE row FIELD "id" INTO id: INT
    PRINT SOURCE id
```

## 4. Type System & Literals
*   **TEXT**: `"hello\nworld"`
*   **INT**: `123`, `-42`
*   **OFFSET**: Opaque position.
*   **SPAN**: Range `{start: OFFSET, end: OFFSET}`.
*   **BOOL**: `true`, `false`
*   **STRUCT**: Typed record. Access via `GET_FIELD`.
*   **ROWS**: List of Structs.

## 5. Operations Reference

### Text Processing
*   `FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE <FIRST|LAST> IGNORE_CASE <BOOL> INTO <OFFSET>`
*   `FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE <FIRST|LAST> INTO <SPAN>`
*   `WINDOW_TEXT SOURCE <TEXT> CENTER <OFFSET> RADIUS <INT> INTO <TEXT>`
*   `SLICE_TEXT SOURCE <TEXT> START <OFFSET> END <OFFSET> INTO <TEXT>`

### Data & Math
*   `OFFSET_ADD OFFSET <OFFSET> AMOUNT <INT> INTO <OFFSET>`
*   `GET_FIELD SOURCE <STRUCT> FIELD <String> INTO <Any>`
*   `JSON_PARSE SOURCE <TEXT> INTO <JSON>`
*   `SUBCALL SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT> INTO <JSON>`

## 6. Common Pitfalls
1.  **Dot Access**: `result.cost` is FORBIDDEN. Use `GET_FIELD`.
2.  **String Concatenation**: `text + " end"` is FORBIDDEN. Use `CONCAT_TEXT`.
3.  **Variable Reuse**: `INTO out` twice is FORBIDDEN.
