# Contributing to Flint

**Version:** 1.0
**Last updated:** May 2026

---

## Before You Start

Read these three docs first. They are not optional.

1. `ARCHITECTURE.md` — understand the full system
2. `TECH_STACK.md` — understand why things are built the way they are
3. `shared/schema.json` — understand the contract between Go and TypeScript

If you change the schema, you change both sides. If you skip reading these, your PR will be sent back.

---

## Monorepo Setup

```bash
# Prerequisites
# Go 1.22+
# Node.js 20+
# make

git clone https://github.com/flint-dev/flint
cd flint

# Install Node dependencies
npm install
cd extension && npm install && cd ..
cd wrapper && npm install && cd ..

# Generate schema types (do this first, always)
node scripts/generate-schema.js

# Build everything
make build

# Run tests
make test

# Start in dev mode
make dev
```

---

## Project Structure

```
core/       Go — touch this for daemon, CLI, AI, storage, IPC
extension/  TypeScript — touch this for VS Code/Cursor/Windsurf UI
wrapper/    TypeScript — touch this for npm distribution only
shared/     JSON — touch this for IPC schema and tool prompts
scripts/    Node.js — touch this for build tooling only
```

---

## The Golden Rules

**1. Never edit generated files.**
`core/pkg/schema/schema.go` and `extension/src/ipc/messages.ts` are generated. Edit `shared/schema.json` and run `node scripts/generate-schema.js`.

**2. Schema changes require both sides.**
If you add a field to a message in `schema.json`, you must handle it in both the Go daemon and the TypeScript extension in the same PR.

**3. The daemon must stay lean.**
If your change increases idle CPU above 1% or idle RAM above 30MB, it will not merge. Profile before submitting.

**4. Human intelligence observations require the highest bar.**
Any new human observation type requires: 0.85 confidence threshold, 7-day cooldown, measured language, and a review from at least two maintainers.

**5. Tool prompts live in `shared/tools.json`.**
Never hardcode a system prompt in Go or TypeScript. All 24 tool prompts (and future ones) live in the shared tools file.

---

## Adding a New Tool

1. Add the tool definition to `shared/tools.json`:
```json
{
  "id": "your-tool",
  "role": "webdev",
  "category": "review",
  "name": "Your Tool Name",
  "description": "One sentence description",
  "placeholder": "Paste your input here...",
  "hint": "Context that helps the user give good input",
  "prompt": "You are a senior [role] engineer. [Specific instructions]. Senior dev tone. One sentence why-this-matters at the end."
}
```

2. Run `make build` — the tool is automatically available in CLI and extension.

3. Add a test in `core/internal/tools/tools_test.go`.

---

## Adding a New Daemon Observation

1. Add the category to `shared/schema.json` enum list.
2. Implement the detection logic in `core/internal/daemon/technical.go` or `human.go`.
3. Set the confidence threshold in `core/internal/daemon/calibration.go`.
4. Add time-gate category in `core/internal/daemon/timegate.go`.
5. Write a test that simulates the trigger condition in `core/internal/daemon/technical_test.go`.
6. Add an example card to `extension/src/sidebar/card.ts`.

---

## Running Individual Tests

```bash
# Go tests
cd core && go test ./...
cd core && go test ./internal/daemon/... -v
cd core && go test ./internal/daemon/... -run TestTechnicalEngine

# TypeScript tests
cd extension && npm test
cd extension && npm test -- --grep "sidebar"
```

---

## Submitting a PR

- One concern per PR. If you're fixing a bug and adding a feature, two PRs.
- All tests must pass.
- `go vet ./...` must pass.
- `node scripts/generate-schema.js && git diff --exit-code` must pass (schema in sync).
- New daemon observation types require a description of: what triggers it, what the confidence threshold is, what the cooldown is, and three example outputs in the senior dev voice.

---

## Releasing

Releases are triggered by pushing a version tag:

```bash
git tag v1.2.3
git push origin v1.2.3
```

The release pipeline:
1. Runs all tests
2. Cross-compiles Go binaries for all platforms
3. Uploads binaries to GitHub Release
4. Publishes npm wrapper
5. Publishes VS Code extension to Marketplace
6. Publishes to Cursor and Windsurf extension registries

---
