# EnvLLM-DSL Specification (v0.2)

## 1. Syntax (EBNF)

The grammar is strictly defined. Whitespace is significant (indentation).

```ebnf
program         = { cell } ;
cell            = "CELL", ws, ident, ":", newline, { stmt_line } ;
stmt_line       = indent, stmt, newline ;
indent          = "  " ; (* Exactly 2 spaces *)

stmt            = op_stmt | set_final | assert_stmt | print_stmt ;

op_stmt         = op_name, { ws, kw_arg }, ws, "INTO", ws, ident, [ ":", ws, type_name ] ;
kw_arg          = kw_name, ws, expr ;

set_final       = "SET_FINAL", ws, "SOURCE", ws, expr ;
assert_stmt     = "ASSERT", ws, "COND", ws, expr, ws, "MESSAGE", ws, string ;
print_stmt      = "PRINT", ws, "SOURCE", ws, expr ;

expr            = literal | ident ;
literal         = string | int | bool | "null" ;
```

## 2. Type System

Variables are strictly typed. The runtime supports:

- **TEXT**: A string or a handle to a large text block (zero-copy).
- **INT**: 64-bit signed integer.
- **BOOL**: Boolean (`true`, `false`).
- **JSON**: A generic JSON object or array.
- **SPAN**: A text range `{start: INT, end: INT}`.
- **NULL**: Represents absence of value.

## 3. Standard Operations (Core Module)

### Text Processing
- `STATS SOURCE <TEXT> INTO <JSON>`: Get metadata (bytes, lines).
- `FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE <FIRST|LAST> IGNORE_CASE <BOOL> INTO <INT>`: Search for exact string.
- `FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE <FIRST|LAST> INTO <SPAN>`: Search using Regex.
- `WINDOW_TEXT SOURCE <TEXT> CENTER <INT> RADIUS <INT> INTO <TEXT>`: Extract context around an offset.
- `SLICE_TEXT SOURCE <TEXT> START <INT> END <INT> INTO <TEXT>`: Extract precise substring.
- `CONCAT A <TEXT> B <TEXT> INTO <TEXT>`: Join two text values.

### Data Extraction
- `JSON_PARSE SOURCE <TEXT> INTO <JSON>`: Parse text into structured data.
- `JSON_GET SOURCE <JSON> PATH <TEXT> INTO <JSON>`: Extract value via dot-path (e.g., "user.id").
- `GET_SPAN_START SOURCE <SPAN> INTO <INT>`: Get start offset.
- `GET_SPAN_END SOURCE <SPAN> INTO <INT>`: Get end offset.

### Control & Recursion
- `SUBCALL SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT> INTO <JSON>`: Delegate a sub-task to the agent.

## 4. Modes

- **STRICT Mode**: Enforces all grammar rules, including indentation and type annotations. Used for generation and CI.
- **COMPAT Mode**: Relaxed parsing for older scripts. Allows loose indentation and inferred types.
