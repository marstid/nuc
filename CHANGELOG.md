# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.0] - 2026-05-08

### Added
- Add get_findings_summary MCP tool (173514a)
- Add team/service params to asset listing tools (a2b7fa6)
- Add team/service params to finding search tools (4aa99d5)
- Update prompts to use team/service convenience params (ccc1039)

### Other
- Add tests for asset group filter helpers (e06dafb)

## [v0.3.2] - 2026-05-07

### Added
- in_group filter on list_teams MCP tool to filter teams by asset group prefix (3d5db4d)

## [v0.3.1] - 2026-05-06

### Added
- ListTeams API client with Team domain model (97df215)
- list_services MCP tool with optional team filter (054abb3)
- Auto-paginate list_findings, search_findings, and get_mitigated_findings when limit is omitted (2a17a5d)
- Service-specific report option in nucleusReportPrompt (0244afe)
- Two-step alphabetical selection for >20 teams/services in report prompt (0244afe)

### Fixed
- Filter out informational placeholders ("No Vulnerabilities Found", Software List) from report counts (0244afe)
- Add missing doc comments for exported Team type and ListTeams method (08fbd9d)

## [v0.3.0] - 2026-05-06

### Added
- Add MCP default project fallback (eb2f694)

### Changed
- Harden release branch automation (e213fdd)

## [v0.2.7] - 2026-05-05

### Fixed
- Set GH_TOKEN for gh CLI in auto-tag workflow

## [v0.2.6] - 2026-05-05

### Changed
- Trigger release workflow from auto-tag via workflow_dispatch

## [v0.2.5] - 2026-05-05

### Added
- Auto-tag workflow for release branch merges

### Fixed
- Set git identity in auto-tag workflow

## [v0.2.4] - 2026-05-05

### Added
- Auto-tag workflow for release branch merges (79df6f9)

## [v0.2.3] - 2026-05-05

### Fixed
- Exclude gosec G117 globally and remove stale nolint directives (c1ac085)

### Changed
- Upgrade GitHub Actions to Node.js 24 compatible versions (658a3b1)
- Upgrade codecov-action to v6 for Node.js 24 support (25d8fdd)
- Upgrade golangci-lint to v2 in CI (8cf3f92)
