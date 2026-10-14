---
name: bootstrap-statusline
description: >-
  statusline Go 바이너리를 `~/.claude/bin`에 빌드하고, 환경에 설치된 Claude Code와 Antigravity(agy)의
  설정 파일에 각각 등록한다. 바이너리가 이미 있으면 main의 최신 커밋으로 새로 빌드한다.
disable-model-invocation: true
---

## 1. 바이너리 빌드

아래 명령으로 `~/.claude/bin/statusline`(Windows에서는 `statusline.exe`)을 빌드한다.

```bash
GOBIN=~/.claude/bin go install github.com/nil-park/claude-workflow-kit/statusline@latest
```

- 바이너리의 절대 경로를 파악한다(예: Windows에서는 `C:/Users/<사용자>/.claude/bin/statusline.exe`, macOS/Linux에서는 `/home/<사용자>/.claude/bin/statusline`).
- 경로 구분자는 항상 슬래시(`/`)로 표기한다.

## 2. CLI 환경 감지 및 설정

어느 환경에서 실행하든 `claude`와 `agy` CLI의 설치 여부를 각각 확인하고, 설치되어 있는 도구의 설정 파일에 statusline을 등록한다. 둘 다 설치되어 있으면 양쪽 모두에 등록한다.

### Claude Code (`claude` 확인 시)

`which claude`(또는 Windows의 `where.exe claude`)가 성공하거나 `~/.claude/` 디렉터리가 있으면, `~/.claude/settings.json`의 `statusLine`을 아래 값으로 설정한다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "<바이너리의 절대 경로>",
    "refreshInterval": 1
  }
}
```

- Windows에 Git Bash가 있으면 Claude Code는 이 명령을 Git Bash로 실행하므로 경로에 슬래시를 사용한다.
- `refreshInterval`이 `1`이어야 세션 유휴 상태에서도 idle 세그먼트가 1초마다 갱신된다.

### Antigravity (`agy` 확인 시)

`which agy`(또는 Windows의 `where.exe agy`)가 성공하거나 `~/.gemini/antigravity-cli/` 디렉터리가 있으면, `~/.gemini/antigravity-cli/settings.json`의 `statusLine`을 아래 값으로 설정한다.

```json
{
  "statusLine": {
    "type": "command",
    "command": "<바이너리의 절대 경로>",
    "enabled": true,
    "stack_with_default": true
  }
}
```

- `stack_with_default`가 `true`여야 agy 기본 줄 아래에 두 줄 구성으로 덧붙여진다.

## 3. 결과 보고

- 어느 CLI(Claude Code, Antigravity)의 설정 파일에 등록했는지 보고한다.
- 이 스킬을 다시 부르면 바이너리를 main의 최신 커밋으로 새로 빌드한다고 안내한다.
