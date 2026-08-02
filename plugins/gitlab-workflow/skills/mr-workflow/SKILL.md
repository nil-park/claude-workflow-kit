---
name: mr-workflow
description: >-
  Create a branch and open/edit a GitLab MR following my step-by-step workflow.
  Use when I ask to create a branch and open an MR — e.g. "브랜치 만들고 MR
  작성하자", "MR 작성", "MR 만들자", "브랜치 만들고 MR 작성", "create the branch
  and open the MR", "open the MR", "create a draft MR".
---

# GitLab branch + MR workflow

Steps 1–2 are **interactive** — work through them with me. Once I OK starting
implementation, run steps 3–5 **autonomously** and report back only when the MR
is ready for me to merge.

1. Create the branch and push it, then open the **draft** MR with only a title
   and description. Write no implementation code in this step — "opening the MR"
   means the title and description, nothing else.
2. Write or update a spec doc under the repo's docs dir (e.g. `docs/architecture/`
   or `docs/spec/` — match whatever the repo already uses) capturing the desired
   state, based on the prompt I gave you.
   - Commit and push the spec-doc draft.
   - Iterate on the design with me until it's solid — before any implementation.
   - Escape hatch: if the change is only a few lines, get my confirmation, skip
     the spec doc, and go straight to step 3.
3. Implement, commit, and push.
4. When you judge the implementation complete, self-review it per the
   `review-criteria` skill's instructions.
   - Repeat the implement → push → self-review cycle until **three consecutive
     clean self-review passes**. Save each cycle's review to
     `.refs/mr-review-<mr>-<cycle>.md` (local scratch — don't commit it).
5. Only when it's truly ready, flip the draft MR to ready
   (`glab mr update --ready`) and tell me it's done. I merge it myself.

On its own, "open the MR" / "MR 작성" means everything up to the draft MR
(steps 1–2), not implementation. Never start implementation unless I explicitly
ask for it.

## Creating the branch + MR (step 1 mechanics)

- **Always check origin's default branch (main/master) first** — it's the
  `--target-branch`.
- Branch names: `<type>/<slug>`, e.g. `chore/add-user-ssh-key`. Keep this form
  consistently — don't embed issue numbers or Jira keys in the branch name; the
  MR title carries the Jira key and the MR body carries any issue link.
- Create via `glab`:
  ```
  glab mr create --remove-source-branch --squash-before-merge \
    --target-branch "<origin default branch: main or master>" --draft -a @me \
    -t "GAI-123 Title goes here" -d "$(cat .refs/mr-body-<branch-slug>.md)"
  ```
- Include the Jira issue key only in the **MR title** (not in individual commit
  messages). When committing directly to main, put the key in the commit message.
- Write the description to a temp file (literal backticks, no escaping) and pass
  via `-d "$(cat ...)"`; never inline `-d "..."`. Give the file a unique-per-branch
  name, e.g. `mr-body-<branch-slug>.md` — a fixed name gets clobbered by a parallel
  session. Put it in `<repo-root>/.refs/` — if `.refs` isn't gitignored yet, add
  it to `.gitignore` first.
- MR descriptions: concise, aim for ≤25 lines. Frame it as the **delta between
  the current state and the desired state** the spec doc lays out — what this MR
  changes to close that gap — not a re-explanation of the spec. No checkboxes
  (`- [ ]`) anywhere in GitLab — use plain bullet lists.
- Leave out anything that goes stale as the MR evolves — exact test counts, line
  numbers, "N of M done", "아직 미구현(not yet implemented)" — it'll be wrong by the
  next push. Describe what the MR does, not where it is in its lifecycle.

## Spec doc (step 2 mechanics)

- Lives under the repo's docs dir (`docs/architecture/`, `docs/spec/`, or
  whatever the repo already uses); captures the **desired state**, not a
  walkthrough of the implementation.
- Keep it **abstract: key decisions, their implications, and the purpose/intent.**
  Use a code snippet only when it's more concise and clear than prose; otherwise
  stay in prose and don't drift into line-by-line detail.
- When it helps, **lead with a ` ```mermaid ` diagram at the very top**, before
  the prose.

## Implementing (step 3 mechanics)

- **Once implementation has started, don't edit the spec doc** unless there's a
  serious bug in the design itself. If you notice something that might need
  changing, don't touch the doc — jot it down in an md file under `.refs/` (e.g.
  `.refs/spec-followup-<branch-slug>.md`, "this may need revising..") and surface
  it to me instead.
- **Before each commit, self-review that commit's diff on its own:**
  - Re-read the staged diff and confirm it does exactly what the commit intends —
    nothing more, nothing less. Flag any unrelated/accidental change that belongs
    in a different commit.
  - Actively look for a reason it's _wrong_ (off-by-one, null/empty input, error
    paths, wrong variable, missed edge case) rather than confirming it's right —
    assume a bug until you've checked.
  - Strip leftover debug output, dead/commented-out code, and stray TODOs.
  - Match the surrounding code's style and naming.
  - Fix trivial issues in place; surface anything uncertain or a judgment call to
    me instead of silently deciding.

## Reviewing and revising (step 4 mechanics)

- **Self-review is a loop, not a one-shot** — the common failure is calling it
  done after one pass. Review the whole MR against the `review-criteria` skill's
  criteria, revise, then **re-review from scratch**, repeating.
- **Exit condition: three consecutive clean passes.** A pass is clean only if it
  surfaces nothing worth changing. The counter **resets to zero on any edit** —
  including a trivial one-line fix, and including edits made in response to the
  pass itself. So the last three passes must run back-to-back over an untouched
  tree.
- Each pass must be a genuine fresh review, not a rubber stamp of the previous
  one — re-read the full diff and actively hunt for a reason it's wrong. Passes 2
  and 3 exist to catch what pass 1 missed, so don't shortcut them just because
  the tree didn't change.
- Expect that to take a while: in practice fixes keep surfacing for ~5 rounds of
  revision, so an early "looks complete" is almost always premature. The number
  is calibration, not a target or cap — the only exit is three genuinely empty
  passes in a row.

## Marking ready (step 5 mechanics)

- Flip the draft to ready with `glab mr update --ready`, then tell me. I merge it
  myself.

## Editing an MR description

- Use `glab mr update` (NOT `glab mr edit`), same temp-file `-d "$(cat ...)"` rule.
