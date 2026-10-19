---
name: bootstrap-agy-compaction-hook
description: >-
  Antigravity 환경을 감지하고, 사용자에게 설치 경로를 확인한 뒤 agy-compaction-hook Go 바이너리를 빌드하여
  ~/.gemini/config/hooks.json에 등록한다. 바이너리가 이미 있으면 main의 최신 커밋으로 새로 빌드한다.
disable-model-invocation: true
---

## 1. agy 설치 여부 확인

`agy`가 `PATH`에 있거나 `~/.gemini/antigravity-cli/` 디렉터리가 존재하는지 확인한다.

## 2. 기존 설정 및 바이너리 경로 확인

`~/.gemini/config/hooks.json` 파일에서 기존 `agy-compaction-hook` 항목과 등록된 바이너리 경로를 확인한다.

## 3. 바이너리 빌드 디렉터리 합의

바이너리를 둘 디렉터리를 사용자와 상의하여 결정한다.

- 기존에 등록된 유효한 바이너리 경로가 있으면 해당 디렉터리를 우선 제안한다.
- 기존 설정이 없으면 `~/.local/bin`을 제안한다.
- 디렉터리가 없으면 생성한다.

## 4. 바이너리 빌드

결정한 디렉터리에 `agy-compaction-hook` 바이너리를 빌드한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/agy-compaction-hook@latest
```

## 5. 설정 파일 백업 및 등록

`~/.gemini/config/hooks.json`을 `hooks.json.bak`으로 백업한 뒤 `agy-compaction-hook` 항목을 등록한다.

- 다른 훅 설정은 그대로 보존한다.
- 바이너리의 절대 경로 구분자는 항상 슬래시(`/`)로 표기한다.

```json
{
  "agy-compaction-hook": {
    "PreInvocation": [
      { "type": "command", "command": "<absolute path of the binary>" }
    ],
    "PostInvocation": [
      { "type": "command", "command": "<absolute path of the binary>" }
    ]
  }
}
```

## 6. 결과 보고

- 빌드된 바이너리 경로와 등록한 이벤트(`PreInvocation`, `PostInvocation`)를 보고한다.
- 훅은 등록 이후에 새로 시작한 세션부터 적용된다고 안내한다.
- 이 스킬을 다시 실행하면 `main`의 최신 커밋으로 바이너리를 새로 빌드한다고 안내한다.
