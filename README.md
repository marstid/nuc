# nuc

A command-line interface and MCP server for the [Nucleus Security](https://nucleussec.com) vulnerability management platform.

- **`nuc`** — CLI for managing findings, assets, scans, and metrics
- **`nuc-mcp`** — MCP server exposing the same capabilities to AI agents (Claude, Cursor, opencode, etc.)

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap marstid/tap
brew install nuc
brew install nuc-mcp
```

### Prebuilt Binary (Windows / Linux)

Download the latest release from [github.com/marstid/nuc/releases](https://github.com/marstid/nuc/releases):

- **Windows**: `nuc_<version>_windows_amd64.zip` — extract `nuc.exe` / `nuc-mcp_<version>_windows_amd64.zip` — extract `nuc-mcp.exe`
- **Linux**: `nuc_<version>_linux_amd64.tar.gz` or `_arm64` / `nuc-mcp_<version>_linux_amd64.tar.gz` or `_arm64`

### From Source

```bash
go install github.com/marstid/nuc/cmd/nuc@latest
```

For the MCP server:

```bash
go install github.com/marstid/nuc/cmd/nuc-mcp@latest
```

### Build Locally

```bash
git clone https://github.com/marstid/nuc.git
cd nuc
make build
```

Binaries will be at `bin/nuc` and `bin/nuc-mcp`.

## Configuration

### Create an API key

1. Log in to your Nucleus Security instance (e.g. `https://nucleus-eu6.nucleussec.com`)
2. Navigate to **Settings → API Keys**
3. Click **Create API Key** and copy the generated key

### Set your API key

```bash
nuc config set api_key <your-api-key>
```

Or use an environment variable:

```bash
export NUC_API_KEY=<your-api-key>
```

### Set your API base URL (required)

```bash
nuc config set base_url https://nucleus-eu6.nucleussec.com/nucleus/api
```

Or use an environment variable:

```bash
export NUC_BASE_URL=https://nucleus-eu6.nucleussec.com/nucleus/api
```

> **Note**: Each Nucleus instance has a unique URL (e.g. `nucleus-eu6.nucleussec.com`). There is no default — you must configure this before using any API commands.

### Set default project

```bash
nuc config set default_project 42
```

### View configuration

```bash
nuc config list
nuc config path
```

### Configuration file

The config file is stored at:
- **Linux**: `$XDG_CONFIG_HOME/nuc/config.yaml` (default `~/.config/nuc/config.yaml`)
- **macOS**: `~/Library/Application Support/nuc/config.yaml`
- **Windows**: `%AppData%/nuc/config.yaml`

### Priority order

Configuration is resolved in this order (highest priority first):
1. Command-line flags (`--api-key`, `--base-url`, `--project`)
2. Environment variables (`NUC_API_KEY`, `NUC_BASE_URL`, `NUC_PROJECT`)
3. Config file

## Quick Start

```bash
# List all projects
nuc projects list

# Get project details
nuc projects get 42

# Get project risk score
nuc projects riskscore 42
```

## Usage Examples

In Nucleus, **teams and services are modeled as asset groups**. Use `--groups` (findings, metrics, trends) or `--group` (assets) to filter by team or service.

### Discover Available Groups

```bash
# List all asset groups (teams/services) in your project
nuc assets groups list

# Just the group names (for scripting)
nuc assets groups list -q
```

### Findings per Team/Service

```bash
# All findings for the "payment-service" group
nuc findings search --groups payment-service

# All findings for "backend-team"
nuc findings search --groups backend-team

# Multiple groups at once
nuc findings search --groups payment-service,auth-service

# Use glob patterns to match group names
nuc findings search --groups "*team-euc*"
nuc findings search --groups "*payment*,*auth*"
nuc metrics groups --groups "*backend*"
```

### Combine Severity + Group

```bash
# Critical findings for a specific service
nuc findings search --groups payment-service --severity Critical

# High severity in the backend team
nuc findings search --groups backend-team --severity High
```

### Findings by Status

```bash
# Active (unresolved) findings for a service
nuc findings search --groups payment-service --status Active

# Accepted-risk findings
nuc findings search --groups backend-team --status "Accepted Risk"

# Search by CVE across all groups
nuc findings search --cve CVE-2024-1234
```

### Exploitable Findings

```bash
# Exploitable findings for a group (1=yes, 0=no)
nuc findings search --groups payment-service --exploitable 1

# Exploitable + Critical severity
nuc findings search --exploitable 1 --severity Critical
```

### Assets per Group

```bash
# List assets in a specific team/service group
nuc assets list --group payment-service

# Filter by asset type within a group
nuc assets list --group backend-team --type Host
```

> **Note**: `assets list` uses `--group` (singular); `findings search`, `metrics groups`, and `findings trend` use `--groups` (plural, comma-separated).

### Metrics per Team/Service

```bash
# Compare risk across teams
nuc metrics groups --groups payment-service,auth-service,frontend

# Specific metrics for a single service
nuc metrics groups --groups payment-service --metrics risk_score,vuln_count_critical,mttr_critical_7d

# Selected security posture metrics
nuc metrics groups --groups backend-team --metrics risk_score,asset_count,vuln_count,vuln_count_critical,vuln_count_high,avg_age_critical,mttr_7d
```

### Vulnerability Trends

```bash
# Discovery trend for a service over time
nuc findings trend --groups payment-service --start-date 2025-01-01 --end-date 2025-06-01
```

### Mitigated Findings

```bash
# Recently mitigated findings
nuc findings mitigated --start-date 2025-05-01

# With pagination
nuc findings mitigated --start 0 --limit 50
```

### Overview

```bash
# Project-wide severity summary
nuc findings overview
```

### Scripting & Pipelines

```bash
# Quiet mode — just IDs for piping
nuc findings search --groups payment-service --severity Critical -q

# JSON output for jq processing
nuc findings search --groups backend-team -o json | jq '.[].finding_number'

# YAML output
nuc metrics groups --groups payment-service -o yaml

# Loop over critical findings per service
for group in payment-service auth-service frontend; do
  echo "=== $group ==="
  nuc findings search --groups "$group" --severity Critical -q
done
```

## Output Formats

By default, `nuc` outputs human-readable tables when connected to a terminal and JSON when piped.

```bash
# Force JSON output
nuc projects list -o json

# Force table output
nuc projects list -o table

# Quiet mode — only print IDs (useful for scripting)
nuc projects list -q
```

## Global Flags

| Flag | Env Var | Description |
|------|---------|-------------|
| `--api-key` | `NUC_API_KEY` | Nucleus Security API key |
| `--base-url` | `NUC_BASE_URL` | API base URL |
| `-p, --project` | `NUC_PROJECT` | Default project ID |
| `-o, --output` | — | Output format: `table`, `json` |
| `-q, --quiet` | — | Only print IDs |

## MCP Server

`nuc-mcp` exposes all Nucleus Security capabilities as [Model Context Protocol](https://modelcontextprotocol.io) tools, resources, and prompts — letting AI agents query findings, triage vulnerabilities, and generate security reports.

### Configuration Priority

The MCP server reuses your `nuc` config file, with environment variables as overrides:

1. **Config file** — set once with `nuc config set api_key` / `nuc config set base_url`
2. **Environment variables** — `NUC_API_KEY` / `NUC_BASE_URL` (override file)
3. **CLI flags** — `--api-key` / `--base-url` (highest priority, for dev/debug)

This means: if you've already configured `nuc`, the MCP server works without any extra setup.

### Transport Modes

| Mode | Use case |
|------|----------|
| `stdio` (default) | Local AI tools (Claude Desktop, opencode, Cursor) |
| `http` | Remote/network deployment |

```bash
# stdio (default)
nuc-mcp

# HTTP on localhost:8080
nuc-mcp --transport=http --addr=localhost:8080
```

### Available Tools (21)

| Tool | Description |
|------|-------------|
| `list_projects` | List all accessible projects |
| `get_project` | Get project details |
| `get_project_risk_score` | Get project risk score |
| `list_findings` | List findings with filters |
| `get_finding` | Get finding details |
| `search_findings` | Advanced search (CVE, exploitability, groups) |
| `update_finding` | Update finding status/severity |
| `bulk_update_findings` | Batch update findings |
| `get_mitigated_findings` | Get mitigated findings |
| `get_finding_trend` | Vulnerability trend over time |
| `get_finding_overview` | Severity distribution summary |
| `get_finding_frameworks` | Compliance frameworks |
| `list_assets` | List assets with filters |
| `get_asset` | Get asset details |
| `update_asset` | Update asset properties |
| `list_asset_groups` | List asset groups |
| `get_asset_group_metrics` | Security metrics per group |
| `list_teams` | List teams (filtered asset groups) |
| `list_scans` | List vulnerability scans |
| `get_finding_metrics` | 30/90/180-day metrics |

### Setup with opencode

Add to your `.opencode/config.json`:

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "nuc-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

If you've already run `nuc config set api_key` and `nuc config set base_url`, no env vars needed. To override:

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "nuc-mcp",
      "args": [],
      "env": {
        "NUC_API_KEY": "your-api-key",
        "NUC_BASE_URL": "https://nucleus-eu6.nucleussec.com/nucleus/api"
      }
    }
  }
}
```

### Setup with Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nucleus": {
      "command": "nuc-mcp",
      "args": []
    }
  }
}
```

## Development

### Prerequisites

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/)

### Commands

```bash
make build       # Build both nuc and nuc-mcp
make build-nuc   # Build CLI only
make build-mcp   # Build MCP server only
make test        # Run tests with race detection
make lint        # Run linter
make fmt         # Format code
make vet         # Run go vet
make install      # Install both to $GOPATH/bin
make run-mcp      # Run MCP server over stdio
make run-mcp-http # Run MCP server over HTTP
make clean       # Remove build artifacts
```

## Disclaimer

This project is an independent, community-maintained open source CLI. It is **not affiliated with, endorsed by, or connected to** [Nucleus Security](https://nucleussec.com) in any way. Nucleus Security is a trademark of its respective owners.

## License

MIT — see [LICENSE](LICENSE).
