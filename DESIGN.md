# Design System
## Flint — Extension Visual Design Language

**Version:** 1.0
**Last updated:** May 2026
**Scope:** VS Code, Cursor, Windsurf extension visual design

---

## Design Philosophy

> Calm and trusted. Theme-adaptive. Unmistakably Flint.

Flint's design must feel like a senior developer sitting nearby — present, unhurried, occasionally leaning over to say something worth hearing. It should never feel like a dashboard, a notification system, or a chatbot.

Three principles govern every design decision:

**Presence without weight.** Flint exists in the editor at all times but takes up no visual space when silent. The quiet state is just code. Nothing competes with the developer's focus.

**Clarity through colour.** Three colours, three meanings. Purple for technical observations. Green for wins. Amber for human observations. The colour tells the developer what kind of thing Flint is saying before they read a word.

**Anchored to the code.** Observations are tied to the code they're about. The glow on the relevant line connects what Flint sees to what the developer sees. The observation is not floating in a separate panel — it is adjacent to the thing it is talking about.

---

## Colour System

Flint uses three semantic colours. All other surfaces adapt to the developer's VS Code theme.

### Purple — technical observations
Flint's primary voice. Used when Flint is speaking about code quality, security, performance, patterns.

```
Overlay border:    rgba(180, 150, 255, 0.18)
Line glow:         rgba(180, 150, 255, 0.13) → transparent
Line accent:       rgba(180, 150, 255, 0.65)   (2px left edge)
Line number:       rgba(180, 150, 255, 0.60)
Indicator dot:     rgba(180, 150, 255, 0.70)
Tag text:          rgba(180, 150, 255, 0.45)
Status dot:        rgba(180, 150, 255, 0.60)
Status text:       rgba(180, 150, 255, 0.65)
Send button:       rgba(180, 150, 255, 0.09) bg / rgba(180, 150, 255, 0.35) border
```

### Green — win acknowledgments
Used when Flint is acknowledging something the developer did well.

```
Overlay border:    rgba(100, 220, 130, 0.18)
Line glow:         rgba(100, 220, 130, 0.11) → transparent
Line accent:       rgba(100, 220, 130, 0.55)
Line number:       rgba(100, 220, 130, 0.55)
Indicator dot:     rgba(100, 220, 130, 0.65)
Tag text:          rgba(100, 220, 130, 0.38)
Status dot:        rgba(100, 220, 130, 0.60)
```

### Amber — human observations
Used for observations about the developer as a person, not their code. Warmer, softer.

```
Overlay border:    rgba(255, 200, 80, 0.15)
Line glow:         none (human observations have no code anchor)
Tag text:          rgba(255, 200, 80, 0.38)
Indicator dot:     rgba(255, 200, 80, 0.55)
Status dot:        rgba(255, 200, 80, 0.50)
```

### Theme adaptation
All surface colours — overlay background, editor chrome, gutter — adapt to the developer's VS Code theme. Flint's three semantic colours are the only constants. A developer on a light theme sees the same purple/green/amber but everything else matches their environment.

---

## The Overlay

The overlay is not a panel. It is not always there. It appears only when Flint has something to say, slides in from the right edge, and exists alongside the code it is talking about.

### Dimensions
```
Width:             268px
Position:          absolute, right: 0, top: 0, bottom: 0
Background:        rgba(16, 17, 28, 0.96) with backdrop-filter: blur(16px)
Border left:       1px solid [semantic colour at 0.18 opacity]
z-index:           10
```

### Entry animation
Slides in from the right edge. Smooth, not bouncy.
```css
animation: slideIn 0.22s cubic-bezier(0.16, 1, 0.3, 1);

@keyframes slideIn {
  from { transform: translateX(100%); opacity: 0; }
  to   { transform: translateX(0);    opacity: 1; }
}
```

### Content structure
```
┌─────────────────────────────────────┐
│ flint                               │  ← tag (9px, semantic colour, 0.45 opacity)
│                                     │
│ Observation text                    │  ← 11.5px, rgba(205,215,245,0.9), lh 1.65
│ in plain senior dev language.       │
│                                     │
│ ┌ agent block ─────────────────── ┐ │  ← only on technical observations
│ │ In file.ts line N, do X.       │ │  ← 9.5px mono, rgba(122,162,247,0.72)
│ │ Don't change anything else.    │ │
│ └───────────────────────────────── ┘ │
│                                     │
│ [Copy] [Send →]   tell me more  ✕   │  ← actions (technical only)
└─────────────────────────────────────┘
```

