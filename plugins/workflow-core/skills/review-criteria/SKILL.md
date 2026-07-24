---
name: review-criteria
description: >-
  Apply my standing code-review criteria to a GitHub PR, branch, or working diff
  (including self-review). Use whenever I ask for a review — e.g. "6번 PR 리뷰해줘",
  "N번 PR 리뷰", "리뷰해줘", "브랜치 리뷰", "셀프 리뷰 해보자", "review PR #6",
  "review this PR / branch / diff". Reviews directly without subagents; flags only
  genuine issues.
---

# Code Review

Standing criteria — follow them by default without my restating them.

## Scope

- Review directly with your own analysis; never spawn subagents or use a built-in
  code-review skill/agent.
- For a PR: identify the project from `origin`, fetch the diff, description, and
  discussion via `gh`. The local branch is usually already synced — read files
  from the local checkout.
- For a branch, working diff, or self-review (no PR yet): review the local
  `git diff` against the base branch.

## What to look for

- Read the PR description and any linked docs; call out missing pieces or
  inconsistencies, and judge whether the description itself is accurate and
  adequate for the changes.
- Check the reverse direction too: does this change make any project doc
  (architecture specs, ADRs) stale or leave a decision undocumented? Flag
  docs that should be updated in the same PR, linked or not. This applies
  only when implementation started without a spec doc (PR description only,
  e.g. the pr-workflow escape hatch); in a design-first workflow the spec
  doc was already written and iterated before coding, so skip this check.
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
  `<repo-root>/.refs/review-<PR-number>.md` (for a branch/self-review, use a
  descriptive slug instead of the PR number) — if `.refs` isn't gitignored yet,
  add it to `.gitignore` first.
