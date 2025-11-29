# bosun Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-29

## Active Technologies
- N/A (in-memory config) (002-label-parser-merger)
- Go 1.24.6 + Cobra (CLI), Docker SDK, go-units, existing config packages (003-cli-config-validate)
- N/A (read-only command, no persistence) (003-cli-config-validate)

- Go 1.24.6 + Standard library (`reflect`, `strings`, `strconv`, `time`), `github.com/docker/go-units` (for byte size parsing) (001-config-schema)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24.6

## Code Style

Go 1.24.6: Follow standard conventions

## Recent Changes
- 003-cli-config-validate: Added Go 1.24.6 + Cobra (CLI), Docker SDK, go-units, existing config packages
- 002-label-parser-merger: Added Go 1.24.6

- 001-config-schema: Added Go 1.24.6 + Standard library (`reflect`, `strings`, `strconv`, `time`), `github.com/docker/go-units` (for byte size parsing)

<!-- MANUAL ADDITIONS START -->
## Serena-first workflow (MANDATORY)

**Always do this at the start of any task, issue, PR review, or chat reply:**
1) Activate Serena on this repo/project → `serena.activate(project="bosun")`.
2) List memories → `serena.memories.list()` and read the most relevant.
3) Update or create memories as needed to save relevant context for future reference.
4) Prefer Serena's navigation/edit tools for all code work.
5) Avoid terminal tools unless there is no other option. **DO NOT USE TERMINAL COMMANDS LIKE READING FILES, SEARCHING, OR EDITING FILES.**


**Tool policy**
- Primary: `serena` (code navigation/edits, context).
- Secondary: `context7` (check/lookup updated dependencies or APIs).
- Avoid direct terminal commands unless absolutely necessary; prefer Serena for file ops, search, refactors.
- **NEVER use terminal commands (cat, echo, heredocs) to create or edit files** - use `create_file` or `replace_string_in_file` tools instead.
- Terminal is acceptable for: running tests, builds, git commands, and other non-file-editing operations.
<!-- MANUAL ADDITIONS END -->
