'use client';

import { useState, useEffect } from 'react';

const ALL_OBS = [
  {
    kind: 'technical', category: 'security',
    text: "jwt-decode 3.1.2 — CVE-2022-21803, CVSS 7.5. Run 'flint fix 12'.",
  },
  {
    kind: 'human', category: 'burnout',
    text: '3h 42m without a break. Same file opened 14 times. Step away.',
  },
  {
    kind: 'win', category: 'quality',
    text: 'Clean PR — tight commits, tests green, no dead code. Ship it.',
  },
  {
    kind: 'technical', category: 'git hygiene',
    text: "Last 4 commits: 'fix', 'fix 2', 'wip', 'wip 2'. Unreadable in 6 months.",
  },
  {
    kind: 'human', category: 'lone wolf',
    text: 'No commits in 4 days. Everyone else pushed yesterday. Check in?',
  },
  {
    kind: 'technical', category: 'spiral',
    text: 'auth.ts edited 11 times in 28 minutes. What are you stuck on?',
  },
];

export function AnimatedCards() {
  const [visibleIdx, setVisibleIdx] = useState([0, 1, 2]);
  const [entering, setEntering] = useState<number | null>(null);
  const [exiting, setExiting] = useState<number | null>(null);
  const [nextCard, setNextCard] = useState(3);

  useEffect(() => {
    const interval = setInterval(() => {
      const outIdx = 0;
      const inCard = nextCard % ALL_OBS.length;

      setExiting(visibleIdx[outIdx]);
      setEntering(inCard);

      setTimeout(() => {
        setVisibleIdx(prev => [...prev.slice(1), inCard]);
        setExiting(null);
        setEntering(null);
        setNextCard(n => (n + 1) % ALL_OBS.length);
      }, 500);
    }, 3500);

    return () => clearInterval(interval);
  }, [visibleIdx, nextCard]);

  const displayed = visibleIdx.map((idx, pos) => ({
    obs: ALL_OBS[idx],
    idx,
    pos,
    isExiting: exiting === idx,
  }));

  return (
    <div className="space-y-2.5">
      {displayed.map(({ obs, idx, pos, isExiting }) => (
        <div
          key={`${idx}-${pos}`}
          style={{
            opacity: isExiting ? 0 : 1,
            transform: isExiting ? 'translateY(-8px)' : 'translateY(0)',
            transition: 'opacity 0.45s ease, transform 0.45s ease',
            animation: entering === idx ? 'slideUpCard 0.45s ease forwards' : undefined,
          }}
          className="border-l-2 border-white/10 rounded-r-lg bg-[#1e1e1e] border border-white/5 p-3"
        >
          <div className="flex items-center justify-between mb-1.5">
            <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded uppercase tracking-wider bg-white/5 text-neutral-500 mono">
              {obs.kind} · {obs.category}
            </span>
            <span className="flex items-center gap-1 text-neutral-700 text-[10px]">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              live
            </span>
          </div>
          <p className="text-neutral-300 text-[11px] leading-relaxed">{obs.text}</p>
          <div className="flex gap-2 mt-2">
            <button type="button" className="text-[10px] text-neutral-600 hover:text-neutral-300 border border-white/5 rounded px-2 py-0.5 transition-colors">Tell me more</button>
            <button type="button" className="text-[10px] text-neutral-600 hover:text-neutral-300 border border-white/5 rounded px-2 py-0.5 transition-colors">Dismiss</button>
          </div>
        </div>
      ))}
    </div>
  );
}
