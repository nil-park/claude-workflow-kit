# claude-queue 개발

이 문서는 claude-queue의 전제와 검증 방법을 정리한다. 동작은 [claude-queue 아키텍처 문서](../architecture/claude-queue.md)에, 받는 쪽 규약은 [Claude Code 세션 inbox 프로토콜](../reference/claude-code-inbox-protocol.md)에 있다.

## 전제

- inbox 프로토콜은 공개되지 않은 규약이다.
  - 인증 줄 형식만 공식 문서에 있고, 나머지는 Claude Code 2.1.306(Windows, 2026-10-21)에서 실측하고 바이너리를 읽어 확인했다.
  - claude-queue는 최신 Claude Code를 대상으로 한다.
  - 규약이 바뀌면 reference 문서와 코드를 함께 고친다.
- macOS와 Linux 동작은 실측하지 않았다.
  - 소켓 경로, `procStart` 형식, 키 파일 위치는 바이너리에서 읽은 값으로 구현했다.
- 호출자와 받는 세션은 같은 호스트에서 같은 OS 사용자로 실행된다.
  - `~/.claude/sessions/`를 다른 호스트나 다른 pid namespace와 공유하지 않으므로 `pidDomain`은 비교하지 않는다.
  - `CLAUDE_CONFIG_DIR`은 설정하지 않는다.
- 받는 세션은 prompting 모드이고 `crossSessionInbound`를 설정하지 않았다.
  - 그 밖의 경우에는 받는 쪽이 메시지를 보류하거나 버릴 수 있다.
  - 받는 쪽이 응답을 보내지 않으므로 claude-queue는 보류나 폐기를 알 수 없다.
- 받는 세션은 claude-queue에 회신하지 못한다.
  - `from`에 실제 소켓 주소가 없으므로, 회신이 필요한 호출자는 duo-relay의 메시지 파일 같은 다른 수단으로 회신을 받는다.
- 본문은 짧은 지시문이다.
  - Claude 송신자는 직렬화 기준 약 100만 자를 넘는 메시지를 보내지 않는다.
  - claude-queue는 길이를 검사하지 않으며, 그보다 긴 메시지를 받는 쪽이 어떻게 처리하는지는 확인하지 않았다.
  - 본문의 유니코드 정규화는 하지 않는다.
- Claude Code 세션 안에서 claude-queue를 실행하면 auto mode 분류기가 키 파일 읽기를 차단할 수 있다.
  - 이 경우 `.claude/settings.json`의 `permissions.allow`에 실행 규칙을 등록한다.

## 검증 방법

- 단위 테스트: `cd claude-queue && go test ./...` (`make test`에 포함된다)
  - 전송 테스트는 Unix domain socket으로 띄운 가짜 수신자를 쓰므로 Windows에서도 실행된다.
- 스모크 테스트
  1. `make claude-queue`로 설치한다.
  2. 터미널 하나에서 Claude Code 세션을 시작하고 입력 대기 상태로 둔다.
  3. 다른 터미널에서 `claude-queue list`를 실행해 그 세션이 나오는지 확인한다.
  4. `claude-queue send --to <이름> --message "smoke test"`를 실행한다.
  5. 세션이 키 입력 없이 새 턴을 시작하고, 대화에 `<cross-session-message from="claude-queue" …>`가 들어오는지 확인한다.

## 위험 요소

- 프로토콜이 바뀌면 `send`가 성공으로 끝나도 메시지가 전달되지 않을 수 있다.
- 등록 파일이나 키 파일의 형식이 바뀌면 `list`에 세션이 나오지 않거나 인증이 실패한다.
