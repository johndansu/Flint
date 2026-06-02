'use client';

import { useState } from 'react';

export function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={copy}
      aria-label="Copy to clipboard"
      className="
        ml-3 px-3 py-1 rounded text-xs font-mono
        bg-neutral-800 hover:bg-neutral-700
        text-neutral-400 hover:text-neutral-200
        border border-neutral-700 hover:border-neutral-500
        transition-all duration-150 shrink-0
      "
    >
      {copied ? 'copied!' : 'copy'}
    </button>
  );
}
