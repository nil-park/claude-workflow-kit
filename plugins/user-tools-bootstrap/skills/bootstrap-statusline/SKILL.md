---
name: bootstrap-statusline
description: >-
  statusline Go 바이너리를 `~/.claude/bin`에 빌드하고, `~/.claude/settings.json`의 `statusLine`이 그
  바이너리를 실행하도록 설정한다. 바이너리가 이미 있으면 main의 최신 커밋으로 새로 빌드한다.
disable-model-invocation: true
---

아래 명령으로 `~/.claude/bin/statusline`(Windows에서는 `statusline.exe`)을 빌드한다.

```bash
GOBIN=~/.claude/bin go install github.com/nil-park/claude-workflow-kit/statusline@latest
```

`~/.claude/settings.json`의 `statusLine`을 아래 값으로 설정한다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "<바이너리의 절대 경로>",
    "refreshInterval": 1
  }
}
```

- `command`의 경로 구분자는 슬래시로 쓴다(예: `C:/Users/<사용자>/.claude/bin/statusline.exe`).
  - Windows에 Git Bash가 있으면 Claude Code는 이 명령을 Git Bash로 실행한다.
  - Git Bash는 역슬래시를 이스케이프 문자로 해석한다.
- `refreshInterval`이 `1`이어야 세션이 유휴 상태일 때도 idle 세그먼트가 1초마다 갱신된다.

결과를 보고할 때, 이 스킬을 다시 부르면 바이너리를 main의 최신 커밋으로 새로 빌드한다고 안내한다.
