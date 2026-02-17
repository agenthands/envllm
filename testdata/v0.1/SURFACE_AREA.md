# EnvLLM-DSL v0.1 Surface Area Inventory

## Operations Used
- `STATS`
- `FIND_TEXT`
- `WINDOW_TEXT`
- `JSON_PARSE`
- `JSON_GET`
- `SUBCALL`
- `FIND_REGEX`
- `READ_FILE`
- `WRITE_FILE`
- `LIST_DIR`
- `GET_SPAN_START`
- `GET_SPAN_END`
- `CONCAT`
- `SLICE_TEXT`

## Types / Literals
- `TEXT` (Handle-based or quoted string)
- `INT` (positive and negative)
- `BOOL` (`true`, `false`)
- `JSON` (passed as string to `JSON_PARSE` or returned by ops)
- `SPAN` (returned by `FIND_REGEX`)
- `NULL` (`null`)

## Syntax Variations / Aliases
- **Op Names**: Always UPPERCASE.
- **Keywords**: Always UPPERCASE.
- **Indentation**: 
  - Canonical: 2 spaces.
  - Variations: 0 spaces (should be rejected in STRICT, accepted in COMPAT).
- **Assignments**: Mandatory `INTO <ident>`.
- **String Escapes**: `
`, `	`, ``, `"`, `` supported.
- **Expressions**: Literals or Identifiers. No complex nesting (`+` allowed only via `CONCAT` op).

## Special Statements
- `SET_FINAL SOURCE <expr>`
- `ASSERT COND <expr> MESSAGE <string>`
- `PRINT SOURCE <expr>`
- `FOR_EACH <ident> IN <ident> LIMIT <int>:`
- `IF <expr>:` / `ELSE:` / `END`

## Block Structure
- `CELL <name>:` followed by lines.
- No explicit `CELL_END` (inferred by EOF or next `CELL`).
