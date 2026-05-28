# Tech Stack
## Flint — Technology Decisions & Rationale

**Version:** 1.0
**Last updated:** May 2026

---

## The Core Decision — Go + TypeScript in a Monorepo

Flint uses two languages with a shared contract, not one language trying to do everything.

**Go** for the runtime — daemon, CLI, all background processing.
**TypeScript** for the interface — VS Code/Cursor/Windsurf extension, npm wrapper shim.

They never touch each other's internals. They communicate through a shared JSON schema that generates types for both sides. One monorepo. One pipeline. Two runtimes that feel like one.

---

## Why Go for the Runtime

| Concern | Why Go wins |
|---|---|
| Daemon memory | Compiles to ~5MB binary. Idle RAM under 30MB. Node would sit at 60–100MB minimum. |
| Daemon CPU | No garbage collection pauses. Goroutines for concurrent watching. Consistently under 1% idle. |
| CLI startup | Single binary. Sub-100ms startup. No Node runtime to initialise. |
| Distribution | Cross-compiles to every platform in one command. No dependency hell. |
| Long-running processes | Go is purpose-built for this. Node is not. |
| File watching | fsnotify is the gold standard. Native OS events. No polling. |

---

## Why TypeScript for the Extension

No choice. The VS Code extension API is TypeScript only. This extends to Cursor and Windsurf since both are built on VS Code. One TypeScript extension ships to three marketplaces.

---

## Why a Monorepo

- One repo, one PR, one CI pipeline
- Schema changes break both sides at compile time, not at runtime
- Shared `tools.json` — one source of truth for all 24 tool prompts
- Release is one command that builds Go + compiles TS + publishes npm + publishes extension
- Contributors see the whole system in one place

---

## Full Stack

| Layer | Technology | Version | Reason |
|---|---|---|---|
| CLI | Go | 1.22+ | Fast, single binary, cross-compile |
| Daemon | Go | 1.22+ | Low memory, goroutines, fsnotify |
| File watching | fsnotify | latest | OS-native events, no polling |
| Git monitoring | go-git | v5 | Pure Go, no git binary dependency |
| CLI framework | cobra | v1.8 | Industry standard Go CLI |
| SQLite | modernc.org/sqlite | latest | Pure Go, no CGO, no system dependency |
| Unix socket IPC | net (stdlib) | — | No external dependency |
| HTTP (bridge, v2) | net/http (stdlib) | — | REST bridge for web dashboard |
| VS Code extension | TypeScript | 5.x | Only option for VS Code API |
| Extension bundler | esbuild | latest | Fast, reliable |
| npm wrapper | Node.js | 20+ | Thin shim only |
| Anthropic SDK | Go HTTP client | — | Direct API calls, full streaming control |
| Schema codegen | Node.js script | 20+ | Generates Go + TS types from JSON schema |
| Build orchestration | Makefile + Turbo | — | Simple, fast, parallel |
| CI/CD | GitHub Actions | — | Native GitHub, free for open source |
| Release | GoReleaser | latest | Cross-compile + GitHub Release + checksums |

---

## v2 Additions

| Layer | Technology | Reason |
|---|---|---|
| JetBrains plugin | Kotlin | Only language for IntelliJ platform |
| Neovim plugin | Lua | Standard for Neovim plugins |
| Web dashboard | Next.js + TypeScript | Fast to build, same TS ecosystem |
| Slack integration | Slack Web API | Official SDK |
| Email | SMTP / SendGrid | Simple, reliable |
| Telegram | Telegram Bot API | Direct HTTP |

---

## v3 Additions

| Layer | Technology | Reason |
|---|---|---|
| AST parsing | tree-sitter (Go bindings) | Multi-language, fast, incremental |
| Codebase index | Custom SQLite FTS5 | Local, no external vector DB needed for v3 |
| WhatsApp | WhatsApp Business API | Requires Meta approval — start early |

---

## v4 Additions

| Layer | Technology | Reason |
|---|---|---|
| Team sync | libp2p or custom | Encrypted P2P, no central server |
| License validation | Custom + Ed25519 | Offline-capable, tamper-evident |

---

## What We Deliberately Avoided

| Thing | Why not |
|---|---|
| Electron | Too heavy for a daemon tool. Not needed. |
| Docker | Developer tool — containers add friction |
| GraphQL | Overkill for local REST bridge |
| MongoDB / Postgres | SQLite is the right database for local-first |
| Central server (v1-v3) | Privacy principle. Local first always. |
| Python | No strong reason over Go for this use case |
| Rust | Better performance than Go but 3x the complexity. Not justified for v1. |
| React Native | Not a mobile app |
| Electron (for web dashboard v2) | Next.js served locally is simpler and lighter |

---

## The npm Wrapper Pattern

Flint's Go binary is distributed through npm — not because it's a Node app, but because npm is where developers look for CLI tools.

Pattern: `npm install -g flint` → post-install script detects OS + arch → downloads the right Go binary from GitHub Releases → thin JS shim in PATH calls the binary.

This is the same pattern used by: esbuild, Turbo, Biome, SWC, and others. Battle-tested. Zero Node runtime overhead at execution time.

---

## Local Storage — Why SQLite

- Single file. Easy to backup, inspect, export.
- Fast enough for everything Flint needs (< 10ms for all queries).
- `modernc.org/sqlite` is pure Go — no CGO, no system library dependency, compiles cleanly on all platforms.
- WAL mode for concurrent reads (daemon writing + CLI reading simultaneously).
- Full SQL — complex queries for human intelligence analytics without an ORM.

---

## The Shared Schema — Why It Matters

`shared/schema.json` is not documentation. It is the build system.

Every time it changes, `node scripts/generate-schema.js` regenerates:
- `core/pkg/schema/schema.go` — Go structs with JSON tags
- `extension/src/ipc/messages.ts` — TypeScript interfaces

If a message type is renamed in the schema, both sides break at compile time. This is a feature. It means Go and TypeScript can never silently diverge.

---
