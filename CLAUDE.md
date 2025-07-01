# Claude Code Instructions for Orochi Project

## Linting Standards

This project uses golangci-lint v1.62.2 to maintain code quality. To ensure consistency between local development and CI:

### Before committing code:
1. Always run `make lint` to check for linting issues
2. The Makefile will automatically install the correct version of golangci-lint if not present

### Common lint issues to watch for:
- Unused parameters in functions - rename to `_` if intentionally unused
- Unused imports - remove them
- Error returns not checked - always check errors
- Exported types/functions without comments - add comments starting with the name

### CI/Local Environment Consistency:
- CI uses golangci-lint v1.62.2 with Go 1.23
- The Makefile ensures the same version is used locally
- Configuration is in `.golangci.yml`

### Pre-commit hooks (optional):
Run `make dev-tools` to install pre-commit hooks that will automatically check your code before each commit.

## Testing Commands

- `make test` - Run all tests with race detection
- `make coverage` - Generate coverage report
- `make lint` - Run linter (same as CI)
- `make build` - Build the binary with UI

## Common Development Tasks

### Adding a new feature:
1. Write tests first (TDD approach)
2. Implement the feature
3. Run `make test` to ensure tests pass
4. Run `make lint` to check code quality
5. Commit changes

### Before pushing to GitHub:
Always run:
```bash
make test
make lint
```

This ensures your code will pass CI checks.