---
name: pre-push
description: >-
  보호 브랜치(main/master)로 직접 push하는 것을 차단해 PR을 거치게 하는 git pre-push 훅을
  설치한다. 이렇게 부를 때 쓴다: "pre-push 훅 깔아줘", "main 직접 push 막아줘".
---

# Pre-push 가드

- `git config core.hooksPath`가 설정돼 있으면(Husky 등) 어디에 설치할지 나에게 되묻는다.
- 그 자리에 `pre-push` 훅이 이미 있으면 덮어쓰지 말고 나에게 보여준 뒤 병합한다.

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
