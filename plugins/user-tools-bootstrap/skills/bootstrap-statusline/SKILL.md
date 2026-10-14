---
name: bootstrap-statusline
description: >-
  환경에 설치된 CLI(Claude Code, Antigravity)를 감지하고, 사용자에게 설치 경로를 확인한 뒤 statusline
  Go 바이너리를 빌드하여 각 CLI 설정 파일에 등록한다. 바이너리가 이미 있으면 main의 최신 커밋으로 새로 빌드한다.
disable-model-invocation: true
---

## 1. CLI 환경 감지

어느 환경에서 실행하든 `claude`와 `agy` CLI의 설치 여부를 각각 확인한다.

- `claude`: `which claude`(또는 Windows의 `where.exe claude`)가 성공하거나 `~/.claude/` 디렉터리가 있는지 확인한다.
- `agy`: `which agy`(또는 Windows의 `where.exe agy`)가 성공하거나 `~/.gemini/antigravity-cli/` 디렉터리가 있는지 확인한다.

## 2. 바이너리 빌드 디렉터리 확인

감지된 환경에 맞춰 바이너리를 둘 디렉터리를 사용자에게 묻고 상의하여 결정한다.

- 설치 후보 경로:
  - `~/.claude/bin` (Claude Code 중심 환경)
  - `~/.gemini/antigravity-cli/bin` (Antigravity 중심 환경)
  - `~/.local/bin` (공통 사용자 바이너리 환경)
  - 또는 사용자가 직접 지정한 디렉터리

선택된 디렉터리를 `<bin-dir>`로 삼고, 디렉터리가 없으면 생성한다.

## 3. 바이너리 빌드

결정된 디렉터리에 바이너리(`statusline`, Windows에서는 `statusline.exe`)를 빌드한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/statusline@latest
```

- 바이너리의 절대 경로를 파악한다. 경로 구분자는 항상 슬래시(`/`)로 표기한다.

## 4. 설정 파일 등록

1단계에서 확인된 도구의 설정 파일에 statusline을 등록한다. 둘 다 설치되어 있으면 양쪽 모두에 등록한다.

### Claude Code (`claude` 확인 시)

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

- Windows에 Git Bash가 있으면 Claude Code는 이 명령을 Git Bash로 실행하므로 경로에 슬래시를 사용한다.
- `refreshInterval`이 `1`이어야 세션 유휴 상태에서도 idle 세그먼트가 1초마다 갱신된다.

### Antigravity (`agy` 확인 시)

`~/.gemini/antigravity-cli/settings.json`의 `statusLine`을 아래 값으로 설정한다.

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

## 5. 결과 보고

- 바이너리가 빌드된 위치와 어느 CLI(Claude Code, Antigravity)의 설정 파일에 등록했는지 보고한다.
- 이 스킬을 다시 부르면 바이너리를 main의 최신 커밋으로 새로 빌드한다고 안내한다.
