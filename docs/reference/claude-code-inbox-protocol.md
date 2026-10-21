# Claude Code 세션 inbox 프로토콜

Claude Code 세션이 다른 프로세스에서 메시지를 받는 inbox socket의 규약이다. 공식 문서에는 인증 줄까지만 나오고, 나머지는 실측과 바이너리 분석으로 확인했다.

- 관측 조건: Claude Code 2.1.306, Windows, 2026-10-21
- Unix에 관한 내용은 바이너리에서 읽은 것이며 실측하지 않았다.
- 공개된 부분은 [cross-session messaging](https://code.claude.com/docs/en/cross-session-messaging#the-sessions-inbox-socket) 문서에 있다.

## 세션 등록 파일

- 경로: `~/.claude/sessions/<pid>.json`
  - `CLAUDE_CONFIG_DIR`을 설정한 환경에서 경로가 바뀌는지는 확인하지 않았다.
- 세션이 시작될 때 생기고, 세션이 끝나면 지워진다.

| 필드                  | 실측 예시                           | 설명                                                                |
| --------------------- | ----------------------------------- | ------------------------------------------------------------------- |
| `pid`                 | `18580`                             | 세션 프로세스의 pid                                                 |
| `name`                | `"claude-workflow-kit-7a"`          | `ListAgents`가 보여 주는 이름. 비어 있으면 `cwd`의 마지막 경로 요소 |
| `nameSource`          | `"derived"`, `"peer"`               | 자동으로 지은 이름이면 `derived`, `--name`으로 지정하면 `peer`      |
| `kind`                | `"interactive"`, `"bg"`             | `interactive`, `bg`, `daemon`, `daemon-worker` 중 하나              |
| `status`              | `"idle"`                            | `busy`, `shell`, `idle`, `waiting` 중 하나                          |
| `cwd`                 | (가림)                              | 세션의 작업 디렉터리                                                |
| `messagingSocketPath` | `"\\.\pipe\LOCAL\cc-msg-<32 hex>"`  | inbox socket 경로                                                   |
| `procStart`           | `"134370625679042977"`              | 세션 프로세스의 생성 시각                                           |
| `pidDomain`           | `"win32:<호스트 이름 소문자>"`      | pid가 속한 호스트와 namespace                                       |
| `peerProtocol`        | `1`                                 | 프로토콜 버전                                                       |
| `peerFeatures`        | `["notify_idle", "artifact_yield"]` | 지원 기능                                                           |
| `sessionId`           | (가림)                              | 세션 UUID                                                           |
| `startedAt`           | `1792588968878`                     | 시작 시각(Unix epoch 밀리초)                                        |

- `procStart`의 형식은 운영체제마다 다르다.
  - Windows: 프로세스 생성 시각의 FILETIME을 10진 문자열로 적은 값
  - Unix: `LC_ALL=C TZ=UTC ps -o lstart= -p <pid>` 출력을 trim한 문자열
- `ListAgents`는 등록 파일의 pid가 살아 있고 생성 시각이 `procStart`와 맞는 세션만 보여 준다.
  - 목록을 만들 때 각 inbox socket에 접속했다가 아무것도 보내지 않고 연결을 끊는다.

## 인증 키

- 세션은 시작할 때 16바이트 무작위 값 두 개를 hex 문자열로 만든다.
  - `childToken`: 세션의 자식 프로세스에 `CLAUDE_CODE_MESSAGING_TOKEN`으로 넘긴다.
  - `peerToken`: 키 파일에 저장한다.
- 키 파일 경로: `~/.claude/sessions/<pid>.<hash>.key`
  - 내용: `{"peerToken":"<32 hex>","procStartFt":"…","pidDomain":"…"}`
    - Unix에서는 `procStartFt` 대신 `procStart`가 들어간다.
  - `<hash>`는 정규화한 소켓 경로를 UTF-8로 인코딩한 값의 SHA-256 hex다.
    - Windows: `\\.\pipe\` 뒤의 이름에서 ASCII 대문자만 소문자로 바꾼다(예: `\\.\pipe\local\cc-msg-<hex>`).
    - Unix: 절대 경로로 바꾼 소켓 경로를 그대로 쓴다.
  - 송신자는 `peerToken`이 `^[0-9a-f]{32}$`에 맞지 않는 키 파일을 쓰지 않는다.
- Unix에서도 키 파일이 생기는지는 확인하지 못했다.

## 메시지 전송

- 한 연결에서 UTF-8 JSON을 줄 단위로 보내고, 줄마다 `\n`으로 끝낸다.
- 받는 쪽은 30초 안에 완성된 줄이 오지 않으면 연결을 끊는다.
- 받는 쪽은 응답을 보내지 않는다.

### 인증 줄

```json
{ "type": "auth", "token": "<token>" }
```

- 연결의 첫 줄로 보낸다.
- Windows에서는 반드시 보내야 한다.
  - 이 줄이 없거나 토큰이 틀리면 받는 쪽은 그 연결로 온 메시지를 전달하지 않는다.
- Unix에서는 생략할 수 있다.
- Claude 송신자는 받는 세션의 키 파일이 있으면 `peerToken`을 보내고, 없으면 이 줄을 생략한다.

### 메시지 줄

```json
{
  "msgV": 1,
  "msg_id": "<UUID>",
  "type": "user",
  "message": {
    "role": "user",
    "content": "<cross-session-message from=\"<from>\" from-name=\"<name>\" from-mode=\"prompting\">\n<본문>\n</cross-session-message>"
  },
  "priority": "next",
  "from": "<from>"
}
```

- `content`의 `<cross-session-message>` 요소는 송신자가 만든다.
- Claude 송신자는 `from`에 `uds:<자기 소켓 경로>`를 넣는다.
  - 받는 쪽은 형식에 맞지 않는 `from`도 받는다.
  - 받는 Claude는 `from`이 가리키는 소켓으로 회신한다.
- `from-mode`는 송신자가 밝히는 permission 모드 분류다.
  - prompting 모드의 Claude 송신자는 `prompting`을 넣는다.

## 받는 쪽 처리

- 인증 줄의 토큰이 `peerToken`과 같으면 `peer`, `childToken`과 같으면 `child`로 분류한다.
- `child` 메시지는 `crossSessionInbound`가 정해지지 않았을 때 바로 전달된다.
- `peer` 메시지에는 [inbound controls](https://code.claude.com/docs/en/cross-session-messaging#control-inbound-messages)가 적용된다.
  - 받는 세션이 prompting 모드이고 `from-mode`가 `prompting`이면 바로 전달된다.
  - 받는 세션이 bypass 모드이면 보류하고 승인 창을 띄운다.
- 전달된 메시지는 받는 세션의 대화에 `<cross-session-message>` 그대로 들어간다.
  - 입력을 기다리던 세션은 키 입력 없이 새 턴을 시작한다.
  - 턴이 진행 중이면 다음 도구 라운드에 들어간다.
