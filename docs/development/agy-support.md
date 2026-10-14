# agy 지원의 전제

## 전제

- 이 리포지토리를 개발할 때 쓰는 도구는 Claude Code와 agy다.
- Claude Code는 `autoCompactEnabled: false`로 자동 압축을 끈다.
- agy에서는 사용자가 자동 압축을 끄거나 압축 기준을 바꿀 수 없다.
  - 공식 CLI 문서에는 자동 압축과 압축 기준을 설정하는 항목이 없다.
  - 압축 기준은 SDK의 `CompactionConfig(token_threshold=...)`로만 지정할 수 있다.
    - 값을 생략하면 백엔드 기본값이 적용된다.
  - 공식 문서에는 CLI의 기본값이 적혀 있지 않다.
  - 공개 이슈에는 서로 다른 값이 보고되어 있다.
    - [#878](https://github.com/google-antigravity/antigravity-cli/issues/878): Gemini 3.8 Flash에서 압축 직전 컨텍스트의 중앙값이 약 255,000토큰이다.
    - [#634](https://github.com/google-antigravity/antigravity-cli/issues/634): 압축 기준이 약 90,000에서 120,000토큰 사이라고 보고한다.
  - 압축을 거치면 스킬 본문이 요약에서 빠진다.
    - 그래서 스킬 본문을 읽어 오는 방식으로는 지침이 유지되지 않는다.
- 따라서 압축에 대응하는 구성은 agy에만 필요하다.
- 프로젝트 지침의 SoT는 `AGENTS.md`다.
  - `CLAUDE.md`는 `@AGENTS.md`로 그 내용을 가져온다.
  - agy는 `AGENTS.md`를 읽는다.
  - Claude Code는 `CLAUDE.md`를 읽는다.
- agy에서는 스킬 대신 스킬의 원본 템플릿을 상시 규칙으로 불러온다.
  - 규칙 파일은 `.agents/rules/<스킬 이름>.md` 경로에 만든다.
  - 대상은 아래 여섯 스킬이다.
    - `docs-standards`, `coding-standards`, `work-cycle`
    - `git-workflow`, `scratch-dir`, `anti-claudeism`
- 규칙 파일은 원본 템플릿을 `@[라벨](경로)`로 불러오고, 본문을 복사하지 않는다.
  - 원본 템플릿이 SoT다.

## 확인한 사실

- 아래 내용은 agy 1.3.5와 `gemini-3.8-flash-medium`에서 확인했다.
- 상시 규칙
  - `.agents/rules/*.md`는 `trigger: always_on` 프런트매터가 있어야 매 턴 컨텍스트에 로드된다.
  - 프런트매터가 없거나 `trigger` 값이 틀린 규칙은 알림 없이 무시된다.
- 인클루드
  - `@[라벨](경로)`를 쓰면 도구 호출 없이 대상 파일의 내용이 본문에 삽입된다.
  - agy는 상대 경로를 규칙 파일이 있는 디렉터리 기준으로 해석하며, 리포지토리 루트는 기준으로 쓰지 않는다.
  - 규칙의 맨 `@경로`는 에이전트가 읽을 수 있는 절대 경로 참조로 바뀐다.
    - 이때 대상 파일의 내용은 삽입되지 않는다.
  - 스킬 본문에서는 `@경로`가 있어도 대상 파일의 내용이 삽입되지 않는다.
- 컨텍스트 압축
  - 규칙은 압축 뒤에도 유지된다.
    - 압축 뒤에 인클루드 원본을 바꾸자 새 내용이 로드되었다.
  - 스킬 본문은 압축 뒤에 남지 않는다.
    - 요약에는 호출할 때 로드된 본문이 빠지고, 경로와 풀어쓴 문장만 들어간다.
- 상한
  - 규칙 파일 하나는 인클루드 내용이 삽입된 뒤 24,000바이트까지만 로드된다.
    - 상한을 넘는 내용은 줄 단위로 잘린다.
  - 상시 규칙은 합계 20,000토큰의 예산을 나눠 쓴다.
    - 예산을 넘으면 큰 파일부터 전문 대신 경로 포인터가 로드된다.
      - 포인터에는 규칙의 `description`이 설명으로 붙는다.
    - 이 예산은 [공식 문서](../setup/agy-docs.md)에 적힌 값이며, 직접 측정한 값은 아니다.
    - 여섯 규칙의 입력 토큰 합계는 약 11,400이다([측정 문서](../experiments/agy-rule-token-size.md)).
- 탐색 위치
  - agy는 프로젝트의 `.agents/`와 `.agent/`에서 스킬과 규칙을 모두 읽는다.
    - 공식 문서는 `.agent/rules/`를 기존 형식(legacy)으로 표시한다.
  - agy는 `.gemini/skills`와 `.gemini/rules`를 읽지 않는다.
  - 루트의 `AGENTS.md`와 `GEMINI.md`는 둘 다 로드된다.
    - 둘 사이의 우선순위는 없다.