### Agent block
The agent prompt sits in its own block, visually separate from the observation text.
```
Background:   rgba(122, 162, 247, 0.06)
Border left:  1.5px solid rgba(122, 162, 247, 0.25)
Border radius: 5px
Padding:      8px 10px
Font:         monospace, 9.5px
Color:        rgba(122, 162, 247, 0.72)
```

### Action buttons
```
Copy button:
  border: 0.5px solid rgba(255,255,255,0.12)
  color:  rgba(200,210,240,0.6)
  bg:     transparent

Send button:
  border: 0.5px solid rgba(180,150,255,0.35)
  color:  rgba(180,150,255,0.9)
  bg:     rgba(180,150,255,0.09)

Tell me more:
  no border, no bg
  color: rgba(122,162,247,0.5)
  text-decoration: underline
  margin-left: auto

Dismiss ✕:
  no border, no bg
  color: rgba(255,255,255,0.18)
  padding: 4px 6px
```

---

## The Collapsed Indicator

After 30 seconds, the full overlay shrinks to a slim indicator on the right edge. The observation is not gone — it is waiting. The glow on the relevant line remains.

### Dimensions
```
Width:          28px
Position:       absolute, right: 0, top: 0, bottom: 0
Background:     rgba(16, 17, 28, 0.85)
Border left:    1px solid [semantic colour at 0.15 opacity]
```

### Contents
A dot and a short vertical line, centered vertically.
```
Dot:   7px circle, semantic colour at 0.70 opacity
Line:  1.5px × 18px, semantic colour at 0.20 opacity
Gap:   6px between dot and line
```

### Behaviour
- Developer clicks indicator → overlay re-expands with full slide-in animation
- Observation is never lost — clicking always brings it back
- Multiple observations collapse to multiple dots stacked vertically

### Collapse timing
```
Technical observations:    collapse after 30 seconds
Win acknowledgments:       collapse after 30 seconds
Human observations:        collapse after 30 seconds
Pre-commit checkpoint:     never auto-collapses (too important)
CVE auto-fix notification: never auto-collapses
```

---

## The Line Glow

The glow connects the observation to the code it is about.

### Technical and win observations — glow present
```
Left accent:
  position: absolute, left: -12px, top: 0, bottom: 0
  width: 2px
  background: [semantic colour at 0.65 opacity]
  border-radius: 0 1px 1px 0

Background wash:
  position: absolute, left: -12px, right: -200px, top: 0, bottom: 0
  background: linear-gradient(90deg,
    [semantic colour at 0.13] 0%,
    [semantic colour at 0.05] 50%,
    transparent 100%)
  pointer-events: none

Line number highlight:
  color: [semantic colour at 0.60 opacity]
```

### Human observations — no glow
Human observations are not about a specific line of code. No line is highlighted. The overlay appears without a code anchor. This visual difference communicates the distinction.

### Multiple glow lines
When an observation covers multiple lines (win acknowledgment on a refactored block), all relevant lines receive the glow. Same colour, same treatment.

---

## The Quiet State

When Flint has nothing to say, there is nothing to see. No panel, no strip, no indicator. Just the editor.

The only signal that Flint is active is the status bar item:
```
● Flint 1.0   flame · web dev   watching
```

Status dot opacity drops to 0.35 in quiet state. Status text drops to 0.50. This communicates "present but silent" without any visual weight.

---

## Status Bar

Always visible. Shows everything the developer needs to know about Flint's state.

### Format
```
[dot] Flint 1.0   [awareness] · [role]   [session state or observation count]
```

### Examples
```
● Flint 1.0   flame · web dev   deep work
● Flint 1.0   flame · web dev   1 observation
● Flint 1.0   forge · devops    debugging spiral
○ Flint (standalone)
○ Flint (reconnecting...)
⚠ Flint (version mismatch)
```

### Dot colours
```
Technical observation active:  rgba(180, 150, 255, 0.60)  — purple
Win observation active:        rgba(100, 220, 130, 0.60)  — green
Human observation active:      rgba(255, 200, 80,  0.50)  — amber
Quiet / watching:              rgba(180, 150, 255, 0.35)  — muted purple
Standalone / reconnecting:     rgba(255, 255, 255, 0.20)  — grey
```

---

## Observation Types — Design Summary

| Type | Glow | Colour | Actions | Collapses |
|---|---|---|---|---|
| Technical | On relevant line | Purple | Copy, Send, Tell me more, ✕ | 30s |
| Vibe coding | On relevant lines | Purple | Copy, Send, Tell me more, ✕ | 30s |
| Intent question | No specific line | Purple | Tell me more | 30s |
| Win | On refactored lines | Green | None | 30s |
| Human | None | Amber | None | 30s |
| Taste | On relevant line | Purple (muted) | Tell me more, ✕ | 30s |
| Pre-commit | No glow (terminal) | Purple | y/n in terminal | Never |
| Auto-fix | None | Green | View diff in terminal | Never |

