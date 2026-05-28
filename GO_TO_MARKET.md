# Go-To-Market Strategy
## Flint — Free Rollout Plan

**Version:** 0.1
**Status:** Draft
**Last updated:** May 2026

---

## The Philosophy

Charge nothing until developers can't imagine working without it.

That is the only bar that matters. Not "is it useful." Not "do people like it." Do they feel the absence of it when it's gone. That is when you charge.

Everything in this plan is in service of reaching that moment as fast as possible with as many developers as possible.

---

## Phase 1 — Private Beta (Weeks 1–8)

### Goal
Find 20–30 developers who will tell you the truth.

Not people who will be polite. Not friends who will say it's great. Developers who will tell you when it's wrong, when it's annoying, when it missed something, and — most importantly — when it said something that stopped them in their tracks.

### Who to target
- Mid-level developers (2–5 years), solo or small teams
- Active on GitHub, Twitter/X, or dev communities
- Working on side projects or early-stage startups — they have full ownership and can install freely
- Ideally across a mix of roles: web dev, DevOps, mobile, data

### How to reach them
- Direct outreach — personal, not a mass email. One paragraph. What Flint is. Why you thought of them specifically. A link to install.
- Dev Twitter/X — post the core idea, not the product. "What if your editor had a senior dev watching quietly and speaking up when it mattered?" See who responds.
- Dev communities — specific ones: Indie Hackers, small Discord servers, local dev communities. Not Hacker News yet — too early, one shot.

### What you're measuring
- Do they install it more than once
- What is the first observation they find genuinely useful
- What observations do they dismiss immediately
- Do they enable the human layer — and if so, what do they say about it
- What do they wish Flint had noticed that it didn't

### What you give them
- Direct access to you. Not a feedback form — a direct line.
- Early access to every new feature before anyone else, forever.
- Their name in the credits if they want it.

---

## Phase 2 — Public Free Beta (Months 2–5)

### Goal
Get to 1,000 active developers. Learn what "active" means for Flint specifically.

### Everything is free. No limits. No paywalls.
All three awareness levels — Spark, Flame, Forge — free. Human intelligence layer free. Every tool, every feature, every update. The goal is trust and data, not revenue.

### How to grow
**Developer word of mouth is the only channel that matters at this stage.**
Developers do not trust marketing. They trust other developers. The product has to be good enough that people tell each other.

Specific tactics:
- **The `flint update` share loop** — when a developer uses `flint update` to generate a stakeholder summary, there is a subtle "generated with Flint" line at the bottom (opt-out, not opt-in). Every summary that goes to a PM or founder is a passive impression.
- **The pre-commit checkpoint tweet** — when Flint catches something real before a commit, developers will want to share it. Make it easy: `flint share` generates a shareable, anonymised snippet of what Flint caught.
- **Show don't tell** — short videos of Flint actually catching something. Real sessions, real observations. Not demos. Real work.
- **The human layer story** — when Flint's human intelligence catches something meaningful ("it noticed I was burning out before I did"), that is a story worth telling. With permission, these stories are the most powerful content Flint can have.

### Communities to seed
- Indie Hackers — solo devs, high ownership, vocal
- Dev Twitter/X — highest reach in the developer community
- Specific Discord servers — not generic ones, niche ones where senior devs talk
- Dev.to and Hashnode — longer form, developer-written content about using Flint
- Local developer meetups — especially in Lagos, London, San Francisco, Berlin

### What you're measuring
- Weekly active users (used at least once in the last 7 days)
- Awareness upgrade rate (Spark → Flame → Forge)
- Human layer opt-in rate
- Observation dismiss rate by category (tells you what's wrong)
- `flint update` usage (tells you who the stakeholder audience is)
- Retention at day 7, day 14, day 30

### The signal you're looking for
At least 30% of users are still active at day 30. That is the bar. Below that — the core product needs work before anything else. Above that — you have something.

---

## Phase 3 — Controlled Growth (Months 6–12)

### Goal
Get to 10,000 active developers before charging anyone.

### Still free. All of it.
The instinct to monetise at this stage is strong and almost always wrong. 1,000 developers who trust Flint completely is worth more than 10,000 who feel nickel-and-dimed.

### How to grow from 1,000 to 10,000
- **Hacker News Show HN** — by this point you have real usage data, real stories, real observations Flint has caught. A Show HN post with genuine substance lands differently than a launch announcement.
- **Product Hunt launch** — timing matters. Launch when you have enough social proof that the comments section does the selling for you.
- **The VS Code Marketplace and npm** — organic discovery. Developers searching for tools find Flint. The description, the screenshots, the reviews — all of this compounds.
- **Content from real users** — developers writing about Flint in their own words, on their own platforms. Amplify everything. Ask permission, share everything.
- **The team feature preview** — by month 6, start talking publicly about what Flint will do for teams. The individual users start pulling it into their companies.

### What you're measuring
- Same as Phase 2 plus:
- Net Promoter Score — do developers recommend Flint unprompted
- Which observation categories drive the most retention
- Which user segments have the highest 30-day retention

---

## Phase 4 — Monetisation (Month 12+)

### The bar to charge
Two conditions must both be true before introducing any paid tier:

1. **30-day retention above 40%** — developers are still using Flint a month after installing it
2. **At least 10 unprompted "I can't work without this"** — not survey responses, not prompted feedback. Organic statements from developers who were not asked.

Until both conditions are met — stay free.

### How to introduce pricing without destroying trust
- **Announce it 60 days before it happens.** Not a surprise. Developers hate surprises from tools they depend on.
- **Everyone currently active gets grandfathered.** Full Forge access, free, forever. This is not optional — it is the thing that turns free users into advocates when you start charging.
- **The free tier stays genuinely useful.** Spark remains free and always will. Not crippled-free. Genuinely useful free. The kind that makes developers recommend Flint to someone who can't afford to pay.

### Pricing model (when ready)
- **Spark** — free forever. Manual tools only. No daemon.
- **Flame** — $9/month. Daemon on current project. Pre-commit checkpoints. Full toolkit.
- **Forge** — $18/month. Everything. Human intelligence layer. Cross-project memory. Error dictionary. Stakeholder translation.
- **Team** — $15/seat/month. Forge for everyone plus team features (v4).

Annual pricing at 2 months free (effectively 17% discount).

---

## What Free Rollout Is Not

- It is not indefinite. There is a clear monetisation trigger (Phase 4 conditions).
- It is not unlimited. The goal of free is trust and data, not infinite user acquisition cost.
- It is not charity. Every free user is generating the signal that makes Flint better and the case that makes investors confident.

---

## The One Thing That Kills This Plan

Shipping too early.

If Flint goes public before the daemon is genuinely useful — before it catches something real that the developer didn't ask for — the first impression is a chatbot wrapper and the word of mouth is negative. Negative word of mouth in developer communities lasts years.

The private beta (Phase 1) exists entirely to prevent this. The product that goes to Phase 2 must have already caught something real for at least 10 developers who didn't ask for it.

That is the gate. Nothing opens until it passes.

---
