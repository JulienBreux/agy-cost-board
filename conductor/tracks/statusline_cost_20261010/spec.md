# Specification: Antigravity CLI Statusline Spend Integration

## 1. Overview
Provide an integration between `agy-cost-board` and the Antigravity CLI (`agy`) status bar. The server serves a customized fork of the official Antigravity `statusline.sh` script via `GET /statusline.sh`. When installed by developers, this script automatically displays their AI spend at the end of the statusline (e.g., ` · cost $123`), featuring zero-latency background caching, intelligent user identity detection, and support for private IAM Cloud Run deployments.

## 2. Functional Requirements

### 2.1 Dynamic Script Serving (`GET /statusline.sh`)
- Expose an HTTP endpoint `GET /statusline.sh` on the `agy-cost-board` server.
- The server dynamically renders the script, embedding:
  - **Base URL:** Auto-detected from incoming HTTP request headers (`X-Forwarded-Proto`, `Host`) with fallback to configured server URL or `AGY_COST_BOARD_URL`.
  - **User Identity:** Configurable via query parameter `?user=...`, with runtime auto-detection fallback.
- Support query parameters:
  - `?user=<email>`: Pre-configure a specific developer email.
  - `?days=<n>`: Pre-configure default calculation window (default: 30).
  - `?ttl=<seconds>`: Pre-configure cache TTL (default: 300).

### 2.2 Statusline Script Execution & Cost Attribution Display
- Fork the official Antigravity CLI statusline script (`https://github.com/google-antigravity/antigravity-cli/blob/main/examples/statusline/statusline.sh`).
- **User Resolution Hierarchy in Bash:**
  1. `AGY_COST_USER` environment variable.
  2. Embedded user email (from download query parameter).
  3. `gcloud config get-value account 2>/dev/null`
  4. `git config user.email 2>/dev/null`
  5. `$USER` system environment variable.
- **Configurable Environment Overrides:**
  - `AGY_COST_BOARD_URL`: Overrides server base URL.
  - `AGY_COST_DAYS`: Time window in days (default: 30).
  - `AGY_COST_CACHE_TTL`: Cache expiration in seconds (default: 300).
  - `AGY_COST_DISABLED`: Set to `true` to disable cost fetching.
- **Cost Formatting:**
  - Renders at the end of the status bar line: ` · cost $123` (or formatted according to currency).
  - Matches the existing statusline styling using `${DOT}${FG_GRAY}cost ${NUM_COLOR}$<amount>${R}`.
  - Gracefully displays ` · cost $--` or hides cost component if server is unreachable or user not found.

### 2.3 Non-Blocking Cache & Cloud Run IAM Authentication
- **Zero-Latency Terminal Execution:**
  - Reads cached cost immediately from a local cache file (e.g., `/tmp/agy_cost_<user_hash>.cache`).
  - If cache is expired or missing, triggers a background asynchronous fetch in a subshell and returns the stale value or placeholder immediately, ensuring the status bar never blocks interactive terminal input.
- **IAM Authentication for Cloud Run:**
  - Automatically attempts to attach `Authorization: Bearer $(gcloud auth print-identity-token 2>/dev/null)` when communicating with IAM-protected Cloud Run instances.

### 2.4 Documentation & Onboarding Integration
- Add documentation in `README.md` and Web UI:
  - One-line command for developers to download and install the statusline:
    ```bash
    curl -sS https://<deployment-url>/statusline.sh -o ~/.config/antigravity/statusline.sh && chmod +x ~/.config/antigravity/statusline.sh
    ```
  - Configuration instructions for `agy` CLI statusline integration.

## 3. Non-Functional Requirements
- **Performance:** Statusline execution overhead < 15ms when reading from local cache.
- **Portability:** Pure POSIX/Bash compatible with macOS and Linux environments, requiring only standard utilities (`curl`, `jq`, `printf`).
- **Reliability:** Network errors, authentication failures, or missing telemetry fail open (no bash crashes, no disruption to `agy` statusline).

## 4. Acceptance Criteria
- [ ] Server exposes `GET /statusline.sh` returning `text/x-shellscript` with dynamically substituted server URL.
- [ ] Script successfully fetches spend from `/api/v1/users/{id}?days={days}`.
- [ ] Script prints ` · cost $123` at the end of `LINE2` in the statusline.
- [ ] Background caching mechanism respects `AGY_COST_CACHE_TTL` (default 300s) with zero terminal latency.
- [ ] Script auto-detects user email via `AGY_COST_USER`, `gcloud`, or `git config`.
- [ ] Script handles IAM authentication tokens for Cloud Run seamlessly.
- [ ] Automated tests validate `/statusline.sh` endpoint, dynamic template variables, and script syntax.
- [ ] `README.md` documents statusline installation and usage.

## 5. Out of Scope
- Modifying the core Antigravity CLI binary itself.
- Custom GUI settings panel inside the React web app.
