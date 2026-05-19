# Contributing to ratnosint7

Thank you for contributing! This guide covers pull requests, code standards, and development workflow.

## Before You Start

ratnosint7 is a **security reconnaissance tool**. All contributions must:

1. **Maintain authorization framework** — Never weaken legal/ethical warnings
2. **Preserve performance** — Streaming pipeline architecture is non-negotiable
3. **Pass goroutine leak checks** — No dangling goroutines or channel sends
4. **Handle context cancellation** — All goroutines must respect `ctx.Done()`

## Pull Request Process

1. **Fork** the repository
2. **Create branch**: `git checkout -b feature/your-feature` or `fix/issue-name`
3. **Make changes** (see Code Standards below)
4. **Test**: `go test ./...` and run manual scans
5. **Commit**: Use conventional commits (see Commit Message Format)
6. **Push** to your fork
7. **Open PR** with description of changes + testing performed

## Code Standards

### Go Style
- **Format**: `gofmt` (no custom formatting)
- **Lint**: No unused variables, imports
- **Comments**: Only document WHY non-obvious logic exists. Skip obvious code.
- **Error handling**: Check all errors; use `fmt.Errorf` with context
- **Concurrency**: Respect context deadlines; drain channels on ctx.Done()

### File Headers
All `.go` files require Apache 2.0 header:
```go
// Copyright 2026 Ratnadeep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yourpackage
```

### Performance
- **No buffering**: Stream to disk, never load all results in memory
- **Channels**: Use specified buffer sizes (raw: 64KB, parsed: 32KB, unique: 16KB)
- **Goroutines**: One per plugin; reuse parser/dedupe workers
- **Benchmarks**: Profile before/after: `go test -bench ./...`

### Testing
- **Unit tests**: All public functions
- **Integration**: Test full pipeline on small domain
- **Goroutine leaks**: Use `runtime.NumGoroutine()` to verify cleanup
- **Context cancellation**: Test timeout/cancel paths

## Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Example**:
```
feat(engine): add streaming deduplication with lock-free sharding

Implement 32-shard dedupe map using fnv-1a hashing. Reduces contention
on high-concurrency scans. Benchmarks show 15% throughput improvement.

Fixes #42
```

## Tool Additions

To add a new enumeration tool:

1. **Add entry to `configs/tools.yaml`**:
   ```yaml
   - name: your_tool
     version: v1.0.0
     install: go install github.com/user/your_tool@v1.0.0
     passive_run: your_tool --passive {domain}
     active_run: your_tool --active {domain}
     rate_limit: 5
   ```

2. **Update THIRD_PARTY_TOOLS.md** with GitHub URL + license link

3. **Test**: `./ratnosint7 update-tools` installs it; `./ratnosint7 scan example.com --passive` runs it

4. **Benchmarks**: Report result count + speed vs. existing tools

## Legal Review

All PRs touching **authorization, legal, or ethical logic** require:
- Clear commit message explaining WHY change is safe
- Test case showing compliance maintained
- No removal of warnings or disclaimers

## Reporting Issues

- **Bug**: Include domain tested, scan mode (passive/active), error output
- **Feature**: Explain use case + performance impact
- **Security**: **DO NOT** open public issue. See SECURITY.md

## Questions?

Open a discussion or issue. Maintainer responds within 48h.

---

**Remember**: This tool can cause legal liability if misused. Always test on owned domains first.
