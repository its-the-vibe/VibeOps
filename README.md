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

4. (Optional) Rename `bootstrap.json.example` to `bootstrap.json` to configure GCP Secret Manager integration:

```json
{
  "GCPSecretName": "projects/PROJECT_ID/secrets/SECRET_NAME/versions/latest"
}
```

The `bootstrap.json` file is optional and allows you to load additional template values from Google Cloud Secret Manager. This is useful for managing sensitive configuration values (credentials, API keys, etc.) in a secure, centralized location.

**Configuration:**
- `GCPSecretName`: The full resource name of the GCP secret to load (e.g., `projects/my-project/secrets/vibeops-secrets/versions/latest`)

**Requirements:**
- The secret in GCP Secret Manager must contain valid JSON
- The JSON will be parsed and merged with values from `values.json` and `ports.json`
- GCP Secret values will override any conflicting keys from local files
- Your environment must have proper GCP credentials configured (via Application Default Credentials or service account)

**Example GCP Secret Content:**
```json
{
  "SlackBotToken": "xoxb-actual-token",
  "RedisPassword": "secure-password",
  "GithubWebhookSecret": "actual-webhook-secret"
}
```

If `bootstrap.json` doesn't exist or `GCPSecretName` is empty, the templating process will work normally using only local values.

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
2. Create a project directory at `source/__.OrgName__/[project-name]` (using `OrgName` from `values.json`)
3. Create an empty `.env.tmpl` file in the project directory

After adding a project, run `./vibeops template` to generate all configuration files. The following files will be automatically generated from `projects.json`:
- `build/[OrgName]/SlackCompose/projects.json`
- `build/[OrgName]/github-dispatcher/config.json`
- `build/[OrgName]/OctoCatalog/catalog.json`

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

#### Dry-Run Mode

To preview what services would be restarted without making any changes:

```bash
./vibeops diff --dry-run
```

Or using the short flag:

```bash
./vibeops diff -n
```

In dry-run mode, the command will:
- Display which services have changed
- Show which services would be restarted
- Not send any restart requests to TurnItOffAndOnAgain
- Not modify any files or state
- Clearly indicate that it is a dry-run and no changes were made

This is useful for verifying changes before applying them.

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
- `bootstrap.json` - Optional bootstrap configuration for GCP Secret Manager (gitignored, use `bootstrap.json.example` as template)
- `config.json` - Configuration for the diff command (gitignored, use `config.json.example` as template)
- `cmd/` - Command implementations (template, link, new-project, diff, validate)
- `internal/utils/` - Shared utility functions
- `main.go` - Main application entry point
- `Makefile` - Build and run commands

## Error Handling

VibeOps provides clear, actionable error messages when JSON parsing or validation fails:

- **File Not Found**: Indicates which file is missing and suggests ensuring the file exists
- **Permission Errors**: Shows specific file and suggests checking file permissions
- **Invalid JSON Syntax**: Shows the specific file and error with line/column information when available, plus suggestions to check JSON formatting
- **Validation Failures**: The `validate` command checks all JSON files and reports any issues
- **GCP Secret Manager Errors**: If configured, shows clear errors when secrets cannot be loaded, including authentication issues or invalid secret names

All JSON files generated by VibeOps use Go's standard JSON marshaling, which guarantees well-formed output.

## GCP Secret Manager Integration

VibeOps supports loading template values from Google Cloud Secret Manager for secure management of sensitive configuration. This is configured via the optional `bootstrap.json` file.

### Prerequisites

To use GCP Secret Manager integration:

1. **GCP Project Setup**: You need a GCP project with Secret Manager API enabled
2. **Authentication**: Your environment must have GCP credentials configured through one of:
   - Application Default Credentials (recommended for production)
   - Service account key file (set `GOOGLE_APPLICATION_CREDENTIALS` environment variable)
   - GCP CLI authentication (`gcloud auth application-default login` for local development)

### Creating a Secret in GCP

1. Navigate to [Secret Manager in GCP Console](https://console.cloud.google.com/security/secret-manager)
2. Create a new secret with JSON content containing your template values
3. Note the full resource name (e.g., `projects/my-project/secrets/vibeops-secrets/versions/latest`)

Example secret content:
```json
{
  "SlackBotToken": "xoxb-your-actual-token",
  "RedisPassword": "secure-redis-password",
  "GithubWebhookSecret": "actual-webhook-secret"
}
```

### Configuring VibeOps

1. Copy `bootstrap.json.example` to `bootstrap.json`
2. Set the `GCPSecretName` to your secret's full resource name
3. Run `./vibeops template` as normal

The secret values will be automatically loaded and merged with your local configuration, with GCP values taking precedence over local values.

### Security Best Practices

- Never commit `bootstrap.json` to version control (it's gitignored by default)
- Use separate secrets for different environments (dev, staging, production)
- Rotate secrets regularly
- Use least-privilege IAM roles (Secret Manager Secret Accessor role is sufficient)
- Consider using secret versions for rollback capability
