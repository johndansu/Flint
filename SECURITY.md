# Security Policy
## Flint — Security Architecture, Practices & Vulnerability Disclosure

**Version:** 1.0
**Last updated:** May 2026
**Scope:** All versions of Flint (CLI, extension, daemon, shared libraries)

---

## Our Security Philosophy

Flint sits in a uniquely sensitive position. It watches code, reads files, monitors git history, captures errors, and over time builds a detailed picture of how a developer works. That level of access requires an unusually high standard of trust.

Our security philosophy is built on three commitments:

**Local first, always.** Every byte of developer data — code, errors, observations, human signals, session history — stays on the developer's machine. Nothing is sent to any server except the Anthropic API when a tool is explicitly invoked. No telemetry. No analytics. No exceptions.

**Minimal attack surface.** Flint is a local tool. It has no backend, no database, no user accounts, no authentication server. There is nothing to breach remotely because there is nothing remote. The attack surface is the local machine and the npm package itself.

**Transparent by default.** Developers can see everything Flint knows at any time with `flint memory`. They can delete everything with `flint memory clear all`. Nothing is hidden, nothing is obscured. If a developer doesn't trust what Flint has stored, they can inspect and delete it in seconds.

---

## Data That Never Leaves the Machine

The following data is stored locally in `~/.flint/` and **never transmitted** to any server under any circumstances:

- All daemon observations (technical and human)
- Error log and error dictionary
- Auto-fix history
- Session data and human intelligence signals
- Calibration data
- Codebase baselines from `flint scan`
- Git history analysis
- Tripwire configurations
- All SQLite database contents
- The human-readable `errors.log` file

**The only data that leaves the machine:**

| What | When | Where | Why |
|---|---|---|---|
| Code or error pasted as input | Developer explicitly runs a tool | Anthropic API | To generate the AI response |
| Daemon observation text | Daemon decides to surface it | Anthropic API | To generate follow-up response (if requested) |
| Dependency names + versions | CVE check | CVE database (read-only) | To check for known vulnerabilities |
| Binary download | `npm install` | GitHub Releases | To install the CLI binary |

Nothing else. Ever.

---

## Local Data Security

### File permissions

```
~/.flint/                    chmod 700  (owner read/write/execute only)
~/.flint/config.json         chmod 600  (owner read/write only)
~/.flint/flint.db            chmod 600  (owner read/write only)
~/.flint/errors.log          chmod 600  (owner read/write only)
~/.flint/daemon.sock         chmod 600  (runtime only, deleted on stop)
```

All files and directories are created with these permissions. If permissions are found to be more permissive on startup, Flint corrects them and logs a warning.

### API key storage

The Anthropic API key is stored in `~/.flint/config.json` with `chmod 600`. On macOS, Flint offers to store the API key in the system Keychain instead — this is recommended and is the default on macOS from v1.1 onward.

The API key is never:
- Logged to `errors.log` or any log file
- Included in `flint memory --export` output (redacted to `sk-ant-...****`)
- Transmitted anywhere except directly to `api.anthropic.com`
- Stored in environment variables by Flint itself

### IPC socket security

The Unix domain socket at `~/.flint/daemon-{workspace-hash}.sock` is created with `chmod 600`. Only processes running as the current user can connect to it. On Windows, named pipes use the current user's SID for equivalent isolation.

The socket is deleted when the daemon stops. If a stale socket file is found on startup (daemon was killed), it is removed and recreated.

### SQLite database

The SQLite database uses WAL (Write-Ahead Logging) mode for concurrent read safety. The database file is not encrypted at rest in v1 — full database encryption is planned for v2 using SQLCipher. Developers handling exceptionally sensitive code should note this and use `flint config --awareness spark` to minimise local storage.

---

## Supply Chain Security

### Binary distribution

Flint's Go binaries are cross-compiled and published to GitHub Releases on every tagged release. Each binary is:

- Built in a clean GitHub Actions environment from source
- Named with platform and architecture: `flint-darwin-arm64`, `flint-linux-amd64`, etc.
- SHA256 checksums published alongside each release
- Signed with a GitHub Actions OIDC token (Sigstore/cosign)

The npm wrapper (`wrapper/install.js`) downloads the binary over HTTPS and verifies the SHA256 checksum before marking it executable. Installation fails if the checksum does not match.

```javascript
// wrapper/install.js — checksum verification
const expectedHash = CHECKSUMS[platform];
const actualHash = crypto.createHash('sha256').update(binaryData).digest('hex');
if (actualHash !== expectedHash) {
    throw new Error(`Checksum mismatch. Expected ${expectedHash}, got ${actualHash}. Aborting.`);
}
```

### npm package integrity

The npm wrapper package is published with npm provenance (npm publish --provenance) linking the package to the GitHub Actions workflow that built it. This provides a verifiable chain from source code to published package.

Dependencies of the npm wrapper are intentionally minimal — only Node.js standard library. No third-party npm dependencies in the wrapper. No supply chain attack surface through transitive dependencies.

### Go dependencies

Go module dependencies are pinned in `go.sum` with cryptographic hashes. `go mod verify` is run in CI on every build. Dependabot is enabled for automated security updates.

### VS Code extension

The extension is published to the VS Code Marketplace, Cursor registry, and Windsurf registry using a verified publisher account. The extension makes no outbound network calls except:
- To `~/.flint/daemon.sock` (local only)
- To `api.anthropic.com` for follow-up responses (when developer taps "Tell me more")

The extension does not use `eval()`, does not load remote scripts, and does not access the file system directly — all file access goes through the daemon.

---

## Threat Model

### In scope