---

## Typography

All text inside the overlay uses system font stack, not the editor's monospace font — except the agent prompt block which uses monospace to signal "this is a command."

```
Tag:              9px, sans-serif, letter-spacing 0.12em
Observation text: 11.5px, sans-serif, line-height 1.65
Agent prompt:     9.5px, monospace, line-height 1.55
Action buttons:   10px, sans-serif
Dismiss / more:   10px, sans-serif
Status bar:       10px, sans-serif
```

---

## Sidebar Modes

The sidebar has three distinct modes. They never overlap.

### Passive mode (default)
The overlay slides in when the daemon has an observation. No input. No persistent panel. The developer never opens it — Flint speaks when it has something to say. The quiet state shows "watching silently" and a subtle "ask flint →" button at the bottom.

### REPL mode (intentional)
Opened by the developer via `Flint: Open REPL` command or clicking "ask flint →" in the quiet state. The sidebar becomes a focused conversation panel:

```
┌──────────────────────────────┐
│ flint repl                   │  ← tag
│ api-service · 47 files · forge│  ← context loaded
├──────────────────────────────┤
│ › why does auth differ...    │  ← developer question
│                              │
│   flint                      │
│   Login uses legacy session  │  ← Flint answer with
│   from 8 months ago...       │    specific files
│                              │
│ › which files to touch?      │
│                              │
│   flint                      │
│   4 files: session.js...     │
│                              │
├──────────────────────────────┤
│ ›  [cursor]                  │  ← input always at bottom
├──────────────────────────────┤
│ clears on close        exit  │  ← always visible
└──────────────────────────────┘
```

**REPL mode suppresses passive observations.** While the developer is in a REPL session, the daemon does not surface overlay cards. Flint does not interrupt a conversation you intentionally started.

**Status bar in REPL mode:** amber dot, `Flint repl` — visually distinct from passive mode.

**Quiet state button:** When Flint has nothing to say and passive mode is quiet, a subtle "ask flint →" button appears at the bottom of the sidebar. Low profile — available without being intrusive.

### Manual tools mode (Spark / standalone)
When no daemon is running (standalone mode) or awareness is set to Spark, the sidebar shows the manual tools panel instead of the passive overlay. Tools listed for the current role. Developer pastes input, output streams inline.

---

## REPL Mode — Design Details

### Context header
Shows: project name, files indexed, awareness level. Developer always knows what Flint has loaded.

```
flint repl
api-service · 47 files · forge
```

### Message styling
```
Developer questions:
  › arrow + plain text, rgba(215,220,240,0.85)

Flint responses:
  Small "flint" tag above (9px, purple, 0.35 opacity)
  Response text: 10.5px, rgba(185,200,230,0.75), lh 1.55
  File references: rgba(166,227,161,0.85) — green, distinct
  Highlighted terms: rgba(137,180,250,0.8) — blue
```

### Input area
```
Background:   rgba(255,255,255,0.04)
Border:       0.5px solid rgba(180,150,255,0.20)
Border radius: 5px
Arrow prefix: › in purple at 0.5 opacity
Cursor:       5×11px, rgba(180,150,255,0.60), blinking
```

### Footer
Always visible. Never hidden. Two elements:
- Left: "clears on close" — 9px, rgba(255,255,255,0.15)
- Right: "exit" — 9px, rgba(249,226,175,0.40), amber

The footer is the developer's constant reminder of the session's boundaries.

### REPL mode does not have:
- Scroll history between sessions
- A "save conversation" option
- File attachment
- Code execution
- Any persistence whatsoever

---

## Manual Tools Panel

The manual tools panel appears in standalone mode or when awareness is Spark. It is not the default state — the overlay is.

### Design
- Lives in the VS Code sidebar
- Lists all tools for the current role with a small icon per tool
- Each tool has a colour consistent with its category:
  - Review: purple
  - Debug: blue
  - Explain/Doc: green
  - Audit/Security: amber
- Selected tool opens an input area below the list
- Output streams into the input area, replacing the placeholder text
- No send button — output streams automatically when input is pasted

### Role selector
A small pill at the top showing the current role. Clicking opens an inline dropdown.

---

## Awareness Picker

Lives in the sidebar, accessible via `Flint: Change Awareness` command or status bar click.

Three rows, each with:
- A small icon (Spark: sparkles, Flame: flame, Forge: hammer)
- Name and one-line description
- Active badge on the current level

