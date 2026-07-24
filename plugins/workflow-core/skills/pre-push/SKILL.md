---
name: pre-push
description: >-
  Install a git pre-push hook that blocks direct pushes to protected
  branches (main/master), forcing changes through a PR. Use when I ask to
  set up a pre-push guard, block direct pushes to main/master, or add the
  pre-push hook to a repo — e.g. "pre-push 훅 깔아줘", "main 직접 push 막아줘".
---

# Pre-push guard

Install the hook below so direct pushes to `main`/`master` are rejected —
changes have to go through a PR.

## Install

- If `git config core.hooksPath` is set (e.g. Husky), install into that
  directory instead of `.git/hooks/`.
- If a `pre-push` hook already exists there, show it to me and merge rather
  than overwrite — don't clobber it.
- Otherwise write the script to `.git/hooks/pre-push` and make it executable
  (`chmod +x .git/hooks/pre-push`).
- `git push --no-verify` bypasses this (standard git behavior).

```bash
#!/bin/bash

protected_branches=("refs/heads/main" "refs/heads/master")

while read -r local_ref local_sha remote_ref remote_sha
do
    for branch in "${protected_branches[@]}"; do
        if [ "$local_ref" = "$branch" ] || [ "$remote_ref" = "$branch" ]; then
            echo "❌ Direct push to '${branch#refs/heads/}' branch is blocked. Please use a PR."
            exit 1
        fi
    done
done

exit 0
```
