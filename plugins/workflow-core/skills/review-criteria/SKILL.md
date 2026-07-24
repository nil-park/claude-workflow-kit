---
name: review-criteria
description: >-
  Apply my standing code-review criteria to a PR/MR, branch, or working diff
  (including self-review). Use whenever I ask for a review — e.g. "6번 PR 리뷰해줘",
  "N번 MR 리뷰", "리뷰해줘", "브랜치 리뷰", "셀프 리뷰 해보자", "review PR #6",
  "review this PR / MR / branch / diff". Reviews directly without subagents; flags
  only genuine issues.
---

# Code Review

Standing criteria — follow them by default without my restating them.

## Scope

- Review directly with your own analysis; never spawn subagents or use a built-in
  code-review skill/agent.
- Apply these criteria to whatever change is in front of you — a hosted PR/MR, a
  branch, or an uncommitted working diff (self-review). Read files from the local
  checkout and review the local `git diff` against the base branch.
- For a hosted PR/MR, first pull its diff, description, and discussion — `gh` for
  a GitHub PR, `glab` for a GitLab MR — then read the code from the local checkout
  (usually already synced).

## What to look for

- Read the PR/MR description and any linked docs; call out missing pieces or
  inconsistencies, and judge whether the description itself is accurate and
  adequate for the changes.
- Check the reverse direction too: does this change make any project doc
  (architecture specs, ADRs) stale or leave a decision undocumented? Flag
  docs that should be updated in the same PR/MR, linked or not. This applies
  only when implementation started without a spec doc (description only); in a
  design-first workflow the spec doc was already written and iterated before
  coding, so skip this check.
- Hunt for potential/latent bugs, not just surface issues.
- Flag memory leaks and inefficient / wasteful code.
- Flag error-suppression that hides a problem instead of fixing it — Python
  `# type: ignore`, `# noqa`, bare `except:` / `except Exception: pass`, etc.
  Only when the underlying issue should have been fixed.
- Flag vacuous tests that assert nothing meaningful.

## Reporting

- Don't manufacture findings to hit some count — flag only what's genuinely wrong
  or clearly worth changing, no filler. If nothing's wrong, say so.
- Lead with the conclusion (BLUF / 두괄식). Write the review to
  `.refs/review-<number>.md` (the PR or MR number; for a branch/self-review, use a
  descriptive slug instead), per the `refs-scratch` skill's convention.
