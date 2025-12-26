# GitHub Copilot Instructions for VibeOps

## Project Overview

VibeOps is a Go-based templating system that processes template files in the `source` folder and generates configuration files in the `build` folder. The system uses Go's text/template package to apply values from `values.json` to `.tmpl` files.

## Build and Test Instructions

### Building the Project

```bash
make build
```

This compiles the `main.go` file into a `vibeops` binary.

### Running the Templating Process

```bash
make template
```

This builds the binary and processes all `.tmpl` files in the `source` directory, generating output files in the `build` directory.

### Clean Up

```bash
make clean
```

Removes the `build` directory and the `vibeops` binary.

### Testing

Currently, there is no test infrastructure in this repository. When adding tests:
- Use Go's standard testing framework (`testing` package)
- Place test files alongside the code they test with `_test.go` suffix
- Run tests with `go test -v ./...`

## Coding Standards and Conventions

### Go Language

- **Go Version**: This project uses Go 1.24 or later (currently 1.25.5)
- **Formatting**: Follow standard Go formatting (`gofmt`)
- **Error Handling**: Always check and handle errors appropriately
- **Naming**: Use idiomatic Go naming conventions (camelCase for unexported, PascalCase for exported)

### Project Structure

- `main.go` - Main application logic containing:
  - `loadValues()` - Reads and parses values.json
  - `processTemplates()` - Walks source directory and processes .tmpl files
  - `processTemplateFile()` - Processes individual template files
  - `createSymlinks()` - Creates symlinks from build directory to BaseDir
- `source/` - Contains template files with `.tmpl` extension
- `build/` - Auto-generated directory (gitignored, not in source control)
- `values.json` - User configuration (gitignored, use `values.json.example` as template)
- `Makefile` - Build automation

### Template Processing

- Template files must have `.tmpl` extension
- Templates use Go's `text/template` syntax with `{{ .VariableName }}` placeholders
- Output files are generated in `build/` with the `.tmpl` extension removed
- Directory structure from `source/` is preserved in `build/`

### Values Configuration

The `values.json` file should contain:
- `PoppitListName` - Name of the Poppit list
- `RedisPassword` - Redis password
- `OrgName` - Organization name (typically "its-the-vibe")
- `BaseDir` - Base directory for symlink targets
- `SlackWebhookSecret` - Slack webhook secret
- `GithubWebhookSecret` - GitHub webhook secret
- `SlackBotToken` - Slack bot token

## Important Notes for Copilot

1. **Never commit sensitive data**: The `values.json` file contains secrets and should never be committed. Always use `values.json.example` as a reference.

2. **Preserve .gitignore**: The `.gitignore` file excludes:
   - Build artifacts (`build/`, `vibeops`)
   - Sensitive files (`values.json`)
   - Go binaries and test artifacts

3. **Minimal changes**: This is a simple, focused templating tool. Avoid adding unnecessary complexity or dependencies.

4. **Error messages**: Provide clear, actionable error messages that help users understand what went wrong and how to fix it.

5. **File permissions**: Generated files and directories should use appropriate permissions (0755 for directories, default for files).

## Common Tasks

### Adding a new template file
1. Create a `.tmpl` file in the `source/` directory
2. Use `{{ .VariableName }}` syntax for placeholders
3. Run `make template` to generate the output

### Modifying template processing
- Edit the `processTemplateFile()` function in `main.go`
- Ensure error handling is comprehensive
- Test with various template files

### Adding new configuration values
1. Update `values.json.example` with the new field
2. Document the new field in README.md
3. Update templates that use the new value
