# duo-session

`duo-session`은 Claude Code 세션(`main`)이 작업을 지시하고 대화형 agy 세션(`peer`)이 작업을 수행하는 Claude Code 스킬이다.

- `peer`는 커스텀 에이전트 `duo-session-peer`로 실행한다.
- `peer`에서 `main`으로는 `claude-queue send`로 메시지를 보내고, `main`에서 `peer`로는 핸드셰이크가 전달한 값을 설정해 `agy agentapi send-message`로 메시지를 보낸다.
- 기존 [duo-session-agy](duo-session-agy.md)는 `duo-session`으로 대체되어 더는 권장하지 않는(deprecated) 스킬이다.

## 역할

| 작업                                                    | main | peer |
| ------------------------------------------------------- | ---- | ---- |
| 지시                                                    | O    |      |
| 작업 범위·전제 문서와 설계 문서의 영문 초안             | O    |      |
| 리포지토리 파일 편집(한국어 문서, 코드)                 |      | O    |
| 포매터 실행, 커밋 메시지와 PR/MR 제목·본문 작성         |      | O    |
| 리뷰, diff 확인, 테스트                                 | O    |      |
| `git`과 `gh` 또는 `glab` 명령 실행(커밋, 브랜치, PR/MR) | O    |      |

- `main`은 지시문과 리뷰 지적을 영어로 작성해 `peer`에 보낸다.
- `main`은 리포지토리 파일에 한국어를 직접 쓰지 않는다.
  - 한국어 문장이 어색하면 `main`이 직접 고치지 않고 `peer`에 지적을 보낸다.

## 메시지 흐름

- 시작할 때 `main`은 공유 디렉터리 경로, `instruction-1.md` 경로, 회신받을 `main`의 이름을 첫 프롬프트로 넣어 `peer`를 시작하는 `agy --agent duo-session-peer -i "..."` 명령을 사용자에게 건넨다.
  - `peer`가 이미 실행 중이면 사용자는 같은 영문 한 줄을 `peer`에 붙여 넣는다.
- `peer`가 `main`에 보내는 첫 회신과 `peer` 재시작 후의 첫 회신에는 `--handshake`가 붙는다.
  - `main`은 핸드셰이크에 담긴 세 가지 `ANTIGRAVITY_*` 값을 하네스 스크래치패드에 기록한다.
- 라운드마다 `main`은 `instruction-<N>.md` 경로를 적은 한 줄 메시지를 `agy agentapi send-message`로 보낸다.
  - `peer`는 `reply-<N>.md` 경로를 적은 한 줄 메시지를 `claude-queue send`로 보낸다.
- `peer`의 회신이 `main` 세션에 새 메시지로 도착하므로 `main`은 백그라운드 대기 명령을 실행하지 않는다.

## 공유 디렉터리

- 경로: `main`의 하네스 스크래치패드(절대 경로)
- `main`과 `peer`가 공유 디렉터리에 추가하는 파일은 다음과 같다.
  - `instruction-<N>.md`: `main`이 작성하는 지시문 본문
  - `reply-<N>.md`: `peer`가 작성하는 회신 본문
  - `commit-message.md`: `peer`가 작성하는 커밋 메시지
  - `pr-title.md`: `peer`가 작성하는 PR/MR 제목
  - `pr-body.md`: `peer`가 작성하는 PR/MR 본문

## 설치 파일

| 파일                                       | 원본                                                                                 |
| ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `.claude/skills/duo-session/SKILL.md`      | 템플릿 `duo-session.md`                                                              |
| `.agents/agents/duo-session-peer/agent.md` | 템플릿 `duo-session-peer.md`                                                         |
| `claude-queue` 바이너리                    | `GOBIN=<dir> go install github.com/nil-park/claude-workflow-kit/claude-queue@latest` |

- 두 템플릿은 모두 `bootstrap-duo-session` 스킬 디렉터리에 있다.
- 부트스트랩 에이전트는 `bootstrap-statusline`과 같은 방식으로 사용자와 함께 `claude-queue` 설치 디렉터리를 고른다.
- 커스텀 에이전트는 프로젝트 안에만 설치하며, `~/.gemini/config/agents/`에는 설치하지 않는다.
- 선행 스킬(`docs-standards`, `coding-standards`, `anti-claudeism`, `git-workflow`, `scratch-dir`)이 Claude Code 스킬로 설치되어 있어야 한다.
- 업데이트 절차는 `SKILL.md`와 `agent.md`를 각 템플릿으로 갱신한다.
