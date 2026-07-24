---
name: refs-scratch
description: >-
  Convention for `.refs/`, the local scratch directory for files that support a
  task but must never be committed. Use whenever you need to feed multi-line text
  to a CLI flag or stdin — creating or editing a `gh`/`glab` PR/MR/issue body
  (`gh pr create`/`gh pr edit`/`gh issue create` with `--body-file <f>`,
  `glab mr create`/`glab mr update` with `-d "$(cat <f>)"`) — or to save a working
  artifact like a code review or a follow-up note. Covers where such files go, how
  to name them, and keeping them out of git.
---

# .refs scratch directory

`<repo-root>/.refs/` is local scratch — files that support the work but never get
committed. Use it for `gh`/`glab` body inputs when creating or editing a
PR/MR/issue description, saved reviews, and "revisit this later" jottings.

- **Never commit these files.** `.refs/` must be gitignored before you write into
  it. If it isn't yet, **ask me before adding it to `.gitignore`** — editing that
  file changes shared repo config, and this may not be a repo I control.
- **Unique, descriptive kebab-case names** — never a fixed name, or a parallel
  session clobbers it. Tie the name to its branch or subject, e.g.
  `pr-body-<branch-slug>.md`, `mr-body-<branch-slug>.md`, `review-<number>.md`.
- For a file passed to a CLI (a PR/MR body), write the content with **literal
  backticks, no escaping**, and pass the path (`--body-file <path>` /
  `-d "$(cat <path>)"`) — never inline the body on the command line.
