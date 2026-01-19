# Repository Guidelines

## Project Structure & Module Organization
- `cmd/gostore/` contains the CLI entrypoint and wiring for the `gostore` binary.
- `internal/` holds the core application, infrastructure, and CLI command implementations.
- `cmd/tests/` includes higher-level test suites and shared test helpers.
- `data/`, `dist/`, and `bin/` are used for runtime data and build artifacts (keep generated output out of source control unless required).

## Build, Test, and Development Commands
- `go build -v -o ./bin/gostore ./cmd/gostore` builds the CLI binary into `./bin/`.
- `go test ./...` runs all Go tests across the module.
- `golangci-lint run` runs lint checks (configured via `mise.toml`).
- `mise run build` produces a release-style build into `./dist/gostore`.
- `mise run test` runs Go tests plus lint tasks.

## Coding Style & Naming Conventions
- Go formatting is expected; use `gofmt` (tabs for indentation).
- Keep packages small and focused under `internal/` (e.g., `internal/gostore/app/...`).
- Name tests with the Go standard `*_test.go` suffix; place CLI-level suites in `cmd/tests/`.

## Testing Guidelines
- Primary tests run through `go test ./...`.
- Favor table-driven tests and the `testify` assertions already in `go.mod` when useful.
- When adding new CLI behavior, include or extend suites in `cmd/tests/`.

## Commit & Pull Request Guidelines
- Commit messages follow a lightweight conventional style like `feat: ...`, `fix: ...`, or `version-increment: ...`.
- PRs should describe the change, list test coverage (commands run), and note any behavior changes to the CLI.

## Security & Configuration Tips
- Do not commit real secrets; gostore stores are typically created under `~/.gostore/`.
- If a change touches encryption or storage, document migration or compatibility notes in the PR.
