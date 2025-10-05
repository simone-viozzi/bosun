# Copilot Instructions

This directory contains path-specific instructions for GitHub Copilot coding agent. These files provide context-aware guidance that automatically applies when working with specific file types.

## How It Works

Each instruction file uses YAML frontmatter with an `applyTo` field that specifies which files it applies to using glob patterns. When you edit a matching file, Copilot automatically loads the relevant instructions.

## Available Instructions

| File | Applies To | Purpose |
|------|-----------|---------|
| `go-files.instructions.md` | `**/*.go` (excluding tests) | Go code style, architecture patterns, error handling |
| `test-files.instructions.md` | `**/*_test.go` | Testing patterns, AAA structure, integration tests |
| `makefile.instructions.md` | `**/Makefile` | Build target conventions, Make best practices |
| `dockerfile.instructions.md` | `**/Dockerfile` | Multi-stage builds, security, optimization |
| `docker-compose.instructions.md` | `**/docker-compose.{yml,yaml}` | Testing patterns, label conventions |
| `commit-msg.instructions.md` | `COMMIT_EDITMSG`, `MERGE_MSG` | Conventional commits format |
| `issue-implementation.instructions.md` | (general) | Workflow for implementing issues |

## Adding New Instructions

To add path-specific instructions:

1. Create a new `.md` file in this directory
2. Add YAML frontmatter with the `applyTo` glob pattern:
   ```yaml
   ---
   applyTo: "**/*.yaml,**/*.yml"
   ---
   ```
3. Write clear, actionable guidance below the frontmatter
4. Test by editing a matching file and verifying Copilot receives the context

## Glob Pattern Examples

- `**/*.go` - All Go files recursively
- `!**/*_test.go` - Exclude test files (use with negation)
- `**/Makefile` - Any Makefile in any directory
- `*.json` - JSON files in root only
- `**/*.{yml,yaml}` - YAML files with either extension

## Best Practices

- Keep instructions focused and actionable
- Use code examples to illustrate patterns
- Reference project-specific conventions
- Update instructions as patterns evolve
- Test instructions by using Copilot with matching files

## Main Instructions

The main Copilot instructions are in `.github/copilot-instructions.md` and apply to all files. Path-specific instructions supplement these with file-type-specific guidance.

## Validation

To verify the instructions are working:
1. Edit a file matching a pattern (e.g., a `.go` file)
2. Check that Copilot provides relevant suggestions based on the instructions
3. Run the copilot-setup-steps workflow to ensure environment setup works

## References

- [GitHub Copilot for Business Documentation](https://docs.github.com/en/copilot)
- [Best practices for Copilot coding agent](https://gh.io/copilot-coding-agent-tips)
