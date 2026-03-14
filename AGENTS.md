# AGENTS.md

Guidance for coding agents working in `commit-gen`.

## Project Snapshot
- Language: Go (`go 1.24.4`)
- Module: `github.com/nguyenanhhao221/commit-gen`
- Entrypoint: `main.go`
- Core package: `internal/generator`
- Purpose: Generate commit messages from staged diff + recent git history using Gemini.

## Repository Layout
- `main.go`: CLI flag parsing, `.env` loading, staged-change gate, output.
- `internal/generator/generator.go`: high-level API + AI request construction.
- `internal/generator/git.go`: git command wrappers and commit context collection.
- `README.md`: setup and usage docs.

## Cursor/Copilot Rules
- `.cursor/rules/` not found.
- `.cursorrules` not found.
- `.github/copilot-instructions.md` not found.
- No extra IDE-agent instruction files are currently defined.

## Prerequisites
- Use Go `1.24.4` (or a compatible patch release).
- Ensure `git` is installed.
- Install and authenticate `claude` CLI (default provider).
- Set `GOOGLE_API_KEY` via environment or `.env` when using `gemini` provider.
- Run commands from repo root: `/Users/haonguyen/Code/commit-gen`.

## Build Commands
- Build all packages: `go build ./...`
- Build CLI binary: `go build -o commit-gen ./main.go`
- Run CLI directly: `go run ./main.go`
- Run short output mode: `go run ./main.go -short`
- Verify modules: `go mod verify`
- Sync `go.mod`/`go.sum` when needed: `go mod tidy`

## Test Commands
Current state: no `*_test.go` files exist yet.

- Run all tests: `go test ./...`
- Verbose test run: `go test -v ./...`
- Disable cache: `go test ./... -count=1`
- Coverage summary: `go test ./... -cover`
- Coverage profile: `go test ./... -coverprofile=coverage.out`

### Run a Single Test
- Exact test name across all packages:
  - `go test ./... -run '^TestName$' -v`
- Exact test in one package:
  - `go test ./internal/generator -run '^TestName$' -v`
- Subtest pattern:
  - `go test ./internal/generator -run 'TestName/Subcase' -v`
- Single benchmark:
  - `go test ./internal/generator -run '^$' -bench '^BenchmarkName$'`

## Lint and Static Analysis
No linter config is committed; use Go-native checks first.

- Format code: `gofmt -w .`
- Static checks: `go vet ./...`
- Optional (if installed): `golangci-lint run`
- Optional import normalizing: `goimports -w .`

## Recommended Validation Order
1. `gofmt -w .`
2. `go test ./...`
3. `go vet ./...`
4. `go build ./...`
5. Smoke test CLI:
	 - `go run ./main.go -short`
	 - `GOOGLE_API_KEY=... go run ./main.go -provider gemini -short`

## Code Style Guidelines
Follow existing patterns in `main.go` and `internal/generator/*.go`.

### Imports
- Keep grouped imports: stdlib, third-party, internal module.
- Keep imports sorted (`gofmt`/`goimports` handles this).
- Avoid aliases unless they remove real ambiguity.
- Do not leave unused imports.

### Formatting
- Always apply `gofmt`.
- Keep functions small and focused on one responsibility.
- Use whitespace to separate logical phases of work.
- Avoid unnecessary inline cleverness.

### Types and API Design
- Use clear option structs for constructor configuration (`Options`).
- Validate required inputs early (e.g., API key).
- Prefer concrete structs for internal behavior.
- Use pointers for stateful structs or large payloads.
- Keep defaults centralized (`DefaultConfig`).

### Naming
- Exported names: `PascalCase`.
- Internal names: `camelCase`.
- Package names: lowercase, concise, noun-like.
- Use descriptive identifiers (`recentCommits`, `workingDir`).
- Keep acronyms readable and conventional (`APIKey`, `GitInfo`).

### Error Handling
- Return errors; avoid panic for expected runtime failures.
- Wrap with context using `%w`:
  - `fmt.Errorf("failed to create AI client: %w", err)`
- Use actionable, specific error messages.
- In CLI path, fail fast for unrecoverable errors (`log.Fatalf`).
- Preserve original error chains for callers.

### Context and Timeouts
- Use `context.WithTimeout` for external/network calls.
- `defer cancel()` immediately after context creation.
- Keep timeout values configurable, not scattered magic numbers.

### Git Process Boundaries
- Use `exec.Command` with explicit args; no shell interpolation.
- Set `cmd.Dir` only when `workingDir` is non-empty.
- Trim outputs before emptiness checks.
- Return wrapped errors that describe failed git intent.

### Logging and CLI Output
- Stdout should contain successful command output (commit message).
- Use logs for warnings/errors.
- Keep user-facing CLI text concise and actionable.

### Comments and Documentation
- Document exported types and functions.
- Keep comments short and behavior-focused.
- Remove comments that only restate obvious code.

### Testing Guidance for New Code
- Prefer table-driven tests.
- Cover happy path, fallback path, and error path.
- Add cases for:
  - missing `GOOGLE_API_KEY`
  - no staged changes
  - missing git history fallback behavior
  - full vs short commit generation
- Mock or isolate external boundaries (git and AI calls) when feasible.

## Security and Secrets
- Never commit `.env` or API keys.
- Never hardcode credentials in source.
- Use provider-specific credentials from environment at runtime.

## Change Scope Guidance
- Keep changes minimal and task-focused.
- Preserve existing public API unless change is explicitly intended.
- Update `README.md` if setup, flags, or behavior changes.

## Quick Command Reference
- Full verification: `gofmt -w . && go test ./... && go vet ./... && go build ./...`
- Build binary: `go build -o commit-gen ./main.go`
- Run binary (default claude-cli): `./commit-gen`
- Run short mode (default claude-cli): `./commit-gen -short`
- Run gemini provider: `GOOGLE_API_KEY=... ./commit-gen -provider gemini`
- Single test: `go test ./... -run '^TestName$' -v`
