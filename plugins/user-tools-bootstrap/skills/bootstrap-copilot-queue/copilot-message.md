---
name: copilot-message
description: >-
  `--ui-server`로 실행 중인 GitHub Copilot CLI 세션에 `copilot-queue`로
  메시지를 보낸다. Copilot 세션에 메시지를 보내야 할 때 사용한다.
---

# `copilot-queue`로 Copilot 세션에 메시지 전송

```text
usage:
  copilot-queue send --port <port> --from <name> [--handshake] [--message <text> | --file <path>]
```

- `--from`에는 `claude-queue list`에서 `CWD`가 현재 작업 디렉터리와 일치하는 세션 이름을 넣는다.
  - 일치하는 세션이 없거나 둘 이상이면 사용자에게 묻는다.
- Copilot 포트를 모르면 사용자에게 묻는다.
- Copilot 세션에 보내는 첫 메시지에는 `--handshake`를 붙인다.
  - Copilot 세션이 재시작되었거나 회신이 오지 않을 때도 `--handshake`를 다시 붙여 보낸다.
- 회신은 이 세션에 `<cross-session-message from="copilot-<port>">` 메시지로 도착하므로, 백그라운드 대기 명령을 실행하지 않는다.
- 긴 텍스트나 여러 줄 텍스트는 `--file`로 전달한다.
