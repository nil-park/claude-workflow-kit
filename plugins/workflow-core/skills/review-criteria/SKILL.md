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
- Hunt for potential/latent bugs, not just surface issues.
- Flag memory leaks and inefficient / wasteful code.
- Flag error-suppression that hides a problem instead of fixing it (e.g. Python
  `# type: ignore`, `# noqa`, bare `except:` / `except Exception: pass`) — only
  when the underlying issue should have been fixed.
- Flag vacuous tests that assert nothing meaningful.
- Flag leftover debug output, dead or commented-out code, and stray TODOs.
- Flag code that doesn't match the surrounding style and naming.
- For client implementations (HTTP/RPC/외부 서비스 클라이언트 등), check that a
  retry policy is properly attached — retries on transient failures, sensible
  backoff, and bounded attempts. Flag missing or misconfigured retry handling.
- Circular dependencies between source files. Check whether an import the diff
  adds closes a cycle — don't audit the whole dependency graph.
- Judge docs and comments against the `docs-standards` skill's criteria. Scope it
  to what the diff touches — never ask for a repo-wide reorganization or a change
  to the repo's existing convention. This axis produces filler easily, so flag
  only what's genuinely hard to read or bound to go stale.

## Reporting

- Don't manufacture findings to hit some count — flag only what's genuinely wrong
  or clearly worth changing, no filler. If nothing's wrong, say so.
- Lead with the conclusion (BLUF / 두괄식). Write the review to its own markdown
  doc — one per review, and one per round when reviewing in a loop.
