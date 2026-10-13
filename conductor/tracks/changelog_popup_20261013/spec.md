# Specification: Changelog Popup from Last Visit/Release

## 1. Overview
Introduce a **"What's new in AGY Cost Board"** changelog release popup modal in the React web dashboard based on the provided screenshot reference (`/Users/julienbreux/Pictures/Screenshots/Screenshot 2026-10-09 at 17.16.41.png`). The popup informs users about new features and improvements whenever a new version is released since their last visit, tracking state using a persistent cookie with `localStorage` backup. The modal features a sparkles header, version badges, release dates, bold feature titles with bullet lists, scrollable historical releases, and a "Got it" confirmation button. Users can also manually re-open the popup anytime from the Navbar or Footer.

## 2. Functional Requirements

### 2.1 Changelog Data Model (`web/src/data/changelog.ts`)
- Maintain a structured list of historical releases with:
  - `version`: string (e.g. `v0.5.1`, `v0.5.0`, `v0.4.0`)
  - `date`: string (e.g. `October 13, 2026`)
  - `title`: string (e.g. `Native Currency Support Across Engine, CLI, TUI, Statusline, and Web`)
  - `items`: string[] (bullet points with bold prefix formatting e.g. `**Multi-Currency Engine**: Support for EUR, GBP, JPY...`)
- Seed with release entries matching the actual project history (`v0.5.1`, `v0.5.0`, `v0.4.0`, `v0.3.1`, `v0.3.0`, `v0.2.0`, `v0.1.0`).

### 2.2 Version Tracking Storage (`web/src/utils/changelogStorage.ts`)
- Browser cookie management for `agy_last_seen_version`:
  - `getLastSeenVersion()`: Reads `agy_last_seen_version` cookie, falling back to `localStorage.getItem('agy_last_seen_version')`.
  - `setLastSeenVersion(version: string)`: Writes `agy_last_seen_version=<version>; Max-Age=31536000; Path=/; SameSite=Lax` and mirrors to `localStorage.setItem('agy_last_seen_version', version)`.
- Semver comparison utility:
  - `hasNewVersion(currentVersion: string, lastSeenVersion: string | null)`: Returns `true` if `lastSeenVersion` is null/empty or if `currentVersion` is strictly newer than `lastSeenVersion` by semantic version precedence. Ignores build metadata and cleans semver prefixes.

### 2.3 Changelog Modal Component (`web/src/components/ChangelogModal.tsx`)
- High-fidelity match of the reference screenshot:
  - Container: Elevated dark modal card (`bg-[#161a23] border border-[#232936] text-white rounded-2xl shadow-2xl`), responsive width (max-w-xl), centered with blurred/dimmed backdrop.
  - Header:
    - ✨ Sparkles icon + "What's new in AGY Cost Board" (bold white headline, text-lg or text-xl).
    - Subtitle: "Updates to AGY Cost Board — scroll for earlier versions." (muted gray text, text-sm).
  - Body (scrollable list):
    - Latest / current release:
      - Version badge: Bright blue rounded pill (`bg-blue-600 text-white font-mono text-xs px-2.5 py-0.5 rounded-full font-medium`).
      - Date: Muted gray text alongside badge (`September 24, 2026`).
      - Section headline: Bold feature title (`text-white font-semibold text-sm`).
      - Bullet points: List items with bold lead-ins and clean typography.
    - Historical / earlier releases:
      - Version badge: Muted neutral text or subtle badge (`text-google-gray-400 font-mono text-xs`).
      - Date & headlines: Formatted similarly to provide clear timeline history.
  - Footer:
    - Bottom bar with right-aligned primary blue button "Got it" (`bg-google-blue hover:bg-blue-600 text-white font-medium px-4 py-2 rounded-xl text-sm transition-colors`).
  - Dismissal:
    - Clicking "Got it", clicking outside the modal dialog backdrop, or pressing `Escape` closes the modal and records `setLastSeenVersion(currentVersion)`.
  - Accessibility: `role="dialog"`, `aria-modal="true"`, `aria-labelledby`, focus trap on open.

### 2.4 User Interface Integration
- **Auto-Open on Visit (`web/src/App.tsx`):**
  - When version info is loaded from `/api/v1/version`, evaluate `hasNewVersion(appVersion, lastSeenVersion)`.
  - If `true`, set `showChangelog(true)`.
  - When modal is closed, invoke `setLastSeenVersion(appVersion)`.
- **Navbar Trigger (`web/src/components/Navbar.tsx`):**
  - Add a "What's new" sparkles ✨ icon button in the header (adjacent to LiveSyncBadge/ThemeToggle).
  - Tooltip: "What's new".
  - Clicking manually opens the Changelog modal.
- **Footer Trigger (`web/src/components/Footer.tsx`):**
  - Version label can be clicked to open the Changelog modal directly, while retaining external GitHub release link.

## 3. Non-Functional Requirements
- **No External Dependencies:** Zero runtime calls to GitHub API to prevent rate-limiting or network latency.
- **Performance:** Instant load with zero layout shift or blocking operations.
- **Accessibility:** WCAG AA contrast compliance, keyboard navigability (Escape key dismiss, Tab navigation), clear focus rings.
- **Resilience:** Safe handling if cookies or localStorage are disabled (private browsing).

## 4. Acceptance Criteria
- [ ] Stored version read/write correctly sets cookie `agy_last_seen_version` and localStorage backup.
- [ ] On initial visit with no stored version, the modal opens automatically.
- [ ] On subsequent reload with the same version, the modal does NOT automatically open.
- [ ] When version increments (e.g. v0.5.0 -> v0.5.1), the modal automatically opens on the next visit.
- [ ] Clicking "Got it", backdrop, or Escape closes the modal and saves the version state.
- [ ] Clicking the Sparkles icon in the Navbar opens the modal.
- [ ] Clicking the version badge in the Footer opens the modal.
- [ ] Modal visual styling faithfully reproduces the reference screenshot.
- [ ] 100% passing tests for storage, semver logic, component rendering, and App integration.

## 5. Out of Scope
- Remote Markdown fetching over the network at runtime.
- Interactive user reactions or upvoting on release items.
