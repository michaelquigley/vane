# AGENTS.md — vane

vane is a roadmap-as-files convention plus a Go reader tool (CLI capture + localhost board). The convention is the product; the tool is a reader. Repo: `git.hq.quigley.com/products/vane`.

## How to arrive

1. **`docs/journal/`** — agent memory, dated entries, newest first. Read the recent entries before anything else; write freely as you work. `docs/journal/2026-07-14.md` is the planning-phase handoff and names the things the artifacts don't shout.
2. **`docs/future/vane-spec.md`** — the design, converged through a seven-round mercurius arc. Its normative rules (slug algorithm, malformed semantics, `order.yaml` evaluation, hash guard, root discovery) are settled; the spec is the authority.
3. **`docs/future/vane-work-order.md`** — the implementation plan, converged through a joint six-round arc. Stages, packages, tests, contracts. Execute it; don't re-open design questions — surface them instead.
4. **`docs/current/`** — built behavior, synthesized as stages land. Empty until stage 1 lands something.

## Posture

Prototype-first. v1 exists to validate the model and the approach; hardening comes after the model proves out. Corner-case defensiveness was explicitly vetoed in review (see the `settled_decisions` guards in `mercurius.yaml` for the genre) — do not re-introduce it in code, and do not gold-plate. The write model plus the git judgment gate is the accepted safety net.

## Load-bearing rules

- **The tool never touches git.** No git imports anywhere — `.git` appears only as a filename string in root discovery. A git import is an automatic review finding.
- **dd reads, hand-patched writes.** `df/dd` is the read/validate path only; every write is surgical byte-level line patching. A write path through any YAML encoder is a review finding.
- **The convention is primary.** Nothing about the file format may exist only in the tool's understanding.
- **The tool never commits, pushes, or deletes items.** All persistence is working-tree file writes; the operator owns history.

## Process

- Work lands in the work order's stages, in order. Each stage is gated by **terminus**: run it, resolve or get an explicit veto on every finding, and bring the stage to Michael `clean`. Findings vetoes are Michael's call, recorded in conversation and journal.
- As behavior lands, synthesize it into `docs/current/` and add a `CHANGELOG.md` entry under `## Unreleased` (`FEATURE`/`CHANGE`/`FIX` prose, in-house format — not Keep a Changelog).
- Run `unfurl -i <file>` on any markdown you author or edit, unconditionally.
- Go conventions: `df/dl` for logging, `df/dd` for YAML/JSON binding. Reference implementation for the ogen/embed/UI wiring: `flo` in `../archive`.
- When the spec and work order are fully realized, they are removed from `docs/future/` (git history keeps the archaeology); still-live deferred concerns get re-synthesized into new, smaller `docs/future/` documents first.
