# EnvLLM Operations Reference (v0.2)

This reference defines all operations available in the standard modules.

## Core Module (`core`)
*Capabilities: `pure`, `llm`*

| Operation | Signature | Returns | Description |
| :--- | :--- | :--- | :--- |
| **STATS** | `SOURCE <TEXT>` | `JSON` | Returns `{bytes, lines}` of the source text. |
| **FIND_TEXT** | `SOURCE <TEXT> NEEDLE <TEXT> MODE <enum> IGNORE_CASE <BOOL>` | `INT` | Finds index of substring. Mode: `FIRST` or `LAST`. |
| **WINDOW_TEXT** | `SOURCE <TEXT> CENTER <INT> RADIUS <INT>` | `TEXT` | Returns text around `CENTER` +/- `RADIUS`. |
| **SLICE_TEXT** | `SOURCE <TEXT> START <INT> END <INT>` | `TEXT` | Returns text substring `[START, END)`. |
| **FIND_REGEX** | `SOURCE <TEXT> PATTERN <TEXT> MODE <enum>` | `SPAN` | Finds regex match. Mode: `FIRST` or `LAST`. Returns `{start, end}`. |
| **JSON_PARSE** | `SOURCE <TEXT>` | `JSON` | Parses string content into a JSON object/array. |
| **JSON_GET** | `SOURCE <JSON> PATH <TEXT>` | `JSON` | Gets nested value using "key.subkey" path. |
| **GET_SPAN_START** | `SOURCE <SPAN>` | `INT` | Returns the start index of a span. |
| **GET_SPAN_END** | `SOURCE <SPAN>` | `INT` | Returns the end index of a span. |
| **CONCAT** | `A <TEXT> B <TEXT>` | `TEXT` | Concatenates two text values. |
| **SUBCALL** | `SOURCE <TEXT> TASK <TEXT> DEPTH_COST <INT>` | `JSON` | Recursively calls the agent on `SOURCE` with `TASK`. |

## Filesystem Module (`fs`)
*Capabilities: `fs_read`, `fs_write`*

| Operation | Signature | Returns | Description |
| :--- | :--- | :--- | :--- |
| **READ_FILE** | `PATH <TEXT>` | `TEXT` | Reads file content. Path must be whitelisted. |
| **WRITE_FILE** | `PATH <TEXT> SOURCE <TEXT>` | `BOOL` | Writes content to file. Path must be whitelisted. |
| **LIST_DIR** | `PATH <TEXT>` | `JSON` | Lists filenames in a directory. |

## Web Module (`web`)
*Capabilities: `web.navigate`, `web.dom.query`*

| Operation | Signature | Returns | Description |
| :--- | :--- | :--- | :--- |
| **NAVIGATE** | `URL <TEXT>` | `BOOL` | Navigates browser to URL. |
| **CLICK** | `SELECTOR <TEXT>` | `BOOL` | Clicks an element matching CSS selector. |
| **TYPE** | `SELECTOR <TEXT> TEXT <TEXT>` | `BOOL` | Types text into an input field. |
