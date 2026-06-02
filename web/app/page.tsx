import { CopyButton } from '@/components/CopyButton';
import { AnimatedCards } from '@/components/AnimatedCards';
import { FAQ } from '@/components/FAQ';
import { HeroCanvas } from '@/components/HeroCanvas';
import { WaveGrid } from '@/components/WaveGrid';
import { OrbitalRings } from '@/components/OrbitalRings';
import { FlowField } from '@/components/FlowField';
import { LorenzAttractor } from '@/components/LorenzAttractor';
import { TorusKnot } from '@/components/TorusKnot';
import { Tesseract } from '@/components/Tesseract';
import { ScrollReveal } from '@/components/ScrollReveal';
import { ThemeToggle } from '@/components/ThemeToggle';
import {
  EyeIcon, SparklesIcon, ChatBubbleIcon,
  ShieldCheckIcon, LockClosedIcon, AdjustmentsIcon, CommandLineIcon,
  KeyIcon, DocumentTextIcon, ServerIcon, NoSymbolIcon,
  GitHubIcon, CheckIcon, ArrowRightIcon,
} from '@/components/Icons';

const GITHUB  = 'https://github.com/flintlang/flint';
const NPM_CMD = 'npm install -g @johndansu/flint';

// ── Shared helpers ───────────────────────────────────────────────────────────

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="mono text-xs text-amber-700 dark:text-amber-400/50 uppercase tracking-widest">
      {children}
    </div>
  );
}

// Geometric F-mark: stem + top bar + middle bar with arrowhead + spark dot
function FlintMark({ size = 20 }: { size?: number }) {
  const h = Math.round(size * 28 / 30);
  return (
    <svg width={size} height={h} viewBox="0 0 30 28" fill="none" aria-hidden="true">
      <path
        className="fill-amber-600 dark:fill-amber-400"
        d="M3 4 L20 4 L20 9 L8 9 L8 14 L19 14 L24.5 16.5 L19 19 L8 19 L8 24 L3 24 Z"
      />
      <circle className="fill-amber-400 dark:fill-amber-200" cx="26.5" cy="16.5" r="2" />
    </svg>
  );
}

function FlintLogo({ size = 20, muted = false }: { size?: number; muted?: boolean }) {
  return (
    <span className="inline-flex items-center gap-2">
      <FlintMark size={size} />
      <span
        className={`mono font-bold tracking-tight leading-none select-none ${
          muted ? 'text-neutral-500' : 'text-neutral-900 dark:text-neutral-100'
        }`}
        style={{ fontSize: Math.round(size * 0.82) }}
      >
        flint
      </span>
    </span>
  );
}

// ── Nav ──────────────────────────────────────────────────────────────────────

function Nav() {
  return (
    <nav className="fixed top-0 inset-x-0 z-50 border-b border-[var(--border-sm)]
      bg-[var(--bg-base)]/90 backdrop-blur-md transition-colors">
      <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
        <FlintLogo size={22} />
        <div className="flex items-center gap-6 text-sm text-neutral-500 dark:text-neutral-500">
          <a href="#how"      className="hidden md:block hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors">How it works</a>
          <a href="#features" className="hidden md:block hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors">Features</a>
          <a href="#compare"  className="hidden md:block hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors">Compare</a>
          <a href="#faq"      className="hidden md:block hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors">FAQ</a>
          <a href={GITHUB} target="_blank" rel="noopener noreferrer"
            className="flex items-center gap-1.5 hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors">
            <GitHubIcon />
            <span className="hidden sm:inline">GitHub</span>
          </a>
          <ThemeToggle />
          <a href="#install"
            className="px-4 py-1.5 rounded-full bg-amber-400 hover:bg-amber-300 text-black text-xs font-semibold transition-colors">
            Get started
          </a>
        </div>
      </div>
    </nav>
  );
}

// ── Hero ─────────────────────────────────────────────────────────────────────

function Hero() {
  return (
    <section className="relative pt-40 pb-28 px-6 overflow-hidden dot-grid">
      <HeroCanvas />
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute top-24 left-1/2 -translate-x-1/2 w-[800px] h-[500px] rounded-full bg-amber-500/6 blur-[130px] glow-pulse" />
      </div>

      <div className="relative max-w-6xl mx-auto grid lg:grid-cols-[1fr_420px] gap-16 items-center">
        <div>
          <div className="inline-flex items-center gap-2 mono text-xs text-amber-700 dark:text-amber-400/80
            border border-amber-500/20 dark:border-amber-400/15
            bg-amber-500/8 dark:bg-amber-400/5
            rounded-full px-3.5 py-1.5 mb-8">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-500 dark:bg-amber-400 animate-pulse" />
            Open source · MIT · No telemetry
          </div>

          <h1 className="text-6xl sm:text-7xl font-bold tracking-tight leading-[1.02] mb-6">
            <span className="text-neutral-900 dark:text-neutral-100 block">The senior dev</span>
            <span className="gradient-text block">that never leaves.</span>
          </h1>

          <p className="text-xl text-neutral-600 dark:text-neutral-400 leading-relaxed mb-10 max-w-lg">
            A passive intelligence layer that watches your codebase, catches what
            code review misses, and speaks up only when it matters.
            Silence is the correct output most of the time.
          </p>

          <div className="flex flex-col sm:flex-row gap-3 mb-8">
            <div className="flex items-center bg-[var(--bg-card)] border border-[var(--border-md)]
              rounded-xl px-4 py-3.5 mono text-sm text-neutral-700 dark:text-neutral-300 flex-1 min-w-0 gap-2">
              <span className="text-amber-500/50 dark:text-amber-400/40 select-none shrink-0">$</span>
              <span className="flex-1 truncate">{NPM_CMD}</span>
              <CopyButton text={NPM_CMD} />
            </div>
            <a href={GITHUB} target="_blank" rel="noopener noreferrer"
              className="flex items-center justify-center gap-2 px-5 py-3.5 rounded-xl
                border border-[var(--border-lg)]
                text-neutral-500 dark:text-neutral-400
                hover:text-neutral-800 dark:hover:text-neutral-100
                hover:border-neutral-300 dark:hover:border-white/20
                text-sm transition-all shrink-0">
              <GitHubIcon />
              Star on GitHub
            </a>
          </div>

          <div className="flex flex-wrap items-center gap-5 text-xs text-neutral-500 dark:text-neutral-600">
            <span className="flex items-center gap-1.5"><CheckIcon size={12} className="text-emerald-500" /> No signup required</span>
            <span className="flex items-center gap-1.5"><CheckIcon size={12} className="text-emerald-500" /> Zero telemetry</span>
            <span className="flex items-center gap-1.5"><CheckIcon size={12} className="text-emerald-500" /> Bring your own API key</span>
            <span className="flex items-center gap-1.5"><CheckIcon size={12} className="text-emerald-500" /> MIT licensed</span>
          </div>
        </div>

        <MockSidebar />
      </div>
    </section>
  );
}

