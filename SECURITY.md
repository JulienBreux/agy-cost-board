# Security Policy

The `agy-cost-board` team takes security issues seriously and welcomes responsible vulnerability reporting. This document outlines our security policies, supported versions, and the process for reporting vulnerabilities.

---

## Supported Versions

Security updates and patches are provided for the following versions:

| Version | Supported          | Notes                              |
| ------- | ------------------ | ---------------------------------- |
| 1.x     | :white_check_mark: | Current active release branch      |
| < 1.0   | :x:                | Development / experimental release |

We strongly recommend always running the latest released version of `agy-cost-board` to ensure you have the latest security patches and stability updates.

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities via public GitHub issues, discussions, or pull requests.**

If you discover a security vulnerability in `agy-cost-board`, please report it privately:

1. **GitHub Security Advisory (Preferred)**:
   - Navigate to the [Security Advisories](https://github.com/JulienBreux/agy-cost-board/security/advisories) tab of the repository.
   - Click **"Report a vulnerability"** to submit a private draft advisory.

2. **Email**:
   - If you cannot use GitHub Security Advisories, send an email to **julien.breux@gmail.com** with the subject line `[SECURITY] agy-cost-board Vulnerability Report`.

### What to Include in Your Report

To help us triage and resolve the issue quickly, please provide:
- A clear description of the vulnerability and its potential impact.
- Affected component(s) (e.g., HTTP REST API, embedded Web UI, BigQuery query builder, CLI parser, configuration loader).
- Step-by-step instructions to reproduce the issue, including proof-of-concept commands, payloads, or sample log fixtures.
- The version or Git commit of `agy-cost-board` you tested against.
- Execution mode used (CLI, TUI, local `serve`, or Cloud Run container).
- Any proposed mitigations or fixes.

---

## Response & Disclosure Process

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within **48 hours**.
2. **Assessment & Triage**: We will verify the vulnerability, determine its severity (CVSS), and keep you informed of our progress.
3. **Fix & Verification**: We will develop and test a patch privately. We may ask you to verify the fix before release.
4. **Coordinated Disclosure**: A security advisory and a patched release will be published simultaneously. We will credit you in the advisory release notes (unless you request anonymity).
5. **Timeline**: We aim to resolve critical vulnerabilities within **14 days** and moderate vulnerabilities within **30 days**.

---

## Security Best Practices for `agy-cost-board` Deployments

When running `agy-cost-board` in production, we recommend the following security measures:

1. **Protect Sensitive GCP Credentials & Configuration**:
   - Never commit `.agy-cost-board.yaml`, service account keys, or environment files containing GCP project metadata to version control.
   - Use Google Cloud Application Default Credentials (ADC) or Workload Identity Federation when running on Google Cloud Run instead of long-lived service account key files.
2. **Apply Principle of Least Privilege**:
   - The identity running `agy-cost-board` only requires read permissions for attribution:
     - `roles/bigquery.jobUser` on the project.
     - `roles/bigquery.dataViewer` on the telemetry dataset and billing export dataset.
   - Only grant `roles/logging.configWriter` or `roles/bigquery.dataEditor` during initial automated resource provisioning (`setup --create`).
3. **Cloud Run & Network Boundary**:
   - By default, `agy-cost-board serve` delivers a web dashboard without built-in authentication.
   - When deploying to Google Cloud Run, protect the service using **Google Cloud Identity-Aware Proxy (IAP)**, Cloud Run IAM authentication (`--no-allow-unauthenticated`), or place it behind an authorized reverse proxy (e.g., Google Cloud Armor / Load Balancer) to prevent unauthorized access to internal billing and developer token metrics.
