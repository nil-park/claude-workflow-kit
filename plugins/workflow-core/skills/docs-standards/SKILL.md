---
name: docs-standards
description: >-
  Standards for writing, placing, and restructuring project docs and code
  comments — the `docs/` directory split, decision-record filenames, prose
  compression, what belongs in a comment, and reference direction. Use when
  creating or editing a doc, deciding which directory it belongs in, or judging
  whether docs/comments are bloated or bound to go stale — e.g. "문서 작성",
  "문서 구조 정리", "이 문서 어디에 둬야 해?", "ADR 추가".
---

# Docs & comment standards

## Placement

Follow the repo's own convention wherever it has one. The layout below is the
fallback, for a kind of doc the repo doesn't have yet. Don't propose changing an
existing convention unless I ask for a docs refactor.

| Path                 | Holds                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------- |
| `docs/architecture/` | How the system is built today — whatever you'd have to update when the implementation changes |
| `docs/decisions/`    | Decision records, fixed to the moment they were made — see the filename rule below            |
| `docs/runbook/`      | Deploy and incident-response procedures                                                       |
| `docs/setup/`        | Environment setup                                                                             |
| `docs/development/`  | How to develop and verify — smoke tests and the like                                          |
| `docs/monitoring/`   | Monitoring setup, how the infra is laid out, Grafana and alert-channel links                  |

- Architecture or decision? Ask whether a change to the implementation would
  force you to edit the sentence. If it would, it's architecture. If the answer
  is "no, I'd write a new decision instead", it's a decision.
- A why that still constrains the code belongs in architecture, dissolved into
  the what — "the gateway validates the token so downstream can trust the
  header", not a rationale section bolted on. Strip it out and the next reader
  undoes the choice. `decisions/` takes the why that constrains nothing anymore:
  alternatives you rejected, the constraints that applied at the time.
- runbook vs monitoring: **where** the dashboards and alert channels live is
  monitoring; **what to do** when an alert fires is runbook.
- Keep one doc to one role. If something fits none of the six, the right move is
  a seventh directory, not squeezing it into the closest match.

## Decision records (`docs/decisions/`)

The filename says whether a doc is frozen or live. The prefix only means
anything here — everywhere else is live docs, with no prefix.

| Filename                  | Kind                         | How to change it                 |
| ------------------------- | ---------------------------- | -------------------------------- |
| `0001-<slug>.md`          | Fixed to when it was written | Never edit the body; append only |
| `<slug>.md` (non-numeric) | Live                         | Edit freely                      |

- Leave a numbered file's body alone apart from typos. Two things can drift:
  - **The decision gets overturned** → write a new numbered file, and append one
    line to the old one pointing at it (`→ superseded by 0007-<slug>.md`).
  - **The decision still holds, you've just learned more since** → append that
    under a date at the bottom of the same file.
- Numbers are 4-digit and sequential. If two branches grab the same one,
  whichever merges second bumps its number.
- Live docs **link** to numbered files, nothing more. Copy a decision's content,
  reasoning, or numbers into one and you have two originals that drift apart.

## Prose

- Reach for a diagram, a table, or an example wherever it beats writing it out.
- If a short line covers it, don't stretch it into paragraphs.
- A sentence that keeps piling on conditions and parentheticals for the sake of
  precision buries the point it was making. Split it.
- Never use checkboxes (`- [ ]`). They record progress, and progress is stale by
  the next commit. Use plain bullets.
- None of this applies to decision docs. Trade-offs, alternatives you rejected,
  and why a workaround exists all stay, however long they run — compress the
  writing, never the substance.

## Comments

- Skip comments that restate the code, and assumptions the types, names, and
  asserts already carry. Keep the why — that's the part code can't show.
- Never restate a doc in a comment. Whatever you write twice, someone will
  update on one side only.
- The doc is the original. A comment can link to a doc, but a doc shouldn't point
  back at a code location — that's what makes the reference bidirectional.
  Naming a symbol is fine.

## References

- Spell out something that goes stale when the implementation changes — exact
  numbers, signatures, the order of internal steps — only where you have to, and
  link to it otherwise. Symbol names and file paths make fine link targets; a
  link pinned to a line number goes stale itself, so don't write one.
- Never paste code that has an original in this repo. The copy drifts from the
  source with nothing to catch it, and a stale snippet misleads harder than no
  snippet at all — describe the shape, or link to the source.
- Code with no original here is the opposite, and often the shortest way to say
  it: a client calling the API you're building, a sample request, a config the
  consumer writes, the commands in `setup`, `development`, and `runbook`. That's
  the doc's own content, not a copy.
- Keep docs loosely coupled. Two docs referencing each other is fine; three or
  more shouldn't close a ring.
