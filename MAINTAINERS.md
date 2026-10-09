# Project Maintainers

This document lists the maintainers of **AGY Cost Board (`agy-cost-board`)** and describes project governance, roles, and responsibilities.

---

## Active Maintainers

| Name | GitHub | Role | Focus Areas |
| :--- | :--- | :--- | :--- |
| **Julien Breux** | [@julienbreux](https://github.com/julienbreux) | Founder & Lead Maintainer | Core Architecture, BigQuery Telemetry, Cost Attribution, Cloud Run & Release Management |

---

## Roles & Responsibilities

Maintainers are responsible for the health, stability, security, and future of the `agy-cost-board` project. Core responsibilities include:

1. **Code Review & Quality Enforcement**:
   - Review pull requests against project engineering principles: idiomatic Go, single-binary architecture with embedded React 19 SPA, mathematical accuracy of proportional cost attribution, thread-safe TTL caching, and zero external runtime dependencies.
   - Ensure the test suite (`make test`), race detector, and static analysis (`make lint`) remain 100% green.
2. **Issue Triage**:
   - Triage reported bugs, BigQuery schema compatibility questions, Cloud Logging sink edge cases, and feature requests.
   - Tag issues and guide contributors toward relevant specifications and architectural guidelines.
3. **Release Management**:
   - Maintain the changelog, cut semantic version tags (`v*`), and oversee automated multi-architecture releases (`.goreleaser.yaml`).
4. **Security Response**:
   - Triage and coordinate security reports in accordance with [`SECURITY.md`](SECURITY.md).
5. **Community & Culture**:
   - Foster an inclusive, constructive, and respectful environment in compliance with the [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

---

## Decision-Making Process

`agy-cost-board` follows a **consensus-seeking model** grounded in clear specifications:

- **Minor Changes & Bug Fixes**:
  - Can be approved and merged by any maintainer once CI passes and acceptance criteria are satisfied.
- **Architectural & Public API Changes**:
  - Major features, new attribution models, changes to CLI commands, or modifications to REST/API contracts require an open issue or specification before implementation.
  - Discussion takes place openly on GitHub issues or pull requests under a **lazy consensus** principle (no sustained objections within 72 hours).

---

## Becoming a Maintainer

We welcome contributors who demonstrate sustained commitment to the project. Criteria for nomination as a maintainer include:

- A track record of high-quality contributions (pull requests, code reviews, issue triaging, or BigQuery/FinOps validation).
- Strong understanding of the codebase, idiomatic Go standards, and Google Cloud telemetry pipelines (Cloud Logging, BigQuery, Cloud Billing).
- Exemplary adherence to our [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

Nomination is proposed by an existing maintainer and confirmed through consensus.
