# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Both scan mode** (`--both` / interactive `[3]`): runs passive and active commands concurrently per tool, merges and deduplicates results into a single output — broadest coverage in one scan
- Initial public release of ratnosint7
- 13-tool orchestration engine with concurrent execution
- `update-tools` per-tool progress (`[n/13]`) with explicit success/failure lines
- tugarecon via [skynet0x01/tugarecon](https://github.com/skynet0x01/tugarecon) with local `.venv` (Python 3.12)
- Streaming pipeline architecture (no intermediate buffering)
- Lock-free deduplication with 32 sharded maps
- Result caching with 24h TTL
- Three output formats: txt, json, csv
- Interactive CLI with Cobra framework
- Full test coverage for core pipeline
- Comprehensive error logging to `~/.ratnosint7/logs/`

### Security
- Explicit authorization warnings in README and CLI
- Legal compliance section covering CFAA, GDPR, CCPA
- Goroutine leak detection and fixes
- Context cancellation on timeout/interrupt
- Atomic file writes (crash-safe)

### Documentation
- Comprehensive README with installation, usage, troubleshooting
- THIRD_PARTY_TOOLS.md with tool URLs and licenses
- CONTRIBUTING.md with PR and code standards
- CODE_OF_CONDUCT.md (Contributor Covenant 2.1)
- SECURITY.md with vulnerability disclosure policy
- Inline code comments for non-obvious logic

### Infrastructure
- Issue templates (bug_report.yml, feature_request.yml)
- Go module dependency management (go.mod/go.sum)
- Makefile with build, test, install targets
- Setup script for automated installation

---

## How to Report Issues

- **Bug**: Open GitHub issue with domain, scan mode, error output
- **Feature**: Describe use case + performance impact
- **Security**: See [SECURITY.md](SECURITY.md) for private reporting

## Future Roadmap

- [ ] Resume-able scans (checkpoint-based)
- [ ] Result sorting (after full buffer)
- [ ] Docker distribution
- [ ] Custom plugin development framework
- [ ] Enhanced web/TUI dashboard features
- [ ] Scheduled scan automation
