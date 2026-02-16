# EnvLLM-DSL Syntax Guide

EnvLLM-DSL is a strict, typed language for controlling an AI environment. It is designed to be deterministic and error-proof.

## 1. Structure
A program consists of one or more `CELL` blocks. Each cell contains a sequence of operations.

```text
CELL <name>:
  <OpName> <Args...> INTO <OutputVar>: <Type>
  <OpName> <Args...> INTO <OutputVar>: <Type>
```

**Rules:**
*   **Indentation:** Exactly 2 spaces.
*   **One Operation Per Line.**
*   **Explicit Output:** Every operation must end with `INTO <var>: <Type>`.
*   **No Shadowing:** Never reuse a variable name. Create new names (e.g., `text_1`, `text_2`).

## 2. Types
*   **`TEXT`**: A string of text.
*   **`INT`**: A whole number.
*   **`OFFSET`**: A position in a text (opaque). Cannot be used as INT.
*   **`SPAN`**: A range `{start: OFFSET, end: OFFSET}`.
*   **`JSON`**: A structured object.
*   **`STRUCT`**: A typed record (e.g., `VmStats`).

## 3. Operations
All operations use `KEYWORD` arguments.

### Text & Search
*   `FIND_TEXT SOURCE <TEXT> NEEDLE <TEXT> MODE <FIRST|LAST> IGNORE_CASE <BOOL> INTO <OFFSET>`
*   `FIND_REGEX SOURCE <TEXT> PATTERN <TEXT> MODE <FIRST|LAST> INTO <SPAN>`
*   `WINDOW_TEXT SOURCE <TEXT> CENTER <OFFSET> RADIUS <INT> INTO <TEXT>`
*   `SLICE_TEXT SOURCE <TEXT> START <OFFSET> END <OFFSET> INTO <TEXT>`

### Data Conversion
*   `TO_TEXT VALUE <Any> INTO <TEXT>`
*   `JSON_PARSE SOURCE <TEXT> INTO <JSON>`
*   `OFFSET_ADD OFFSET <OFFSET> AMOUNT <INT> INTO <OFFSET>`

### Structs & JSON
*   `GET_FIELD SOURCE <STRUCT> FIELD <String> INTO <Any>` (For typed structs)
*   `JSON_GET SOURCE <JSON> PATH <String> INTO <JSON>` (For raw JSON)

## 4. Example

**Task:** Extract the user ID from a log line "User: 12345 logged in".

```text
CELL extract_user:
  FIND_TEXT SOURCE PROMPT NEEDLE "User: " MODE FIRST IGNORE_CASE false INTO prefix_pos: OFFSET
  OFFSET_ADD OFFSET prefix_pos AMOUNT 6 INTO id_start: OFFSET
  FIND_TEXT SOURCE PROMPT NEEDLE " " MODE FIRST IGNORE_CASE false INTO space_pos: OFFSET
  SLICE_TEXT SOURCE PROMPT START id_start END space_pos INTO id_text: TEXT
  SET_FINAL SOURCE id_text
```
