# Implementation Plan: Changelog Popup from Last Visit/Release

## Phase 1: Changelog Data & Storage Utilities (TDD) [checkpoint: 3787c26]
- [x] Task: Storage & Version Comparison Unit Tests (5f8d5af)
  - [x] Write unit tests in `web/src/utils/changelogStorage.test.ts` for cookie read/write, localStorage fallback, and semver comparison logic (`hasNewVersion`).
- [x] Task: Implement Changelog Data and Storage Helpers (3787c26)
  - [x] Create `web/src/data/changelog.ts` with structured release notes for existing versions (`v0.5.1`, `v0.5.0`, `v0.4.0`, etc.).
  - [x] Implement `web/src/utils/changelogStorage.ts` with cookie reading/writing (`agy_last_seen_version`), localStorage sync, and semver comparison.
- [x] Task: Phase Verification & Checkpoint (Refer to workflow.md) (3787c26)

## Phase 2: ChangelogModal Component (TDD)
- [~] Task: ChangelogModal Component Unit Tests
  - [ ] Write component tests in `web/src/components/ChangelogModal.test.tsx` testing modal rendering, sparkles header, latest version badge highlight, historical release list, Escape key dismiss, and "Got it" callback.
- [ ] Task: Implement ChangelogModal Component
  - [ ] Build `web/src/components/ChangelogModal.tsx` matching the screenshot design: dark theme card, sparkles icon, pill badges, bold feature bullet list, scrollable list, and "Got it" button.
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 3: Navbar, Footer, and App Integration (TDD)
- [ ] Task: Integration Unit Tests
  - [ ] Update `web/src/components/Navbar.test.tsx` and `web/src/components/Footer.test.tsx` to verify "What's new" trigger buttons.
  - [ ] Update `web/src/App.test.tsx` to verify auto-popup when new version is detected and dismissal behavior.
- [ ] Task: Integrate Changelog Triggers in Navbar and Footer
  - [ ] Add "What's new" sparkles ✨ button in `web/src/components/Navbar.tsx`.
  - [ ] Enable opening the changelog modal from `web/src/components/Footer.tsx`.
- [ ] Task: App-Level Orchestration & State Management
  - [ ] Wire modal visibility and auto-open check into `web/src/App.tsx`.
  - [ ] Ensure dismissing the modal updates the cookie and localStorage.
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

## Phase 4: Full Verification, Web Bundle Build & Docs
- [ ] Task: Run Full Automated Verification Suite
  - [ ] Run `npm --prefix web test`, `go test -race ./...`, `npm --prefix web run build`, and `golangci-lint run`.
- [ ] Task: Documentation & Changelog Entry
  - [ ] Update `README.md` documentation detailing the changelog popup feature.
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)
