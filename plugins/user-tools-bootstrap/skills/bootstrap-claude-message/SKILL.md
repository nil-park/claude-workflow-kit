---
name: bootstrap-claude-message
description: >-
  agy 세션이 claude-queue로 Claude Code 세션에 메시지를 보낼 수 있게 하는 claude-message
  agy 스킬을 설치하고, claude-queue가 PATH에 없으면 함께 설치한다.
disable-model-invocation: true
---

## 1. agy 설치 여부 확인

`agy`가 `PATH`에 있거나 `~/.gemini/antigravity-cli/` 디렉터리가 존재하는지 확인한다.

## 2. `claude-queue` 확인 및 설치

`claude-queue`가 `PATH`에 있는지 확인한다.

- `claude-queue`가 이미 `PATH`에 있으면 건드리지 않는다.
- `claude-queue`가 `PATH`에 없으면:
  1. 바이너리를 설치할 디렉터리를 사용자와 합의한다. `~/.local/bin`을 제안하며, `PATH`에 들어 있는 디렉터리여야 한다. 디렉터리가 없으면 생성한다.
  2. 합의한 디렉터리에 `claude-queue` 바이너리를 설치한다. `go`가 없으면 사용자에게 설치를 요청한다.

```bash
GOBIN=<bin-dir의 절대 경로> go install github.com/nil-park/claude-workflow-kit/claude-queue@latest
```

## 3. `claude-message` 스킬 설치

이 스킬이 로드될 때 시스템이 알려주는 베이스 디렉터리(`Base directory for this skill`) 아래의 `claude-message.md`를 읽어 `~/.gemini/antigravity-cli/skills/claude-message/SKILL.md`에 쓴다. 기존 파일이 있으면 덮어쓴다.

## 4. 결과 보고

- 설치한 스킬 경로(`~/.gemini/antigravity-cli/skills/claude-message/SKILL.md`)를 보고한다.
- `claude-queue`를 새로 설치했는지, 이미 `PATH`에 있었는지 보고한다.
- 새로 여는 agy 세션부터 이 스킬을 사용할 수 있다고 안내한다.
