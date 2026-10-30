---
name: claude-message
description: >-
  `claude-queue`로 실행 중인 Claude Code 세션을 조회하고 메시지를 보낸다.
  Claude 세션에 메시지를 보내거나 보낼 대상 세션을 확인할 때 사용한다.
---

# `claude-queue`로 Claude 세션에 메시지 전송

```text
usage:
  claude-queue list
  claude-queue send --to <name|pid> [--handshake] [--message <text> | --file <path>]
```

- `--to`에는 `claude-queue list`의 `NAME` 열 값을 넣는다.
- 대상 세션에 보내는 첫 메시지에는 `--handshake`를 붙인다.
  - 보낼 본문이 있으면 `--message`나 `--file`과 함께 넘기고, 없으면 단독으로 넘긴다.
  - agy가 재시작된 뒤에도 다시 붙인다.
- 여러 줄이거나 긴 메시지는 `--file`로 보낸다.
- 다른 지시 없이 이 스킬만 호출되면, `claude-queue list`에서 `CWD`가 현재 작업 디렉터리인 Claude 세션을 찾아 `--handshake`만 보낸다.
  - 해당 세션이 없거나 둘 이상이면 사용자에게 묻는다.
