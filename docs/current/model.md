---
title: the domain model
created: 2026-07-14
---

# the domain model

`internal/model` is vane's pure domain layer: types and functions with no I/O, no bytes, no rendering, no transport. Everything here is testable as pure functions, and is — the ordering and flag semantics of the convention live entirely in this package's test suite.

## states

`State` covers the seven lifecycle states; `LaneOrder` fixes the canonical presentation order: inbox, horizon, researching, building, evaluating, done, dropped. `ParseState` accepts exactly those seven strings and nothing else.

## the slug rule

`Slug(title)` implements the convention's ASCII-mechanical rule over code points: map `A`–`Z` to `a`–`z`; keep `a`–`z`, `0`–`9`, space, hyphen; discard every other code point; spaces become hyphens; hyphen runs collapse; hyphens trim from the ends. The spec's normative vectors — `Retry Semantics (v2)` → `retry-semantics-v2`, and `naïve K-scale` → `nave-scale` with the Kelvin-sign discard asserted — are pinned in the tests. A title that reduces to nothing returns the empty string; the caller decides what that means (hand-picked filenames, capture instructions).

## flags

`Flag` carries a `FlagKind` and a diagnostic: `FlagMalformed` for schema-table failures, `FlagFilenameMismatch` for a filename that isn't the slug of its readable title. `MismatchesSlug(filename, title)` owns the mismatch decision, including the exemption: a title whose slug is empty legitimately carries a hand-picked filename and never flags, under any name.

## the board computation

`ComputeBoard(cards, order)` is the one place the ordering semantics live. Input is a set of `CardInput` classifications (filename, readable state or unreadable, valid created date or absent, title, flags) plus the parsed `order.yaml` lane lists; output is lanes in lifecycle order plus a disposition for every order entry.

The evaluation runs in the spec's fixed order:

1. **Discard invalid entries.** An entry naming a nonexistent file is prunable unconditionally. A lane-mismatch entry is prunable only with positive evidence — the card exists and parses with a valid, *different* state. An entry for an existing card whose state can't be read is retained.
2. **Effective lanes.** An unreadable-state card belongs to inbox for all ordering purposes. Its retained entry in inbox participates normally — it actively ranks the card. Its retained entries in any other lane are *inert*: held on disk, participating in no duplicate resolution, shadowing nothing.
3. **First-occurrence-wins** runs among the surviving, participating entries; later duplicates are prunable.
4. **The unranked tail** — cards in the lane absent from its surviving list — sorts `created` ascending, then filename ascending, with undated cards after every dated card, by filename.

Each lane reports `RankedCount`, so consumers know where the ranked prefix ends without inferring it. The dispositions (`EntryActive` / `EntryInert` / `EntryPrunable`) are a report: the workspace layer uses them for opportunistic pruning on its next order write; the model never decides when to prune.
