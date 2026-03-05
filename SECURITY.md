# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately:

**Email:** security@propifly.com

**Do not** open a public GitHub issue for security vulnerabilities.

## What to Include

- Description of the vulnerability
- Steps to reproduce
- Impact assessment (what an attacker could do)
- Suggested fix (if you have one)

## Response Timeline

- **48 hours**: acknowledgment of your report
- **7 days**: initial assessment and severity classification
- **30 days**: fix released (for critical/high severity)

## Scope

This policy covers:
- taskprim, stateprim, knowledgeprim, and queueprim binaries
- The primkit shared library
- Configuration file handling (credential interpolation)
- HTTP API authentication
- SQLite database access

Out of scope:
- Cloudflare R2 / AWS S3 infrastructure security
- Third-party dependencies (report upstream)