Colours:
- Spark: blue tint
- Flame: purple tint (Flint's primary)
- Forge: amber tint

---

## What This Design Is Not

**Not a panel that is always open.** The overlay exists only when Flint has something to say. The rest of the time there is nothing except the quiet state.

**Not a notification popup.** Notifications interrupt. The overlay slides in alongside the code without taking focus or interrupting flow.

**Not a chatbot.** The overlay has no input field. One follow-up per observation, then closes. The REPL is intentional and bounded — not a general conversation partner that runs indefinitely.

**Not themed.** The overlay does not try to match the developer's colour scheme. It has its own character — dark, focused, minimal — consistent across all themes. Only the surfaces adapt.

**Not persistent.** REPL sessions clear on close. Overlay observations collapse after 30 seconds. Nothing accumulates in the UI between sessions.

---

## Implementation Notes for VS Code Webview

### CSP header (required)
```typescript
const nonce = generateNonce();
view.webview.html = `
  <!DOCTYPE html>
  <html>
  <head>
    <meta http-equiv="Content-Security-Policy"
      content="default-src 'none';
               style-src 'nonce-${nonce}';
               script-src 'nonce-${nonce}';
               connect-src https://api.anthropic.com;">
  </head>
  ...
```

### VS Code theme variables to use
```css
/* Use these instead of hardcoding — they adapt to the developer's theme */
--vscode-editor-background
--vscode-editor-foreground
--vscode-editorLineNumber-foreground
--vscode-editor-font-family
--vscode-editor-font-size
```

### Overlay positioning
The overlay must be positioned inside the webview, not as a VS Code decoration. VS Code does not support arbitrary overlays on the editor surface — the sidebar webview renders alongside the editor, not on top of it.

The visual design shown uses the sidebar as the "right edge" surface. The glow on the relevant line is implemented as a VS Code `TextEditorDecorationType` — a standard decoration API that works in all VS Code-based editors.

### Line glow implementation
```typescript
// extension/src/editor/decorations.ts
const technicalDecoration = vscode.window.createTextEditorDecorationType({
    backgroundColor: 'rgba(180, 150, 255, 0.08)',
    borderColor: 'rgba(180, 150, 255, 0.65)',
    borderWidth: '0 0 0 2px',
    borderStyle: 'solid',
    isWholeLine: true,
});

const winDecoration = vscode.window.createTextEditorDecorationType({
    backgroundColor: 'rgba(100, 220, 130, 0.07)',
    borderColor: 'rgba(100, 220, 130, 0.55)',
    borderWidth: '0 0 0 2px',
    borderStyle: 'solid',
    isWholeLine: true,
});

const humanDecoration = vscode.window.createTextEditorDecorationType({
    backgroundColor: 'rgba(255, 200, 80, 0.05)',
    borderWidth: '0',
    isWholeLine: true,
});

// Apply decoration to the relevant line
function applyGlow(
    editor: vscode.TextEditor,
    lineNumber: number,
    type: 'technical' | 'win' | 'human'
) {
    const line = editor.document.lineAt(lineNumber - 1);
    const decoration = type === 'win' ? winDecoration
                     : type === 'human' ? humanDecoration
                     : technicalDecoration;
    editor.setDecorations(decoration, [line.range]);
}

// Clear all glows
function clearGlows(editor: vscode.TextEditor) {
    editor.setDecorations(technicalDecoration, []);
    editor.setDecorations(winDecoration, []);
    editor.setDecorations(humanDecoration, []);
}
```

### Collapse timer
```typescript
// In sidebar/provider.ts
private collapseTimer?: NodeJS.Timeout;

showObservation(obs: Observation) {
    this.view?.webview.postMessage({ type: 'show', obs });
    clearTimeout(this.collapseTimer);

    // Never auto-collapse pre-commit or auto-fix
    if (obs.category === 'precommit' || obs.category === 'auto_fix') return;

    // Collapse to indicator after 30 seconds
    this.collapseTimer = setTimeout(() => {
        this.view?.webview.postMessage({ type: 'collapse' });
    }, 30_000);
}
```

---

## File Structure

```
extension/
├── src/
│   ├── editor/
│   │   ├── decorations.ts   # line glow via TextEditorDecorationType
│   │   └── codelens.ts      # CodeLens above annotated lines
│   ├── sidebar/
│   │   ├── provider.ts      # WebviewViewProvider, collapse timer
│   │   ├── overlay.html     # overlay HTML template
│   │   ├── overlay.css      # overlay styles (all colours defined here)
│   │   └── tools-panel.html # manual tools panel HTML
│   └── ...
└── media/
    ├── icon.svg             # Flint icon for sidebar and marketplace
    └── icon-dark.svg        # Dark mode variant
```

---

## Design Changelog

| Version | Change |
|---|---|
| 1.0 | Initial design — ambient overlay with line glow, three-colour system, collapse to indicator after 30s |

---
