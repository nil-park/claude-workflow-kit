---
name: bootstrap-copilot-queue
description: >-
  Claude Code 세션이 copilot-queue로 대화형 Copilot CLI 세션에 메시지를 보낼 수
  있게 하는 copilot-message 스킬과 copilot-queue, claude-queue 바이너리를 설치한다.
disable-model-invocation: true
---

## 1. Copilot 설치 여부 확인

`copilot`이 `PATH`에 있거나 `~/.copilot/` 디렉터리가 존재하는지 확인한다.

## 2. `copilot-queue` 설치 및 업데이트

`copilot-queue`가 `PATH`에 있는지 확인한다.

- `copilot-queue`가 이미 `PATH`에 있으면 그 바이너리가 들어 있는 디렉터리를 그대로 사용한다.
- `copilot-queue`가 `PATH`에 없으면 바이너리를 설치할 디렉터리를 사용자와 합의한다.
  - `~/.local/bin`을 제안하며, `PATH`에 들어 있는 디렉터리여야 한다.
  - 디렉터리가 없으면 생성한다.
- 이 스킬을 실행할 때마다 아래 명령으로 `copilot-queue`를 설치하거나 최신 버전으로 업데이트한다.
  - `go`가 없으면 사용자에게 설치를 요청한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/copilot-queue@latest
```

## 3. `claude-queue` 확인 및 설치

`claude-queue`가 `PATH`에 있는지 확인한다.

- `claude-queue`가 이미 `PATH`에 있으면 건드리지 않는다.
- `claude-queue`가 `PATH`에 없으면 Copilot이 회신할 때 사용할 수 있도록 2단계의 디렉터리에 설치한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/claude-queue@latest
```

## 4. `copilot-message` 스킬 설치

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래의 `copilot-message.md`를 읽어 `~/.claude/skills/copilot-message/SKILL.md`에 쓴다. 기존 파일이 있으면 덮어쓴다.

## 5. 결과 보고

- 설치한 스킬 경로(`~/.claude/skills/copilot-message/SKILL.md`)를 보고한다.
- `copilot-queue`와 `claude-queue`를 새로 설치했는지, 기존 바이너리나 경로를 재사용했는지 보고한다.
- Copilot 세션은 `copilot --ui-server --port <port>`로 시작해야 한다고 안내한다.