| Threat | Mitigation |
|---|---|
| Malicious npm package substitution | Binary checksum verification, npm provenance |
| Compromised GitHub Actions build | OIDC signing, reproducible builds |
| API key extracted from config | chmod 600, optional Keychain storage |
| IPC socket hijacked by another process | chmod 600, owner-only access |
| Auto-fix applies malicious change | Staged only, never committed, always notifies |
| Extension exfiltrates code | Extension makes no outbound calls except Anthropic API |
| Daemon sends data to Flint servers | There are no Flint servers |
| Footgun library poisoned via community contribution | Review process required for all contributions (v2+) |

### Out of scope (by design)

| Threat | Why out of scope |
|---|---|
| Full disk encryption | OS-level concern, not Flint's responsibility |
| Compromise of Anthropic API | Outside our control — use a dedicated API key |
| Keylogger on developer's machine | If the machine is compromised, Flint is the least of the problems |
| Social engineering of developer | Out of scope for any tool |

### Anthropic API usage

Flint sends code to the Anthropic API only when a tool is explicitly invoked or when a developer taps "Tell me more" on an observation. Developers should:

- Use a dedicated Anthropic API key for Flint (not shared with other applications)
- Review Anthropic's data usage policy at anthropic.com/privacy
- Use `flint awareness spark` for codebases with exceptionally sensitive IP — Spark mode never sends code automatically

---

## Auto-Fix Security

The auto-fix engine is the only part of Flint that modifies files. It operates under strict constraints:

1. **Staged only, never committed.** Auto-fix always calls `git add` and never `git commit`. The developer reviews and commits. This is enforced in code — there is no configuration option to enable auto-commit.

2. **Always notifies.** Every auto-fix surfaces a notification card in the sidebar or prints to the terminal. Nothing is ever truly silent. The developer sees every change Flint makes.

3. **Category-scoped.** Auto-fix only operates on the specific categories the developer has explicitly enabled: `cve`, `import`, or `format`. It never auto-fixes logic errors, type errors, or anything that requires understanding intent.

4. **Reversible.** Every auto-fix is logged in the `auto_fixes` table with the full diff. `git checkout -- <file>` reverts it. `flint fix --log` shows the complete history.

5. **Fails safe.** If auto-fix encounters an unexpected error, it aborts and logs the failure. It never leaves files in a partially modified state.

---

## Vulnerability Disclosure

### Reporting a vulnerability

If you discover a security vulnerability in Flint, please report it privately:

**Email:** security@flint.dev *(replace with actual address before launch)*

**PGP key:** *(publish PGP key before launch)*

**What to include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested mitigations

**What to expect:**
- Acknowledgement within 48 hours
- Status update within 7 days
- Fix timeline communicated within 14 days
- Credit in the security advisory (unless you prefer anonymity)

### Disclosure policy

We follow responsible disclosure:

- Reporters give us 90 days to fix before public disclosure
- We will not pursue legal action against good-faith security researchers
- We will credit researchers in security advisories unless they prefer anonymity
- Critical vulnerabilities (data exfiltration, supply chain) get a fix within 7 days

### What qualifies as a security vulnerability

**In scope:**
- Any data leaving the machine that shouldn't
- API key extraction or exposure
- IPC socket access by unauthorised processes
- Supply chain attacks on the npm package or binaries
- Auto-fix committing changes without developer consent
- Extension making unauthorised network calls

**Not in scope:**
- Vulnerabilities in the Anthropic API itself
- Vulnerabilities requiring physical access to the developer's machine
- Social engineering
- Issues in outdated versions (report against the latest release)

---

## Security Checklist for Contributors

Before submitting any PR that touches security-relevant code, verify:

- [ ] No new outbound network calls added without explicit documentation
- [ ] No data written to any location outside `~/.flint/`
- [ ] No API key, token, or secret logged or exported
- [ ] No new auto-fix category that commits automatically
- [ ] All new files created with correct permissions (600 for sensitive files)
- [ ] No new IPC message type that exposes file contents to the extension
- [ ] `go test ./internal/security/...` passes (security-specific test suite, v1.1+)

---

## Known Limitations (v1)

These are known limitations in v1 that will be addressed in future versions:

**SQLite database is not encrypted at rest.**
The `~/.flint/flint.db` file contains observations, session data, and human signals in plaintext. Full database encryption with SQLCipher is planned for v2. Mitigation: use full-disk encryption at the OS level (FileVault on macOS, LUKS on Linux).

**No integrity verification on `shared/footguns.json` at runtime.**
The footgun library loaded from disk is not verified against a checksum at runtime. A compromised footgun library could produce false or misleading observations. Mitigation: the footgun library is embedded in the binary at compile time in v1.1.

**Extension API key stored in VS Code extension settings.**
The Anthropic API key used for follow-up responses is stored in VS Code's extension settings (not the system Keychain). This is less secure than the CLI's Keychain storage. Planned for v1.1: use the system Keychain from the extension via native messaging.

**Clipboard monitoring is opt-in and off by default.**
When enabled for vibe coding detection, Flint monitors clipboard change frequency (not content). This is documented but some developers may not be aware it is available as an option.

---

## Security Audit Status

| Component | Status | Last reviewed |
|---|---|---|
| npm wrapper + binary distribution | Planned pre-launch | — |
| Go daemon + CLI | Planned pre-launch | — |
| VS Code extension | Planned pre-launch | — |
| IPC socket implementation | Planned pre-launch | — |
| Auto-fix engine | Planned pre-launch | — |
| SQLite storage | Planned pre-launch | — |

A third-party security audit is planned before the public free beta launch. Findings will be published.

---
