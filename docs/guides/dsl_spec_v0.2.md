# EnvLLM-DSL Specification (v0.2)

## 1. Syntax (EBNF)

The grammar is strictly defined. Whitespace is significant (indentation).

```ebnf
(* =========================
   EnvLLM-DSL 0.2 — Full EBNF
   ========================= *)

program         = ws, [header], ws, { cell, ws }, EOF ;

header          = "RLMDSL", req_ws, version, line_end ;
version         = "0.1" | "0.2" ;

cell            = "CELL", req_ws, ident, ":", line_end,
                  { stmt_line } ;

stmt_line       = indent, stmt, line_end ;

stmt            = op_stmt
                | set_final
                | assert_stmt
                | print_stmt
                | empty_stmt ;

empty_stmt      = (* empty line is allowed inside CELL *) ;

(* ---------- Statements ---------- *)

op_stmt         = op_name, { req_ws, kw_arg }, req_ws, "INTO", req_ws, ident, [ ":", req_ws, type_name ] ;

kw_arg          = kw_name, req_ws, expr ;

set_final       = "SET_FINAL", req_ws, "SOURCE", req_ws, expr ;

assert_stmt     = "ASSERT", req_ws, "COND", req_ws, expr,
                  req_ws, "MESSAGE", req_ws, string ;

print_stmt      = "PRINT", req_ws, "SOURCE", req_ws, expr ;

(* ---------- Expressions ---------- *)

expr            = literal
                | ident ;

literal         = string
                | int
                | bool
                | "null" ;

(* ---------- Lexical elements ---------- *)

op_name         = UIDENT ;
kw_name         = UIDENT ;

ident           = IDENT ;
type_name       = UIDENT ;

UIDENT          = UPPER, { UPPER | DIGIT | "_" } ;
IDENT           = ( LETTER | "_" ), { LETTER | DIGIT | "_" } ;

(* ---------- Literals ---------- *)

int             = [ "-" ], ( "0" | ( NONZERO, { DIGIT } ) ) ;

bool            = "true" | "false" ;

string          = DQUOTE, { str_char }, DQUOTE ;

str_char        = unescaped_char | escape_seq ;

unescaped_char  = ? any Unicode scalar value except DQUOTE and backslash and line breaks ? ;

escape_seq      = "\\\\"
                | "\\\""
                | "\\n"
                | "\\r"
                | "\\t"
                | "\\u", hex, hex, hex, hex ;

hex             = DIGIT | "a" | "b" | "c" | "d" | "e" | "f"
                        | "A" | "B" | "C" | "D" | "E" | "F" ;

(* ---------- Whitespace / layout ---------- *)

indent          = "  " ;

req_ws          = ( " " | "\t" ), { " " | "\t" } ;
ws              = { " " | "\t" | line_end } ;

line_end        = "\r\n" | "\n" ;

EOF             = ? end of input ? ;

(* ---------- Character classes ---------- *)

UPPER           = "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J"
                | "K" | "L" | "M" | "N" | "O" | "P" | "Q" | "R" | "S" | "T"
                | "U" | "V" | "W" | "X" | "Y" | "Z" ;

LETTER          = UPPER
                | "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j"
                | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t"
                | "u" | "v" | "w" | "x" | "y" | "z" ;

DIGIT           = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;

NONZERO         = "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
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
