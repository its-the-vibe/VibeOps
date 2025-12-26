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
- `values.json` - Values to be applied to templates
- `main.go` - Go templating program
- `Makefile` - Build and run commands
