# duo-session-agy

`duo-session-agy`는 Claude Code 세션(`main`)이 작업을 지시하고 대화형 agy 세션(`peer`)이 작업을 수행하는 Claude Code 스킬이다.

- `peer`는 커스텀 에이전트 `duo-session-agy-peer`로 실행한다.
- 두 세션은 [duo-relay](duo-relay.md)로 메시지를 주고받는다.

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

## 공유 디렉터리

- 경로: `<OS-temp>/agent/<project-slug>/duo-session-agy/`
  - `project-slug`는 [scratch-dir 템플릿](../../plugins/project-skills-bootstrap/skills/bootstrap-scratch-dir/scratch-dir.md#경로-결정)의 규칙을 따른다.
  - 경로에는 타임스탬프가 없으며, 프로젝트에서는 실행할 때마다 같은 경로를 쓴다.
- `main`은 세션을 시작할 때 디렉터리를 비우며, 디렉터리가 없으면 새로 만든다.
- duo-relay가 만드는 파일 외에 `main`과 `peer`가 공유 디렉터리에 추가하는 파일은 다음과 같다.
  - `instruction-<N>.md`: `main`이 `push --to peer`에 넘기는 지시문 본문
  - `reply-<N>.md`: `peer`가 `push --to main`에 넘기는 회신 본문
  - `commit-message.md`: `peer`가 작성하는 커밋 메시지
  - `pr-title.md`: `peer`가 작성하는 PR/MR 제목
  - `pr-body.md`: `peer`가 작성하는 PR/MR 본문
- duo-relay가 관리하는 파일(`endpoint-*.json`, `round-*.md`, `wait-*.pid`)은 [duo-relay 설계 문서](duo-relay.md#공유-디렉터리)에 정의되어 있다.

## 설치 파일

| 파일                                           | 원본                                                                              |
| ---------------------------------------------- | --------------------------------------------------------------------------------- |
| `.claude/skills/duo-session-agy/SKILL.md`      | 템플릿 `duo-session-agy.md`                                                       |
| `.agents/agents/duo-session-agy-peer/agent.md` | 템플릿 `duo-session-agy-peer.md`                                                  |
| `duo-relay` 바이너리                           | `GOBIN=<dir> go install github.com/nil-park/claude-workflow-kit/duo-relay@latest` |
| `.claude/settings.json`의 `permissions.allow`  | 항목 `Bash(duo-relay:*)`                                                          |

- 두 템플릿은 모두 `bootstrap-duo-session` 스킬 디렉터리에 있다.
- 부트스트랩 에이전트는 `bootstrap-statusline`과 같은 방식으로 사용자와 함께 `duo-relay` 설치 디렉터리를 고른다.
- 커스텀 에이전트는 프로젝트 안에만 설치하며, `~/.gemini/config/agents/`에는 설치하지 않는다.
- 선행 스킬(`docs-standards`, `coding-standards`, `anti-claudeism`, `git-workflow`, `scratch-dir`)이 Claude Code 스킬로 설치되어 있어야 한다.
- 업데이트 절차는 `SKILL.md`와 `agent.md`를 각 템플릿으로 갱신한다.
