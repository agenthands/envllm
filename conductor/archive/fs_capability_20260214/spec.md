# Track Spec: File System Capability with Path Whitelisting

## Overview
Implement secure file system operations for EnvLLM. This track adds the ability for the model to read and write files, but only within explicitly permitted directory paths defined in the session policy.

## Requirements

1. **Path Whitelisting (`internal/runtime`)**:
    - Update `Policy` to include `AllowedReadPaths` and `AllowedWritePaths` (slices of strings).
    - Implement a path validation helper that ensures a requested file path is contained within one of the whitelisted directories.
    - Prevent directory traversal attacks (e.g., using `..` to escape the whitelist).

2. **File Operations (`internal/ops/capability`)**:
    - **`READ_FILE`**: Reads the content of a file into a `TEXT` handle.
    - **`WRITE_FILE`**: Writes the content of a `TEXT` handle to a file.
    - **`LIST_DIR`**: Lists files in a directory, returning a `JSON` list.

3. **Capability Gating**:
    - Operation `READ_FILE` requires the "fs_read" capability.
    - Operation `WRITE_FILE` requires the "fs_write" capability.
    - Both must fail if the requested path is not whitelisted, even if the capability is enabled.

4. **Operations Update (`assets/ops.json`)**:
    - Register `READ_FILE`, `WRITE_FILE`, and `LIST_DIR` with their required capabilities and signatures.

## Success Criteria
- A script can successfully read a file from an allowed path.
- A script fails with a `security_error` (or `capability_denied`) if it attempts to write to a path outside the whitelist.
- 90% test coverage for path validation and file I/O operations.
