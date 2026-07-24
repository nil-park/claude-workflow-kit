---
name: pr-workflow
description: >-
  Create a branch and open/edit a GitHub PR following my step-by-step workflow.
  Use when I ask to create a branch and open a PR — e.g. "브랜치 만들고 PR
  작성하자", "PR 작성", "PR 만들자", "브랜치 만들고 PR 작성", "create the branch
  and open the PR", "open the PR", "create a draft PR".
---

# GitHub branch + PR workflow

Steps 1–2 are **interactive** — work through them with me. Once I OK starting
implementation, run steps 3–5 **autonomously** and report back only when the PR
is ready for me to merge.

1. Create the branch and push it.
2. Write or update a spec doc under `docs/architecture/` capturing the desired
   state, based on the prompt I gave you.
   - Commit and push the spec-doc draft, then open a **draft** PR.
   - Iterate on the design with me until it's solid — before any implementation.
   - Escape hatch: if the change is only a few lines, get my confirmation, skip
     the spec doc, and go straight to step 3 — opening the draft PR at the first
     implementation commit instead.
3. Implement, commit, and push.
4. When you judge the implementation complete, self-review it per the
   `review-criteria` skill's instructions.
   - Repeat the implement → push → self-review cycle until a clean self-review
     pass. Save each cycle's review to `.refs/review-<pr>-<cycle>.md`
     (local scratch — don't commit it).
5. Only when it's truly ready, flip the draft PR to **ready for review**
   (`gh pr ready`) and tell me it's done.

On its own, "open the PR" / "PR 작성" means everything up to the draft PR
(steps 1–2), not implementation.

## Creating the branch + PR

- **Always check origin's default branch (main/master) first** — it's the `--base`.
- Branch names: `<type>/<slug>`, e.g. `chore/add-user-ssh-key`. When a backing
  issue exists, insert its number: `<type>/<issue>-<slug>`, e.g.
  `chore/123-add-user-ssh-key`.
- Reference the issue with a plain `#123` in the PR **body** to cross-link both
  ways. Don't repeat the number in the title — keep that descriptive; the branch
  name already carries it.
- Create via `gh`:
  ```
  gh pr create --assignee @me --base main --draft \
    --title "Add user SSH key" \
    --body-file .refs/pr-body-<branch-slug>.md
  ```
- PR descriptions: concise, aim for ≤25 lines. Frame it as the **delta between
  the current state and the desired state** the spec doc lays out — what this PR
  changes to close that gap — not a re-explanation of the spec. No checkboxes
  (`- [ ]`) anywhere in GitHub — use plain bullet lists.
- Leave out anything that goes stale as the PR evolves — exact test counts, line
  numbers, "N of M done" — it'll be wrong by the next push. Describe what the PR
  does, not where it is in its lifecycle.

## Spec doc

- Lives under `docs/architecture/`; captures the **desired state**, not a
  walkthrough of the implementation.
- Keep it **abstract: key decisions, their implications, and the purpose/intent.**
  Use a code snippet only when it's more concise and clear than prose; otherwise
  stay in prose and don't drift into line-by-line detail.
- When it helps, **lead with a ` ```mermaid ` diagram at the very top**, before
  the prose.

## Implementing

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

## Reviewing and revising

- **Self-review is a loop, not a one-shot** — the common failure is calling it
  done after one pass. Review the whole PR against the `review-criteria` skill's
  criteria, revise, then **re-review from scratch**, repeating until a full pass
  (taken after your last edit) surfaces nothing worth changing.
- Expect that to take a while: in practice fixes keep surfacing for ~5 rounds of
  revision, so an early "looks complete" is almost always premature. The number
  is calibration, not a target or cap — the only exit is a genuinely empty pass.

## Marking ready

- Flip the draft to ready with `gh pr ready`, then tell me. I merge it myself.

## Editing a PR description

- Use `gh pr edit --body-file <path>`, same temp-file convention as above.
