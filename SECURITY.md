# Security Policy

## Supported Versions

Security fixes are provided on a best-effort basis for the latest release on `main`.

| Version | Supported |
| :--- | :--- |
| Latest release | Yes |
| Older releases | No |

## Reporting a Vulnerability

Please report vulnerabilities privately.

Preferred method:

- Use GitHub Private Vulnerability Reporting in this repository:
  `Security` tab -> `Report a vulnerability`

Please do not open public issues for security-sensitive reports.

Include as much detail as possible:

- Affected version or commit
- Reproduction steps or proof of concept
- Impact assessment
- Suggested fix (if known)

## Response Process

After a valid report is received, maintainers will:

1. Acknowledge receipt.
2. Investigate and confirm impact.
3. Prepare and validate a fix.
4. Publish a release and disclose remediation details.

## Scope Notes

Security reports are especially relevant for:

- Installer and upgrade flows (`install.sh`, self-upgrade)
- Artifact integrity and checksum verification
- Lockfile and tool execution trust boundaries (`rig.lock`, `.rig/bin`)
