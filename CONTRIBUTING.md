# Contributing to AGY Cost Board

Thank you for your interest in contributing to **AGY Cost Board (`agy-cost-board`)**! 🎉

`agy-cost-board` is a single-binary cloud solution and developer AI FinOps platform that reconciles Antigravity inference telemetry with Google Cloud Billing BigQuery exports to attribute developer costs and optimize Gemini Enterprise seat allocations. We welcome contributions of all kinds: bug fixes, query optimizations, UI enhancements, documentation, and new features.

---

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please treat all contributors with kindness, empathy, and respect.

---

## How Can I Contribute?

### 1. Reporting Bugs
- Search existing [GitHub Issues](https://github.com/JulienBreux/agy-cost-board/issues) to verify the bug hasn't already been reported.
- Open a new issue with a clear title and description. Include:
  - `agy-cost-board` version / commit hash (`./bin/agy-cost-board --help`).
  - Execution mode (CLI, TUI, `serve`, or Cloud Run container).
  - Diagnostic report output (`agy-cost-board setup --format=json` with project IDs redacted).
  - Exact steps to reproduce the issue.
  - Relevant logs or stack traces.

### 2. Proposing Features & Improvements
For significant architectural changes or new attribution formulas:
1. Open a GitHub Issue or Discussion to outline your proposal.
2. Discuss feasibility, Google Cloud BigQuery compatibility, and user experience.
3. Once aligned, proceed with implementation following our test-driven development workflow.

---

## Development Environment Setup

### Prerequisites
- **Go**: Version 1.27 or newer ([Download Go](https://go.dev/dl/)).
- **Node.js**: Version 22 or newer & npm ([Download Node.js](https://nodejs.org/)).
- **Make**: Standard build automation tool.
- *(Optional)* Google Cloud SDK (`gcloud`) with active BigQuery credentials.
  - **No GCP access? No problem!** `agy-cost-board` includes a complete **Synthetic Demo Engine** (`--demo`) providing 30 days of realistic multi-user/multi-model telemetry and billing records for 100% offline development.

### Quick Setup Steps

1. **Clone the Repository**:
   ```bash
   git clone git@github.com:JulienBreux/agy-cost-board.git
   cd agy-cost-board
   ```

2. **Build Web Frontend SPA**:
   ```bash
   cd web
   npm ci
   npm run build
   cd ..
   ```

3. **Build the Binary**:
   ```bash
   make build
   ```

4. **Run in Demo Mode** (Default for offline development):
   ```bash
   ./bin/agy-cost-board serve --demo --port 8080
   ```
   Access the dashboard at **[http://localhost:8080](http://localhost:8080)**.

---

## Engineering & Architectural Principles

All code submitted to `agy-cost-board` must adhere to these foundational principles:

### 1. Single-Binary Self-Contained Architecture
- The compiled React 19 SPA is embedded directly into the Go binary via Go's `embed.FS` (`web/dist/`).
- The binary has **zero external runtime dependencies** (no Node.js or Python required in production).
- Compliant with Google Cloud Run dynamic `$PORT` binding and graceful signal shutdowns (`SIGINT`, `SIGTERM`).

### 2. Strict Mathematical Attribution Reconciliation
- Cost attribution must strictly reflect the proportional formula:
  $$\text{User Cost} = \frac{\text{User Tokens}}{\text{Project Model Tokens}} \times \text{Net Invoiced SKU Cost}$$
- Ensure rounding differences are handled deterministically without losing cents or dropping records.

### 3. Thread-Safe Caching & Query Safety
- Prevent redundant BigQuery scan charges by utilizing the thread-safe in-memory TTL cache layer.
- Never construct SQL queries via untrusted string interpolation; use parameterized queries or strictly validated identifiers.

### 4. Modern Go Standards
- Use `errors.Is(err, ...)` and `errors.As(err, &target)`.
- Use `any` instead of legacy `interface{}`.
- Use standard library `slices` and `maps` packages where appropriate.
- Test concurrency paths with the Go race detector (`go test -race ./...`).

---

## Quality Gates & Testing

Before submitting a pull request, ensure all quality gates pass:

```bash
# Run all unit and integration tests with Go race detector
make test

# Verify code formatting and linting
make lint

# Clean build
make clean && make build
```

---

## Git Workflow & Commit Guidelines

We use **Conventional Commits** to keep git history clean, readable, and machine-parsable for automated changelogs:

### Commit Types
- `feat:` A new feature or user-facing functionality.
- `fix:` A bug fix.
- `docs:` Documentation-only changes (README, architecture docs).
- `refactor:` Code restructuring without changing external behavior.
- `perf:` A code change that improves performance or reduces memory/query costs.
- `test:` Adding or correcting tests.
- `chore:` Build process, dependency updates, or tool configurations.

### Examples
- `feat(cli): add CSV export flag to license subcommand`
- `fix(attribution): prevent divide-by-zero when project token count is 0`
- `docs: update deployment instructions for Cloud Run`

---

## Submitting a Pull Request (PR)

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/my-new-feature
   ```
2. Implement your changes, adhering to the coding principles and testing requirements above.
3. Verify that `make test` and `make lint` pass with zero warnings or errors.
4. Push your branch to GitHub and open a Pull Request against `JulienBreux/agy-cost-board:main`.
5. Fill out the PR description detailing what problem it solves and how it was tested.
