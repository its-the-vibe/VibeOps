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

To add a new project to the configuration files:

```bash
./vibeops new-project [project-name]
```

This command will:
1. Add the project to `source/its-the-vibe/SlackCompose/projects.json.tmpl`
2. Add the project to `source/its-the-vibe/github-dispatcher/config.json.tmpl` with default commands:
   - `git pull`
   - `docker compose build`
   - `docker compose down`
   - `docker compose up -d`

Example:
```bash
./vibeops new-project MyNewService
```

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
- `values.json` - Values to be applied to templates (gitignored, use `values.json.example` as template)
- `ports.json` - Optional port mappings to be merged with values (gitignored, use `ports.json.example` as template)
- `cmd/` - Command implementations (template, link, new-project)
- `internal/utils/` - Shared utility functions
- `main.go` - Main application entry point
- `Makefile` - Build and run commands
