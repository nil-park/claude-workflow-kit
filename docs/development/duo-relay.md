# duo-relay 개발

## 전제

- Claude Code 세션의 직접 알림
  - Claude Code의 auto mode 분류기는 `agy`와 인증 토큰을 호출하는 명령을 `Auto-Mode Bypass`로 차단한다(2026-10 실측).
  - `permissions.allow`에 등록된 명령은 분류기 검사 전에 허용되므로, duo-relay 실행 규칙을 `.claude/settings.json`에 등록한다.
- agy 세션의 접속 정보
  - 대화형 agy 세션은 자기가 실행하는 셸 명령에 `ANTIGRAVITY_CONVERSATION_ID`, `ANTIGRAVITY_LS_ADDRESS`, `ANTIGRAVITY_CSRF_TOKEN`을 넣어 준다(agy 1.4.0, 2026-10 실측).
  - 공식 문서에는 이 환경 변수와 `agy agentapi` 하위 명령이 없다.
- agy에서 agy로 보내는 직접 알림(agy 1.4.2, 2026-10-20 실측)
  - 대화형 agy 세션 두 개는 각자 자체 언어 서버를 띄운다.
  - agy 세션 안에서 실행한 `push`는 endpoint 값으로 보내는 쪽의 `ANTIGRAVITY_*` 변수를 덮어써서 받는 세션에 알림을 전달한다.
  - 받는 쪽은 알림을 `sender=<보내는 쪽의 대화 ID>`로 기록한다.
- Claude Code 세션을 깨우는 방법
  - 백그라운드 명령이 끝나면 유휴 세션이 새 턴을 시작한다.
    - 공식 문서에는 이 동작이 명시되어 있지 않다(2026-10 실측).
    - 터미널 세션의 백그라운드 명령에는 실행 시간 제한이 없다([tools reference](https://code.claude.com/docs/en/tools-reference#time-limit-for-background-commands), 2026-10-17 확인).
  - inbox socket은 인증 줄 다음에 보낼 메시지 형식이 공개되지 않아 쓰지 않는다([cross-session messaging](https://code.claude.com/docs/en/cross-session-messaging#the-sessions-inbox-socket), 2026-10-16 확인).
  - channels는 research preview이고 세션을 `--channels`로 시작해야 해서 쓰지 않는다([channels](https://code.claude.com/docs/en/channels), 2026-10-16 확인).

## 위험 요소

- anti-claudeism Stop hook과의 충돌
  - anti-claudeism Stop hook은 마지막 사용자 입력 이후 `Write`나 `Edit`로 고친 파일을 위치와 관계없이 모두 검사한다.
  - Claude Code 세션이 `push --file`에 넘길 메시지 본문을 `Write`로 쓰면, 그 턴이 끝날 때 hook이 본문을 검사한다.
  - hook은 `push`가 메시지를 보낸 뒤에 실행되므로, 지적이 나오면 이미 보낸 메시지를 퇴고하는 턴이 추가로 시작된다.
  - 메시지 본문은 한국어 지시문이어서 라운드마다 지적이 나온다.
