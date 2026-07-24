---
name: refs-scratch
description: >-
  The `.refs/` scratch-directory convention used across this kit — where to put
  local throwaway files (PR/MR body drafts, saved reviews, follow-up jottings)
  and how to name them. Referenced by the review and branch/PR/MR workflow
  skills; use when writing any temp file that shouldn't be committed.
---

# .refs scratch directory

`<repo-root>/.refs/` is local scratch — files that support the work but never get
committed. Use it for `gh`/`glab` body-file inputs, saved reviews, and
"revisit this later" jottings.

- **Never commit these files.** `.refs/` must be gitignored before you write into
  it. If it isn't yet, **ask me before adding it to `.gitignore`** — editing that
  file changes shared repo config, and this may not be a repo I control.
- **Unique, descriptive kebab-case names** — never a fixed name, or a parallel
  session clobbers it. Tie the name to its branch or subject, e.g.
  `pr-body-<branch-slug>.md`, `mr-body-<branch-slug>.md`, `review-<number>.md`.
- For a file passed to a CLI (a PR/MR body), write the content with **literal
  backticks, no escaping**, and pass the path (`--body-file <path>` /
  `-d "$(cat <path>)"`) — never inline the body on the command line.
