# VibeOps
Templating and configuration

## Overview

This repository provides a Go-based templating system that processes template files in the `source` folder and generates configuration files in the `build` folder.

## Usage

### Prerequisites

- Go 1.24 or later
- Make

### Configuration

1. Rename `values.json.example` to `values.json` and set the values for your templates:

```json
{
  "PoppitListName": "example-list",
  "RedisPassword": "example-redis-password",
  "OrgName": "its-the-vibe",
  "BaseDir": "/home/user/projects",
  "SlackWebhookSecret": "example-slack-webhook-secret",
  "GithubWebhookSecret": "example-github-webhook-secret",
  "SlackBotToken": "example-slack-bot-token"
}
```

2. (Optional) Rename `ports.json.example` to `ports.json` to manage port mappings:

```json
{
  "github-webhook-port": 8081,
  "SlackRelayPort": 8082
}
```

The `ports.json` file is optional and allows you to centrally manage port mappings for your web projects. When present, the port values will be merged into your template values and can be referenced in templates.

**Note:** For port names with hyphens (like `github-webhook-port`), use the `index` function in templates:
```
{{ index . "github-webhook-port" }}
```

For port names without hyphens (like `SlackRelayPort`), use the standard syntax:
```
{{ .SlackRelayPort }}
```

If `ports.json` doesn't exist, the templating process will work normally without the port mappings.

3. Rename `projects.json.example` to `projects.json` to define your projects:

```json
[
  {
    "name": "MyProject",
    "allowVibeDeploy": true,
    "isDockerProject": true,
    "useWithSlackCompose": true,
    "useWithGitHubIssue": true
  }
]
```

The `projects.json` file defines all projects in your organization. Each project can have the following properties:
- `name` (required): The project/repository name
- `allowVibeDeploy` (optional, default: true): Whether the project can be deployed via VibeDeploy
- `isDockerProject` (optional, default: true): Whether to use Docker commands (git pull, docker compose build/down/up)
- `buildCommands` (optional): Custom build commands (used when `isDockerProject` is false)
- `useWithSlackCompose` (optional, default: true): Include in SlackCompose project list
- `useWithGitHubIssue` (optional, default: true): Include in GitHub issue integration

This file is used to generate configuration files for SlackCompose, github-dispatcher, and OctoCatalog.

### Running the Templating Process

To process all template files and generate configuration files:

```bash
make template
```

Or use the CLI directly:

```bash
./vibeops template
```

This command will:
1. Build the templating program
2. Read values from `values.json`
3. Process all `.tmpl` files in the `source` folder
4. Generate output files in the `build` folder (without the `.tmpl` extension)

You can specify a custom build directory:

```bash
./vibeops template --build-dir /path/to/custom/build
```

### Creating Symlinks

To create symlinks from the build directory to the `BaseDir` specified in `values.json`:

```bash
make link
```

Or use the CLI directly:

```bash
./vibeops link
```

You can specify a custom build directory:

```bash
./vibeops link --build-dir /path/to/custom/build
```

### Adding a New Project

To add a new project to the configuration:

```bash
./vibeops new-project [project-name]
```

This command will:
1. Add the project to `projects.json` with default settings:
   - `allowVibeDeploy`: true
   - `isDockerProject`: true (with default Docker commands)
   - `useWithSlackCompose`: true
   - `useWithGitHubIssue`: true
2. Create a project directory at `source/its-the-vibe/[project-name]`
3. Create an empty `.env.tmpl` file in the project directory

After adding a project, run `./vibeops template` to generate all configuration files. The following files will be automatically generated from `projects.json`:
- `build/its-the-vibe/SlackCompose/projects.json`
- `build/its-the-vibe/github-dispatcher/config.json`
- `build/its-the-vibe/OctoCatalog/catalog.json`

Example:
```bash
./vibeops new-project MyNewService
./vibeops template
```

To customize project settings, edit `projects.json` directly. See `projects.json.example` for available options.

### Detecting and Restarting Changed Services

To compare configuration changes between builds and automatically restart affected services:

```bash
make diff
```

Or use the CLI directly:

```bash
./vibeops diff
```

This command will:
1. Compare the `prev-build` and `build` directories using `diff -qr`
2. Extract service names from changed files
3. Send restart requests to the TurnItOffAndOnAgain service for each unique service
4. If TurnItOffAndOnAgain itself changed, restart it first with a configurable wait time

Before running the diff command, you need to:

1. Copy `config.json.example` to `config.json` and configure it:

```json
{
  "TurnItOffAndOnAgainUrl": "http://localhost:8080",
  "RestartWaitSeconds": 5
}
```

2. Create a `prev-build` directory with the previous build state:

```bash
make prev-build
```

3. Make changes to your templates or configuration, then generate new build:

```bash
make template
```

4. Run the diff command to restart changed services:

```bash
make diff
```

You can specify a custom config file:

```bash
./vibeops diff --config /path/to/config.json
```

### Validating JSON Configuration Files

To validate all JSON configuration files and ensure they are properly formatted:

```bash
make validate-json
```

Or use the CLI directly:

```bash
./vibeops validate
```

This command will:
1. Validate `values.json` (required)
2. Validate `ports.json` (optional)
3. Validate `projects.json` (required)
4. Validate `config.json` (optional)
5. Display clear error messages for any invalid JSON files

If any file contains invalid JSON, the command will:
- Show the specific file with the error
- Display the error message with line/column information if available
- Provide a suggestion to check JSON syntax
- Exit with a non-zero status code

This is useful to run before deploying or after making manual changes to configuration files.

### Other Commands

Build the templating program only:
```bash
make build
```

Clean up generated files:
```bash
make clean
```

View all available commands:
```bash
./vibeops --help
```

## Directory Structure

- `source/` - Contains template files (`.tmpl` extension)
- `build/` - Generated configuration files (created automatically, not in source control)
- `prev-build/` - Previous build state for comparison (created automatically, not in source control)
- `values.json` - Values to be applied to templates (gitignored, use `values.json.example` as template)
- `projects.json` - Project definitions (gitignored, use `projects.json.example` as template)
- `ports.json` - Optional port mappings to be merged with values (gitignored, use `ports.json.example` as template)
- `config.json` - Configuration for the diff command (gitignored, use `config.json.example` as template)
- `cmd/` - Command implementations (template, link, new-project, diff, validate)
- `internal/utils/` - Shared utility functions
- `main.go` - Main application entry point
- `Makefile` - Build and run commands

## Error Handling

VibeOps provides clear, actionable error messages when JSON parsing or validation fails:

- **File Not Found**: Indicates which required file is missing and suggests checking file permissions
- **Invalid JSON Syntax**: Shows the specific file and error with suggestions to check JSON formatting
- **Validation Failures**: The `validate` command checks all JSON files and reports any issues

All JSON files generated by VibeOps are automatically validated to ensure they are well-formed before being written to disk.
