'use client';

import { useState } from 'react';

const ITEMS = [
  {
    q: 'Is Flint free?',
    a: 'Yes. Flint is MIT-licensed and fully open source. You pay only for the Anthropic API calls you make — Flint itself never charges anything.',
  },
  {
    q: 'How is this different from GitHub Copilot?',
    a: 'Copilot is a code completion assistant — you pull suggestions on demand. Flint is passive and observational. It watches your codebase and speaks up when it notices something worth saying, without you having to ask. It also tracks human signals (burnout, spirals) that Copilot has no awareness of.',
  },
  {
    q: 'What languages are supported?',
    a: 'Go, TypeScript/JavaScript, and Python are first-class. The daemon watches any file type, and CVE monitoring works for any language whose dependencies are in a supported manifest (go.mod, package.json, requirements.txt, pyproject.toml).',
  },
  {
    q: 'Does my code ever leave my machine?',
    a: 'Only when you explicitly run a tool (flint explain, flint diff, etc.) — and only the specific code you\'re analysing. The daemon never sends code automatically. Dependency names go to OSV.dev for CVE checks. Nothing else. Ever.',
  },
  {
    q: 'Do I need to keep a terminal open?',
    a: 'No. flint start detaches the daemon into the background. It keeps running after you close the terminal. flint stop kills it cleanly.',
  },
  {
    q: 'What AI model does Flint use?',
    a: 'Flint uses your own Anthropic API key and defaults to Claude Sonnet. You can switch to any Claude model via flint config model <id>. Bring your own key — Flint never manages or stores your key beyond your local config file.',
  },
  {
    q: 'Can I tune when Flint speaks up?',
    a: 'Yes. Every threshold is configurable via .flintrc — burnout hours, spiral edit count, lone-wolf days, and more. Dismissing an observation also lets you raise the threshold in one step, so Flint calibrates to you over time.',
  },
  {
    q: 'Does it work without the VS Code extension?',
    a: 'Fully. Flint is a CLI-first tool. The extension adds a sidebar panel and inline decorations, but every feature is accessible from the terminal.',
  },
];

export function FAQ() {
  const [open, setOpen] = useState<number | null>(null);

  return (
    <div className="divide-y divide-black/8 dark:divide-white/5">
      {ITEMS.map((item, i) => (
        <div key={i}>
          <button
            type="button"
            onClick={() => setOpen(open === i ? null : i)}
            className="w-full flex items-center justify-between text-left py-5 gap-4 group"
          >
            <span className="text-[15px] text-neutral-800 dark:text-neutral-200 group-hover:text-neutral-900 dark:group-hover:text-white transition-colors">
              {item.q}
            </span>
            <span
              className="text-neutral-500 shrink-0 transition-transform duration-200"
              style={{ transform: open === i ? 'rotate(45deg)' : 'rotate(0deg)' }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
                <path d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
            </span>
          </button>
          <div
            style={{
              maxHeight: open === i ? '300px' : '0',
              opacity: open === i ? 1 : 0,
              overflow: 'hidden',
              transition: 'max-height 0.3s ease, opacity 0.25s ease',
            }}
          >
            <p className="text-sm text-neutral-500 leading-relaxed pb-5 max-w-2xl">{item.a}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