function MockSidebar() {
  return (
    <div className="hidden lg:block">
      <div className="relative">
        <div className="absolute -inset-8 bg-amber-500/5 rounded-3xl blur-3xl pointer-events-none" />
        <div className="relative rounded-2xl overflow-hidden border border-[var(--border-md)]
          shadow-2xl shadow-black/20 dark:shadow-black/70 font-mono text-[12px]">
          <div className="flex items-center gap-1.5 px-3 py-2.5 bg-[var(--bg-title)] border-b border-[var(--border-sm)]">
            <span className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
            <span className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
            <span className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
            <span className="ml-3 text-neutral-500 dark:text-neutral-500 text-[11px]">Flint — observations</span>
            <div className="ml-auto flex items-center gap-1.5 text-[10px] text-emerald-600 dark:text-emerald-500/70">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              live
            </div>
          </div>
          <div className="bg-[var(--bg-sidebar)] p-3 min-h-[310px]">
            <AnimatedCards />
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Stats strip ───────────────────────────────────────────────────────────────

const STATS = [
  { value: '0',    label: 'bytes sent without consent' },
  { value: '6',    label: 'signal types detected' },
  { value: '3',    label: 'awareness levels' },
  { value: '~1s',  label: 'observation latency' },
  { value: 'MIT',  label: 'open source license' },
];

function StatsStrip() {
  return (
    <section className="border-y border-[var(--border-sm)] bg-[var(--bg-stats)] py-10 px-6 transition-colors">
      <ScrollReveal>
        <div className="max-w-6xl mx-auto grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-8">
          {STATS.map((s, i) => (
            <ScrollReveal key={s.label} delay={i * 70} className="text-center">
              <div className="mono text-2xl font-bold text-amber-600 dark:text-amber-400 mb-1">{s.value}</div>
              <div className="text-xs text-neutral-500">{s.label}</div>
            </ScrollReveal>
          ))}
        </div>
      </ScrollReveal>
    </section>
  );
}

// ── How it works ──────────────────────────────────────────────────────────────

const STEPS = [
  { Icon: EyeIcon,        number: '01', title: 'Watches',   body: 'A lightweight daemon tracks file edits, git events, error logs, and build output continuously, in the background. Zero setup after flint start.' },
  { Icon: SparklesIcon,   number: '02', title: 'Thinks',    body: 'Pattern-matches against CVE databases, commit history, edit frequency, and team activity — the signals that matter and are easy to miss.' },
  { Icon: ChatBubbleIcon, number: '03', title: 'Speaks up', body: 'One observation. Specific. Actionable. Then silence. Flint interrupts only when the signal is strong enough to be worth your attention.' },
];

function HowItWorks() {
  return (
    <section id="how" className="relative py-24 px-6 border-t border-[var(--border-sm)] overflow-hidden">
      {/* Silk particle flow field — 3000 particles tracing a vector field */}
      <FlowField />
      <div className="absolute inset-0 bg-[var(--bg-base)]/80 pointer-events-none" />
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>How it works</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
            Like pair programming —{' '}
            <span className="text-neutral-400 dark:text-neutral-500">without the noise.</span>
          </h2>
          <p className="text-neutral-500 mb-16 max-w-lg text-[15px]">
            Most tools wait for you to ask. Flint pays attention so you don&apos;t have to.
          </p>
        </ScrollReveal>
        <div className="grid sm:grid-cols-3 gap-1">
          {STEPS.map((s, i) => (
            <div key={s.number} className="flex flex-col sm:flex-row items-start gap-0">
              <ScrollReveal delay={i * 120} className="flex-1">
                <div className="group flex-1 rounded-2xl p-6 hover:bg-[var(--bg-card)] transition-all">
                  <div className="flex items-center gap-3 mb-5">
                    <div className="w-10 h-10 rounded-xl bg-amber-400/10 dark:bg-amber-400/8
                      border border-amber-400/20 dark:border-amber-400/15
                      flex items-center justify-center text-amber-600 dark:text-amber-400
                      group-hover:bg-amber-400/15 dark:group-hover:bg-amber-400/12 transition-colors shrink-0">
                      <s.Icon size={18} />
                    </div>
                    <span className="mono text-xs text-neutral-400 dark:text-neutral-700">{s.number}</span>
                  </div>
                  <h3 className="text-lg font-semibold text-neutral-900 dark:text-neutral-100 mb-2">{s.title}</h3>
                  <p className="text-neutral-500 leading-relaxed text-sm">{s.body}</p>
                </div>
              </ScrollReveal>
              {i < STEPS.length - 1 && (
                <div className="hidden sm:flex items-center self-center pt-4 text-neutral-400 dark:text-neutral-800">
                  <ArrowRightIcon size={16} />
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ── Wave break ────────────────────────────────────────────────────────────────

function WaveBreak() {
  return (
    <div className="relative h-56 overflow-hidden border-t border-[var(--border-sm)]">
      <WaveGrid />
      <div className="absolute inset-0 bg-gradient-to-b from-[var(--bg-base)] via-transparent to-[var(--bg-alt)] pointer-events-none" />
      <div className="absolute inset-0 bg-gradient-to-r from-[var(--bg-base)] via-transparent to-[var(--bg-base)] pointer-events-none" />
    </div>
  );
}

// ── Signal types ──────────────────────────────────────────────────────────────

const SIGNALS = [
  { kind: 'technical', category: 'security',    description: 'CVE hits in your dependency graph, matched against OSV.dev in real time.',                           example: "jwt-decode 3.1.2 — CVE-2022-21803, CVSS 7.5." },
  { kind: 'technical', category: 'spiral',      description: 'Same file edited many times in a short window — a classic debugging loop.',                          example: "auth.ts edited 11× in 28 minutes. What are you stuck on?" },
  { kind: 'technical', category: 'git hygiene', description: "Commit messages like 'fix', 'wip', and 'fix 2' spotted before they're permanent.",                  example: "Last 4 commits: 'fix', 'fix 2', 'wip', 'wip 2'." },
  { kind: 'human',     category: 'burnout',     description: 'Hours worked continuously, combined with edit churn, signals mental fatigue.',                       example: "3h 42m without a break. Same file opened 14 times." },
  { kind: 'human',     category: 'lone wolf',   description: "Engineer hasn't committed while the rest of the team is active.",                                    example: "No commits in 4 days. Everyone else pushed yesterday." },
  { kind: 'win',       category: 'quality',     description: 'Tight commits, tests green, no dead code — Flint calls it out when you nail it.',                   example: "Clean PR — tight commits, tests green. Ship it." },
];

function SignalTypes() {
  return (
    <section className="py-24 px-6 border-t border-[var(--border-sm)] bg-[var(--bg-alt)] transition-colors">
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>Signal types</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
            Six things Flint watches for.
          </h2>
          <p className="text-neutral-500 mb-16 max-w-lg text-[15px]">
            Three categories. All grounded in what&apos;s actually happening in your repo — not heuristics.
          </p>
        </ScrollReveal>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {SIGNALS.map((s, i) => (
            <ScrollReveal key={`${s.kind}-${s.category}`} delay={i * 60}>
              <div
                className="rounded-2xl bg-[var(--bg-card)] border border-[var(--border-sm)]
                  p-5 hover:bg-[var(--bg-card-hi)] transition-colors h-full">
                <div className="mb-3">
                  <span className="mono text-[10px] font-semibold px-1.5 py-0.5 rounded uppercase tracking-wider
                    bg-black/[0.05] dark:bg-white/5 text-neutral-600 dark:text-neutral-500">
                    {s.kind} · {s.category}
                  </span>
                </div>
                <p className="text-neutral-800 dark:text-neutral-300 text-sm font-medium leading-snug mb-2">{s.description}</p>
                <p className="mono text-[11px] text-neutral-400 dark:text-neutral-600 leading-relaxed italic">&ldquo;{s.example}&rdquo;</p>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
}

// ── Awareness levels ──────────────────────────────────────────────────────────

const AWARENESS_LEVELS = [
  { name: 'Spark', tagline: 'Manual queries only.',        body: 'The daemon stays dark. Use flint explain, flint diff, and flint check on demand. Zero background CPU. Perfect for focused work.',                              cmd: 'flint awareness spark', dot: 'bg-neutral-400',               color: 'text-neutral-600 dark:text-neutral-400', featured: false },
  { name: 'Flame', tagline: 'Watching your current repo.', body: 'The daemon monitors your active project — file changes, git events, error logs, build output. Observations fire only when the signal is strong.',             cmd: 'flint awareness flame', dot: 'bg-amber-500 dark:bg-amber-400', color: 'text-amber-700 dark:text-amber-400',   featured: true  },
  { name: 'Forge', tagline: 'Watching everything.',        body: 'Daemon covers all configured repos. CVE scans cross every dependency manifest. For developers running multiple active services.',                              cmd: 'flint awareness forge', dot: 'bg-orange-500',               color: 'text-orange-700 dark:text-orange-400', featured: false },
];

function AwarenessLevels() {
  return (
    <section className="relative py-24 px-6 border-t border-[var(--border-sm)] overflow-hidden transition-colors">
      {/* Lorenz chaos attractor — the butterfly builds itself as you watch */}
      <LorenzAttractor />
      <div className="absolute inset-0 bg-[var(--bg-base)]/83 pointer-events-none" />
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>Awareness</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
            You control how much Flint watches.
          </h2>
          <p className="text-neutral-500 mb-16 max-w-lg text-[15px]">
            Three levels of daemon activity. Switch any time. One command.
          </p>
        </ScrollReveal>
        <div className="grid sm:grid-cols-3 gap-4">
          {AWARENESS_LEVELS.map((lv, i) => (
            <ScrollReveal key={lv.name} delay={i * 100}>
              <div
                className={`relative rounded-2xl border bg-[var(--bg-card)] p-6 flex flex-col transition-colors h-full
                  ${lv.featured ? 'border-amber-400/25 ring-1 ring-amber-400/20' : 'border-[var(--border-sm)]'}`}>
                {lv.featured && (
                  <div className="absolute -top-3 left-6">
                    <span className="mono text-[10px] font-bold bg-amber-400 text-black px-2.5 py-0.5 rounded-full tracking-wider">DEFAULT</span>
                  </div>
                )}
                <div className="flex items-center gap-2.5 mb-4">
                  <span className={`w-2.5 h-2.5 rounded-full ${lv.dot} shrink-0 ${lv.featured ? 'animate-pulse' : ''}`} />
                  <span className={`mono font-bold text-xl ${lv.color}`}>{lv.name}</span>
                </div>
                <p className="text-neutral-800 dark:text-neutral-200 text-sm font-semibold mb-2">{lv.tagline}</p>
                <p className="text-neutral-500 text-sm leading-relaxed mb-6 flex-1">{lv.body}</p>
                <div className="mono text-[11px] bg-[var(--bg-terminal)] border border-[var(--border-sm)] rounded-xl px-3.5 py-2.5 text-neutral-500">
                  <span className="text-amber-500/50 dark:text-amber-400/40 mr-1.5 select-none">$</span>
                  {lv.cmd}
                </div>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
}

// ── Bento features ────────────────────────────────────────────────────────────

function BentoFeatures() {
  const iconBox   = 'w-10 h-10 rounded-xl bg-amber-400/10 dark:bg-amber-400/8 border border-amber-400/20 dark:border-amber-400/15 flex items-center justify-center text-amber-600 dark:text-amber-400 mb-5 group-hover:bg-amber-400/15 dark:group-hover:bg-amber-400/12 transition-colors';
  const iconBoxSm = 'w-9 h-9 rounded-lg bg-amber-400/10 dark:bg-amber-400/8 border border-amber-400/20 dark:border-amber-400/15 flex items-center justify-center text-amber-600 dark:text-amber-400 mb-4 group-hover:bg-amber-400/15 dark:group-hover:bg-amber-400/12 transition-colors';
  const card      = 'group bg-[var(--bg-card)] border border-[var(--border-sm)] rounded-2xl hover:border-[var(--border-md)] hover:bg-[var(--bg-card-hi)] transition-all';

  return (
    <section id="features" className="py-24 px-6 border-t border-[var(--border-sm)] bg-[var(--bg-alt)] transition-colors">
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>Capabilities</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
            Built for the way developers actually work.
          </h2>
          <p className="text-neutral-500 mb-16 max-w-lg text-[15px]">
            Eight capabilities. Zero configuration beyond flint init.
          </p>
        </ScrollReveal>

        <div className="grid grid-cols-6 gap-4">
          {/* Large — Passive */}
          <ScrollReveal className="col-span-6 lg:col-span-3">
            <div className={`${card} p-7 flex flex-col h-full`}>
              <div className={iconBox}><EyeIcon size={18} /></div>
              <h3 className="font-semibold text-neutral-900 dark:text-neutral-100 text-lg mb-2">Passive by default</h3>
              <p className="text-sm text-neutral-500 leading-relaxed mb-6 flex-1">
                No commands required once running. Flint observes as you code and only interrupts when
                the signal is worth your attention. Most hours pass in total silence.
              </p>
              <div className="mono text-[11px] bg-[var(--bg-terminal)] border border-[var(--border-sm)] rounded-xl p-4 space-y-1.5">
                <div className="flex items-center justify-between">
                  <span className="text-neutral-400 dark:text-neutral-600">flint status</span>
                  <span className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-500/70 text-[10px]">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" /> running
                  </span>
                </div>
                <div className="text-neutral-500">● watching <span className="text-neutral-700 dark:text-neutral-300">auth-service/</span></div>
                <div className="text-neutral-400 dark:text-neutral-700">✓ 0 observations · last 2 hours</div>
                <div className="text-neutral-400 dark:text-neutral-700">✓ 0.1% CPU · 18 MB RAM</div>
              </div>
            </div>
          </ScrollReveal>

          {/* Large — Human layer */}
          <ScrollReveal className="col-span-6 lg:col-span-3" delay={80}>
            <div className={`${card} p-7 flex flex-col h-full`}>
              <div className={iconBox}><SparklesIcon size={18} /></div>
              <h3 className="font-semibold text-neutral-900 dark:text-neutral-100 text-lg mb-2">Human intelligence layer</h3>
              <p className="text-sm text-neutral-500 leading-relaxed mb-6 flex-1">
                Detects burnout (3+ hours without a break), debugging spirals, and lone-wolf isolation —
                the signals no linter, code review, or AI autocomplete ever catches.
              </p>
              <div className="border-l-2 border-neutral-300 dark:border-white/10 rounded-r-xl bg-[var(--bg-inset)] border border-[var(--border-sm)] p-4">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded uppercase tracking-wider
                    bg-black/[0.05] dark:bg-white/5 text-neutral-600 dark:text-neutral-500 mono">
                    human · burnout
                  </span>
                  <span className="text-neutral-400 dark:text-neutral-700 text-[10px]">just now</span>
                </div>
                <p className="text-neutral-700 dark:text-neutral-300 text-[11px] leading-relaxed">
                  3h 42m without a break. Same file opened 14 times. Step away.
                </p>
                <div className="flex gap-2 mt-2.5">
                  <button type="button" className="text-[10px] text-neutral-500 border border-[var(--border-sm)] rounded px-2 py-0.5">Tell me more</button>
                  <button type="button" className="text-[10px] text-neutral-500 border border-[var(--border-sm)] rounded px-2 py-0.5">Dismiss</button>
                </div>
              </div>
            </div>
          </ScrollReveal>

          {/* Medium tiles */}
          {([
            { Icon: ShieldCheckIcon, title: 'CVE monitoring',       body: 'Weekly dependency scans via OSV.dev. No API key needed. When a CVE hits, it fires immediately — bypassing all cooldown gates.' },
            { Icon: AdjustmentsIcon, title: 'Personal calibration', body: "Dismiss an observation and Flint offers to raise the threshold. It learns your tolerance, not the average developer's." },
            { Icon: LockClosedIcon,  title: 'Fully local',          body: 'Every byte stays on your machine. Nothing transmitted except the API calls you explicitly trigger. No telemetry. Ever.' },
          ] as const).map(({ Icon, title, body }, i) => (
            <ScrollReveal key={title} delay={i * 70} className="col-span-6 sm:col-span-3 lg:col-span-2">
              <div className={`${card} p-6 h-full`}>
                <div className={iconBoxSm}><Icon size={17} /></div>
                <h3 className="font-semibold text-neutral-900 dark:text-neutral-100 mb-2 text-[15px]">{title}</h3>
                <p className="text-sm text-neutral-500 leading-relaxed">{body}</p>
              </div>
            </ScrollReveal>
          ))}

          {/* Small tiles */}
          {([
            { Icon: CommandLineIcon,  title: 'REPL mode',    body: 'Ask anything. Flint pre-loads git history, error baselines, and file edit frequency before you type.' },
            { Icon: DocumentTextIcon, title: 'Git hygiene',  body: "Flags 'fix', 'wip', and mixed-concern commits before they become permanent artifacts." },
            { Icon: ServerIcon,       title: 'CLI + VS Code', body: 'Full feature parity. The extension adds sidebar cards and gutter decorations. Everything works from the terminal.' },
          ] as const).map(({ Icon, title, body }, i) => (
            <ScrollReveal key={title} delay={i * 70} className="col-span-6 sm:col-span-3 lg:col-span-2">
              <div className={`${card} p-6 h-full`}>
                <div className={iconBoxSm}><Icon size={17} /></div>
                <h3 className="font-semibold text-neutral-900 dark:text-neutral-100 mb-2 text-[15px]">{title}</h3>
                <p className="text-sm text-neutral-500 leading-relaxed">{body}</p>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
}

// ── REPL demo ─────────────────────────────────────────────────────────────────

function ReplDemo() {
  return (
    <section className="relative py-24 px-6 border-t border-[var(--border-sm)] overflow-hidden transition-colors">
      {/* 4D hypercube — inner cubes become outer cubes as it rotates through dimensions */}
      <Tesseract />
      <div className="absolute inset-0 bg-[var(--bg-base)]/82 pointer-events-none" />
      <div className="relative max-w-6xl mx-auto">
        <div className="grid lg:grid-cols-[1fr_500px] gap-16 items-center">
          <ScrollReveal>
            <SectionLabel>Explain anything</SectionLabel>
            <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
              Ask the question<br />you&apos;re already thinking.
            </h2>
            <p className="text-neutral-500 leading-relaxed mb-8 text-[15px]">
              <span className="mono text-neutral-700 dark:text-neutral-300">flint explain</span> opens a REPL with your
              codebase context pre-loaded — git history, error baselines, file edit frequency, and
              recent scan results. No copy-pasting. No context setup.
            </p>
            <ul className="space-y-3">
              {['"Why does auth fail after 7pm?"', '"What is this file even doing?"', '"Explain the last 3 commits in plain English."', '"Is there a safer way to handle this?"', '"What changed between these two errors?"'].map((q, i) => (
                <li key={q} className="flex items-start gap-2.5 text-sm">
                  <ArrowRightIcon size={14} className="text-amber-500/40 dark:text-amber-400/40 shrink-0 mt-0.5" />
                  <span className="mono text-neutral-500">{q}</span>
                </li>
              ))}
            </ul>
          </ScrollReveal>

          <ScrollReveal delay={120}>
            <div className="relative">
              <div className="absolute -inset-6 bg-amber-500/4 rounded-3xl blur-3xl pointer-events-none" />
              <div className="relative bg-[var(--bg-terminal)] border border-[var(--border-md)] rounded-2xl overflow-hidden font-mono text-[12px] shadow-xl shadow-black/10 dark:shadow-black/70">
                <div className="flex items-center gap-1.5 px-4 py-3 border-b border-[var(--border-sm)] bg-[var(--bg-title)]">
                  <span className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
                  <span className="text-neutral-500 text-[11px] ml-2.5">flint explain</span>
                </div>
                <div className="p-5 space-y-4">
                  <div className="flex gap-2 text-neutral-500 dark:text-neutral-400">
                    <span className="text-amber-500/50 dark:text-amber-400/50 select-none shrink-0">$</span>
                    <span>flint explain</span>
                  </div>
                  <div className="text-neutral-400 dark:text-neutral-700 text-[11px] pl-4 border-l border-[var(--border-sm)] space-y-0.5">
                    <div>Loading git log (142 commits)...</div>
                    <div>Loading error baseline · file events...</div>
                  </div>
                  <div className="flex gap-2 text-neutral-700 dark:text-neutral-200">
                    <span className="text-sky-500/50 dark:text-sky-400/50 select-none shrink-0">›</span>
                    <span>Why does auth fail after 7pm?</span>
                  </div>
                  <div className="bg-[var(--bg-inset)] rounded-xl p-4 text-[11px] leading-relaxed space-y-3 border border-[var(--border-sm)]">
                    <p className="text-neutral-600 dark:text-neutral-400">
                      <span className="text-sky-600 dark:text-sky-400/80">auth.ts:142</span> compares{' '}
                      <span className="text-amber-600 dark:text-amber-300/80">exp</span> against{' '}
                      <span className="text-amber-600 dark:text-amber-300/80">Date.now() / 1000</span>{' '}
                      with no timezone handling.
                    </p>
                    <p className="text-neutral-500 dark:text-neutral-400">
                      Tokens minted by the mobile client carry a{' '}
                      <span className="text-amber-600 dark:text-amber-300/80">+07:00</span> offset —
                      they expire 7 hours early.
                    </p>
                    <p className="text-neutral-400 dark:text-neutral-500">
                      Fix at <span className="text-sky-600 dark:text-sky-400/70">src/auth.ts:87</span>.
                    </p>
                  </div>
                  <div className="flex gap-2 text-neutral-300 dark:text-neutral-700">
                    <span className="text-sky-400/20 select-none shrink-0">›</span>
                    <span className="border-r-2 border-neutral-400 dark:border-neutral-600 h-3.5 self-center animate-pulse" />
                  </div>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
}

// ── Pre-commit ────────────────────────────────────────────────────────────────

function PreCommit() {
  return (
    <section className="py-24 px-6 border-t border-[var(--border-sm)] bg-[var(--bg-alt)] transition-colors">
      <div className="max-w-6xl mx-auto">
        <div className="grid lg:grid-cols-2 gap-16 items-center">
          <ScrollReveal delay={80} className="order-2 lg:order-1">
            <div className="relative">
              <div className="absolute -inset-6 bg-amber-500/3 rounded-3xl blur-3xl pointer-events-none" />
              <div className="relative bg-[var(--bg-terminal)] border border-[var(--border-md)] rounded-2xl overflow-hidden font-mono text-[12px] shadow-xl shadow-black/10 dark:shadow-black/70">
                <div className="flex items-center gap-1.5 px-4 py-3 border-b border-[var(--border-sm)] bg-[var(--bg-title)]">
                  <span className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
                  <span className="text-neutral-500 text-[11px] ml-2.5">pre-commit hook</span>
                </div>
                <div className="p-5 space-y-4">
                  <div>
                    <div className="text-neutral-400 dark:text-neutral-700 text-[10px] uppercase tracking-wider mb-2">One-time setup</div>
                    <div className="flex gap-2 text-neutral-500 dark:text-neutral-400">
                      <span className="text-amber-500/50 dark:text-amber-400/50 select-none shrink-0">$</span>
                      <span>flint hook install</span>
                    </div>
                    <div className="text-neutral-400 dark:text-neutral-700 mt-1 pl-4">✓ .git/hooks/pre-commit written</div>
                  </div>
                  <div className="border-t border-[var(--border-sm)]" />
                  <div>
                    <div className="text-neutral-400 dark:text-neutral-700 text-[10px] uppercase tracking-wider mb-2">On every commit</div>
                    <div className="flex gap-2 text-neutral-500 dark:text-neutral-400">
                      <span className="text-amber-500/50 dark:text-amber-400/50 select-none shrink-0">$</span>
                      <span>git commit -m &quot;fix: update token handling&quot;</span>
                    </div>
                  </div>
                  <div className="bg-[var(--bg-inset)] rounded-xl p-4 border border-[var(--border-sm)] space-y-2.5 text-[11px]">
                    <div className="text-neutral-500">● Flint pre-commit check...</div>
                    <div className="text-red-500 dark:text-red-400/80">
                      ✗ jwt-decode 3.1.2 — CVE-2022-21803 · CVSS 7.5<br />
                      <span className="text-neutral-400 dark:text-neutral-600 pl-2">Run &apos;flint fix 12&apos; for the upgrade path.</span>
                    </div>
                    <div className="text-amber-600 dark:text-amber-400/70">
                      ⚠ Last 4 commits: &apos;fix&apos;, &apos;fix 2&apos;, &apos;wip&apos;, &apos;wip 2&apos;<br />
                      <span className="text-neutral-400 dark:text-neutral-600 pl-2">These will be unreadable in six months.</span>
                    </div>
                    <div className="border-t border-[var(--border-sm)] pt-2 text-neutral-400 dark:text-neutral-600">
                      1 blocking · 1 advisory · commit blocked
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal className="order-1 lg:order-2">
            <SectionLabel>Pre-commit integration</SectionLabel>
            <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
              Catch issues before<br />they&apos;re permanent.
            </h2>
            <p className="text-neutral-500 leading-relaxed mb-8 text-[15px]">
              <span className="mono text-neutral-700 dark:text-neutral-300">flint hook install</span> writes a pre-commit hook
              that runs <span className="mono text-neutral-700 dark:text-neutral-300">flint check</span> on every commit.
              CVE hits block. Advisories warn. One command, no YAML.
            </p>
            <ul className="space-y-3 text-sm">
              {['CVEs block commits — advisories warn only', 'Skippable with git commit --no-verify when you mean it', 'Same check runs in CI via flint check --ci', 'No CI secrets needed — uses your local config'].map(item => (
                <li key={item} className="flex items-start gap-2.5 text-neutral-500">
                  <CheckIcon size={14} className="text-emerald-500/70 shrink-0 mt-0.5" />
                  {item}
                </li>
              ))}
            </ul>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
}

// ── Comparison ────────────────────────────────────────────────────────────────

const COMPARE_FEATURES = [
  { label: 'Passive — no prompting needed',  Icon: EyeIcon },
  { label: 'Catches human signals',           Icon: SparklesIcon },
  { label: 'CVE dependency monitoring',       Icon: ShieldCheckIcon },
  { label: 'Stays fully local',              Icon: LockClosedIcon },
  { label: 'Self-calibrates to you',          Icon: AdjustmentsIcon },
  { label: 'Free & open source',              Icon: KeyIcon },
];

const COMPARE_TOOLS = [
  { name: 'Flint',           highlight: true,  values: [true, true, true, true, true, true], notes: [] },
  { name: 'GitHub Copilot',  highlight: false, values: [false, false, false, false, false, false],        notes: ['You prompt it', null, null, 'Cloud model', null, 'Paid'] },
  { name: 'ESLint / linters',highlight: false, values: [true, false, 'partial', true, false, true],       notes: [null, null, 'dep plugins only', null, null, null] },
  { name: 'Code review',     highlight: false, values: [false, 'sometimes', false, true, false, null],    notes: ['Manual trigger', 'Reviewer-dependent', null, null, null, 'N/A'] },
];

function CellValue({ val, note }: { val: boolean | string | null | undefined; note?: string | null }) {
  let content: React.ReactNode;
  if (val === true)          content = <CheckIcon size={14} className="text-emerald-500 mx-auto" />;
  else if (val === false)    content = <span className="text-neutral-400 dark:text-neutral-600 block text-center text-lg leading-none">—</span>;
  else if (val === 'partial' || val === 'sometimes')
                             content = <span className="mono text-[10px] text-amber-600 dark:text-amber-600 block text-center">{val}</span>;
  else                       content = <span className="mono text-[10px] text-neutral-400 block text-center">N/A</span>;
  return (
    <div>
      {content}
      {note && <div className="text-[10px] text-neutral-400 dark:text-neutral-700 text-center mt-0.5">{note}</div>}
    </div>
  );
}

function CompareTable() {
  return (
    <section id="compare" className="py-24 px-6 border-t border-[var(--border-sm)] transition-colors">
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>How Flint differs</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-4">
            It sees what no other tool sees.
          </h2>
          <p className="text-neutral-500 mb-16 max-w-lg text-[15px]">
            Linters catch syntax. Copilot completes code. Flint watches the developer.
          </p>
        </ScrollReveal>
        <ScrollReveal delay={100}>
          <div className="overflow-x-auto -mx-6 px-6">
            <table className="w-full text-sm border-collapse min-w-[580px]">
              <thead>
                <tr className="border-b border-[var(--border-md)]">
                  <th className="text-left py-3 pr-8 w-52" scope="col"><span className="sr-only">Feature</span></th>
                  {COMPARE_TOOLS.map(t => (
                    <th key={t.name} className={`py-3 px-5 text-center text-[13px] font-semibold
                      ${t.highlight ? 'text-amber-600 dark:text-amber-400' : 'text-neutral-500'}`}>
                      {t.highlight ? (
                        <span className="inline-flex items-center gap-1.5"><FlintMark size={13} />{t.name}</span>
                      ) : t.name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {COMPARE_FEATURES.map(({ label, Icon }, fi) => (
                  <tr key={label} className="border-b border-[var(--border-sm)] hover:bg-[var(--bg-card-hi)] transition-colors">
                    <td className="py-4 pr-8">
                      <span className="flex items-center gap-2.5 text-neutral-600 dark:text-neutral-400 text-[13px]">
                        <Icon size={14} className="text-neutral-400 dark:text-neutral-600 shrink-0" />
                        {label}
                      </span>
                    </td>
                    {COMPARE_TOOLS.map(t => (
                      <td key={t.name} className={`py-4 px-5
                        ${t.highlight ? 'bg-amber-400/[0.04] dark:bg-amber-400/[0.03] border-x border-amber-400/10' : ''}`}>
                        <CellValue val={t.values[fi]} note={t.notes?.[fi]} />
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
}

// ── Testimonials ──────────────────────────────────────────────────────────────

const TESTIMONIALS = [
  { quote: "It caught a CVE in a package we'd been shipping for five weeks. The CVE was filed, OSV picked it up, and Flint fired within the hour. We'd never have caught it manually.", name: 'Priya S.', role: 'Backend Engineer, fintech startup', initials: 'PS' },
  { quote: "I laughed the first time it told me to step away. Three hours later, after I finally took a break, I came back and fixed the bug in ten minutes. It wasn't wrong.",           name: 'Marcus T.', role: 'Full-stack Developer',              initials: 'MT' },
  { quote: "The lone-wolf detection is subtle but real. One of my engineers went quiet for three days — no commits, no reviews. Flint flagged it. Turned out they were stuck. We catch that early now.", name: 'Aiko R.', role: 'Engineering Lead', initials: 'AR' },
];

function Testimonials() {
  return (
    <section className="py-24 px-6 border-t border-[var(--border-sm)] bg-[var(--bg-alt)] transition-colors">
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>What developers say</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-16">
            From the people using it.
          </h2>
        </ScrollReveal>
        <div className="grid sm:grid-cols-3 gap-4">
          {TESTIMONIALS.map((t, i) => (
            <ScrollReveal key={t.name} delay={i * 110}>
              <div className="bg-[var(--bg-card)] border border-[var(--border-sm)] rounded-2xl p-7 flex flex-col hover:border-[var(--border-md)] transition-all h-full">
                <div className="text-amber-400/30 dark:text-amber-400/20 text-5xl font-serif leading-none mb-4 select-none">&ldquo;</div>
                <p className="text-neutral-600 dark:text-neutral-400 text-sm leading-relaxed flex-1 mb-6">{t.quote}</p>
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-amber-400/10 border border-amber-400/20 flex items-center justify-center shrink-0">
                    <span className="mono text-[10px] font-bold text-amber-600 dark:text-amber-400">{t.initials}</span>
                  </div>
                  <div>
                    <div className="text-neutral-800 dark:text-neutral-200 text-sm font-medium">{t.name}</div>
                    <div className="text-neutral-400 dark:text-neutral-600 text-xs">{t.role}</div>
                  </div>
                </div>
              </div>
            </ScrollReveal>
          ))}
        </div>
      </div>
    </section>
  );
}

// ── Install ───────────────────────────────────────────────────────────────────

const CLI_STEPS = [
  { cmd: 'npm install -g @johndansu/flint', comment: '# or: brew install flintlang/tap/flint' },
  { cmd: 'flint init',                    comment: '# add your Anthropic API key' },
  { cmd: 'flint start',                   comment: '# daemon starts in background' },
];
const VSCODE_FEATURES = ['Observation cards in the sidebar', 'Gutter decorations + CodeLens hints', '"Tell me more" streaming follow-up', 'Works offline in Spark mode', 'Standalone panel when daemon is off'];

function Install() {
  return (
    <section id="install" className="relative py-24 px-6 border-t border-[var(--border-sm)] overflow-hidden transition-colors">
      {/* Torus knot — two interlocked mathematical knots counter-rotating */}
      <TorusKnot />
      <div className="absolute inset-0 bg-[var(--bg-base)]/82 pointer-events-none" />
      <div className="max-w-6xl mx-auto">
        <ScrollReveal>
          <SectionLabel>Get started</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-3">Running in three commands.</h2>
          <p className="text-neutral-500 mb-12 max-w-lg text-[15px]">macOS, Linux, and Windows (WSL2). Requires Node 18+ or Go 1.21+.</p>
        </ScrollReveal>
        <div className="grid lg:grid-cols-2 gap-6">
          <ScrollReveal delay={60}>
            <div>
              <p className="mono text-xs text-neutral-400 dark:text-neutral-600 uppercase tracking-widest mb-3">CLI</p>
              <div className="bg-[var(--bg-terminal)] border border-[var(--border-md)] rounded-2xl overflow-hidden">
                <div className="flex items-center gap-1.5 px-4 py-3 border-b border-[var(--border-sm)] bg-[var(--bg-title)]">
                  <span className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
                  <span className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
                  <span className="mono text-[11px] text-neutral-500 ml-2.5">terminal</span>
                </div>
                <div className="p-5 space-y-5">
                  {CLI_STEPS.map((s, i) => (
                    <div key={i} className="flex items-start gap-2.5">
                      <span className="mono text-amber-500/40 dark:text-amber-400/40 text-sm mt-0.5 select-none shrink-0">$</span>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="mono text-sm text-neutral-700 dark:text-neutral-200 flex-1 truncate">{s.cmd}</span>
                          <CopyButton text={s.cmd} />
                        </div>
                        <div className="mono text-xs text-neutral-400 dark:text-neutral-700 mt-0.5">{s.comment}</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </ScrollReveal>
          <ScrollReveal delay={120}>
            <div>
              <p className="mono text-xs text-neutral-400 dark:text-neutral-600 uppercase tracking-widest mb-3">VS Code Extension</p>
              <div className="bg-[var(--bg-card)] border border-[var(--border-md)] rounded-2xl p-6 h-full flex flex-col">
                <p className="text-sm text-neutral-500 leading-relaxed mb-6">
                  Observation cards, gutter decorations, CodeLens hints, and the streaming &ldquo;Tell me more&rdquo; panel — connected to the daemon automatically.
                </p>
                <ul className="space-y-2.5 mb-8 flex-1">
                  {VSCODE_FEATURES.map(item => (
                    <li key={item} className="flex items-center gap-2.5 text-sm text-neutral-600 dark:text-neutral-400">
                      <span className="w-4 h-4 rounded-full bg-amber-400/10 border border-amber-400/20 flex items-center justify-center shrink-0">
                        <CheckIcon size={10} className="text-amber-600 dark:text-amber-400" />
                      </span>
                      {item}
                    </li>
                  ))}
                </ul>
                <a href="https://marketplace.visualstudio.com/items?itemName=flintlang.flint" target="_blank" rel="noopener noreferrer"
                  className="flex items-center justify-center gap-2 px-5 py-3 bg-amber-400 hover:bg-amber-300 text-black font-semibold rounded-xl text-sm transition-colors">
                  Install from Marketplace
                </a>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  );
}

// ── FAQ ───────────────────────────────────────────────────────────────────────

function FAQSection() {
  return (
    <section id="faq" className="py-24 px-6 border-t border-[var(--border-sm)] bg-[var(--bg-alt)] transition-colors">
      <div className="max-w-3xl mx-auto">
        <ScrollReveal>
          <SectionLabel>FAQ</SectionLabel>
          <h2 className="text-4xl font-bold text-neutral-900 dark:text-neutral-100 mt-3 mb-12">Common questions.</h2>
        </ScrollReveal>
        <ScrollReveal delay={80}>
          <FAQ />
        </ScrollReveal>
      </div>
    </section>
  );
}

// ── CTA ───────────────────────────────────────────────────────────────────────

function CTABanner() {
  return (
    <section className="py-28 px-6 border-t border-[var(--border-sm)] relative overflow-hidden transition-colors">
      {/* Three orbital rings replace the globe — more dynamic, mouse-responsive */}
      <OrbitalRings />
      <div className="pointer-events-none absolute inset-0 bg-[var(--bg-base)]/78" />
      <div className="relative max-w-2xl mx-auto text-center">
        <ScrollReveal>
          <div className="inline-flex mb-6"><FlintMark size={48} /></div>
          <h2 className="text-5xl font-bold text-neutral-900 dark:text-neutral-100 mb-4 tracking-tight">
            Start in three<br /><span className="gradient-text">commands.</span>
          </h2>
          <p className="text-neutral-500 mb-10 text-lg">No account. No signup. No telemetry. Open source, MIT, always free.</p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <div className="flex items-center bg-[var(--bg-card)] border border-[var(--border-md)] rounded-xl px-4 py-3.5 mono text-sm text-neutral-700 dark:text-neutral-300 gap-2">
              <span className="text-amber-500/40 dark:text-amber-400/40 select-none shrink-0">$</span>
              <span className="flex-1 truncate">{NPM_CMD}</span>
              <CopyButton text={NPM_CMD} />
            </div>
            <a href={GITHUB} target="_blank" rel="noopener noreferrer"
              className="flex items-center justify-center gap-2 px-5 py-3.5 rounded-xl border border-[var(--border-lg)] text-neutral-500 dark:text-neutral-400 hover:text-neutral-800 dark:hover:text-neutral-100 hover:border-neutral-300 dark:hover:border-white/20 text-sm transition-all shrink-0">
              <GitHubIcon />View on GitHub
            </a>
          </div>
        </ScrollReveal>
      </div>
    </section>
  );
}

// ── Footer ────────────────────────────────────────────────────────────────────

const TRUST = [
  { Icon: NoSymbolIcon,     label: 'No telemetry' },
  { Icon: ServerIcon,       label: 'All data local' },
  { Icon: KeyIcon,          label: 'API key stays on device' },
  { Icon: DocumentTextIcon, label: 'MIT licence' },
];

function Footer() {
  return (
    <footer className="border-t border-[var(--border-sm)] bg-[var(--bg-footer)] transition-colors">
      <div className="max-w-6xl mx-auto px-6 py-8 border-b border-[var(--border-sm)]">
        <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
          {TRUST.map(({ Icon, label }) => (
            <div key={label} className="flex items-center gap-2 text-xs text-neutral-400 dark:text-neutral-700">
              <Icon size={13} /><span>{label}</span>
            </div>
          ))}
          <span className="text-neutral-300 dark:text-neutral-800 text-xs">·</span>
          <p className="text-xs text-neutral-400 dark:text-neutral-700">
            The only data that leaves: code you explicitly paste, and dependency names for CVE checks.
          </p>
        </div>
      </div>
      <div className="max-w-6xl mx-auto px-6 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-neutral-400 dark:text-neutral-600">
        <FlintLogo size={16} muted />
        <div className="flex items-center gap-6">
          {[['GitHub', GITHUB], ['License', `${GITHUB}/blob/main/LICENSE`], ['Security', `${GITHUB}/blob/main/SECURITY.md`], ['Discussions', `${GITHUB}/discussions`]].map(([label, href]) => (
            <a key={label} href={href} target="_blank" rel="noopener noreferrer"
              className="hover:text-neutral-700 dark:hover:text-neutral-300 transition-colors flex items-center gap-1.5">
              {label === 'GitHub' && <GitHubIcon size={13} />}{label}
            </a>
          ))}
        </div>
        <span className="text-neutral-400 dark:text-neutral-700">Built by developers, for developers.</span>
      </div>
    </footer>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function Page() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <StatsStrip />
        <HowItWorks />
        <WaveBreak />
        <SignalTypes />
        <AwarenessLevels />
        <BentoFeatures />
        <ReplDemo />
        <PreCommit />
        <CompareTable />
        <Testimonials />
        <Install />
        <FAQSection />
        <CTABanner />
      </main>
      <Footer />
    </>
  );
}
