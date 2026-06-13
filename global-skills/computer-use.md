---
name: computer-use
description: Methodology for interacting with the local OS, file system, and Docker containers using shell commands and human-in-the-loop approvals.
category: System
---

# Computer Use & Local OS Interaction

This skill provides a methodology for interacting with the local operating system and file system using the `run_command`, `read_file`, `list_dir`, and `write_file` tools.

## Core Methodology

### 1. Verification Before Execution
- Always use `list_dir` to explore the directory structure before running complex commands or writing files.
- Verify the current working directory or target paths to avoid accidental data loss.

### 2. Precise Shell Interaction
- Use `run_command` for tasks that cannot be accomplished with native file tools (e.g., `git`, `grep`, `find`, `npm`, `go`).
- **Docker Integration**: If configured, you can interact with other containers on the network using the `docker` CLI (e.g., `docker ps`, `docker logs`, `docker exec`).
- **Sudo Permissions**: If a command requires elevated privileges, prefix it with `sudo`. The environment may handle sudo authentication if configured.
- Keep commands concise and avoid long-running processes that might hang the process.
- If a command fails, analyze the output and error message provided by the tool.

### 3. Safety and Permissions
- `run_command`, `write_file`, and `delete_file` are **SENSITIVE** operations that require explicit user approval.
- Do not attempt to bypass security restrictions or access unauthorized paths outside allowed directories.
- Always explain the intent of the command to the user before requesting approval.

### 4. Step-by-Step Execution
- Break down complex OS tasks into small, verifiable steps.
- After running a command that modifies the filesystem, use `list_dir` or `read_file` to verify the result.

## Example Workflows

### Scenario: Investigating a local repository
1. `list_dir(path: "/path/to/project")` to see the structure.
2. `run_command(command: "grep -r 'TODO' /path/to/project")` to find tasks.
3. `read_file(path: "/path/to/project/main.go")` to understand implementation.

### Scenario: Creating a temporary project
1. `run_command(command: "mkdir -p /tmp/my-app")`
2. `write_file(path: "/tmp/my-app/hello.go", content: "...")`
3. `run_command(command: "go run /tmp/my-app/hello.go")`

## Formatting Rules
- Always wrap command outputs in code blocks.
- If output is too long, summarize the key parts.
