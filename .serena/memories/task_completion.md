# What to Do When a Task is Completed

After completing any development task (adding features, fixing bugs, refactoring), follow these steps to ensure code quality and consistency:

1. **Format the code**: Run `make fmt` to apply Go formatting standards
2. **Run static analysis**: Execute `make vet` to check for common issues
3. **Run tests**: Use `make test` to ensure all tests pass
4. **Run pre-commit hooks**: Execute `pre-commit run --all-files` to apply all quality checks
5. **Tidy dependencies**: Run `make tidy` if any dependencies were added/modified
6. **Manual review**: Check that the code follows the established conventions and architecture
7. **Update documentation**: If needed, update TODO.md or add comments

## Commit Guidelines
- Ensure all pre-commit hooks pass before committing
- Write clear, descriptive commit messages
- Follow any commit message conventions (check `.github/instructions/commit-msg.instructions.md` if present)

## Before Pushing
- Run the full test suite locally
- Ensure CI will pass (tests and linting)
- Check that coverage is maintained or improved

This process ensures that all code contributions maintain high quality and consistency with the project's standards.
