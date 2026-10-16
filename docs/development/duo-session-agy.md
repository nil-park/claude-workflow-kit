# duo-session-agy 개발

## 전제

- 리포지토리의 `.agents/rules/`에 상시 로드(always-on) 규칙을 두지 않는다.
  - 상시 로드 규칙이 필요한 `multi-session-agy`와는 같은 리포지토리에서 함께 쓰지 않는다.
  - 프로젝트 루트의 `AGENTS.md`는 `peer`에도 함께 로드될 수 있다.
  - `.agents/rules/`에 규칙이 있으면 부트스트랩이 사용자에게 경고한다.
- agy는 승인 요청 없이 도구를 실행한다.
  - 사용자가 `~/.gemini/antigravity-cli/settings.json`에서 `toolPermission: always-proceed`와 `allowNonWorkspaceAccess: true`를 설정한다.
  - 부트스트랩은 이 설정값을 확인만 하고, 변경은 사용자가 직접 한다.
- duo-relay를 실행하기 위한 Claude Code 허용 규칙은 duo-relay의 전제이며 [duo-relay 개발 문서](duo-relay.md#전제)에 정리되어 있다.

## 커스텀 에이전트 동작 (agy 1.4.0, 2026-10-17)

- 상시 로드 규칙, `AGENTS.md`, `~/.gemini/GEMINI.md`는 커스텀 에이전트에서도 모두 로드된다.
  - 상시 로드 규칙을 제외하는 프런트매터 필드는 없다.
- 셸 도구와 파일 편집 도구는 `tools`에 나열했을 때만 사용할 수 있다.
- `agent.md` 본문에 적은 `@[label](path)`는 펼쳐지지 않고 문자열 그대로 남는다.
- agy TUI에서 `!`로 실행하는 셸 명령은 인자를 모두 버린다.
  - `peer`는 `!`가 아니라 모델의 셸 도구로 `register`를 실행한다.

## 제약 사항

- 이 리포지토리에서는 주로 `core.symlinks=false`인 Windows 환경에서 작업한다.
  - 여기에 설치한 스킬은 템플릿을 가리키는 심볼릭 링크가 될 수 없다.
  - `.claude/skills/<name>/SKILL.md`는 한 줄짜리 `@` 참조로 템플릿을 가리킨다.
  - 스킬 본문과 `agent.md` 안의 `@` 참조는 펼쳐지지 않는다.
- 이 리포지토리의 `agent.md`는 템플릿과 한 곳이 다르다.
  - `peer`가 세션에서 문서나 코드를 처음 편집할 때 `plugins/` 아래의 `docs-standards`와 `coding-standards` 템플릿을 직접 읽도록 지시한다.
  - 심볼릭 링크가 동작하는 환경에서는 이 차이가 필요하지 않다.

## 검증 방법

- 스모크 테스트
  1. `agy --agent duo-session-agy-peer`를 시작한다.
  2. 공유 디렉터리를 알려 주고 `endpoint-peer.json`이 생기는지 확인한다.
  3. `peer`에 라운드를 `push`하고, `peer`가 깨어나는지와 `main`의 백그라운드 `wait`가 회신을 출력하는지 확인한다.

## 위험 요소

- agy가 컨텍스트를 압축할 때 `agent.md` 본문이 유지되는지는 측정하지 않았다.
  - 컨텍스트 압축 뒤에 긴 세션에서 `peer`가 자기 역할이나 앞서 읽은 기준 문서의 내용을 잃을 수 있다.
